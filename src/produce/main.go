package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "kafka Tutorial Produce Commandline"
	app.Usage = "Run Produce"
	app.Version = "1.0.0"
	app.Flags = args()
	sort.Sort(cli.FlagsByName(app.Flags))
	app.Action = action
	app.Run(os.Args)
}

// args 命令行参数定义
func args() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "listen-address, la",
			Value:  "0.0.0.0:9000",
			Usage:  "Listen address for api",
			EnvVar: "LISTEN_ADDRESS",
		},
		cli.StringFlag{
			Name:   "kafka-brokers, kb",
			Value:  "kfk1:19092,kfk2:29092,kfk3:39092",
			Usage:  "Kafka brokers in comma separated value",
			EnvVar: "KAFKA_BROKERS",
		},
		cli.StringFlag{
			Name:   "kafka-topic, kt",
			Value:  "hello",
			Usage:  "Kafka topic to push",
			EnvVar: "KAFKA_TOPIC",
		},
	}
}

// action 创建 Kafka 生产者并启动路由服务
func action(c *cli.Context) error {
	log.Info().Msg("kafka tutorial produce.")
	log.Info().Msg("(c) Erik 2019")

	listenAddr := c.String("listen-address")
	kafkaBrokers := c.String("kafka-brokers")
	topic := c.String("kafka-topic")

	log.Info().Msgf("listen-address: %s", listenAddr)
	log.Info().Msgf("kafka-brokers: %s", kafkaBrokers)
	log.Info().Msgf("kafka-topic: %s", topic)

	brokerUrls := strings.Split(kafkaBrokers, ",")

	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForAll

	config.Producer.Partitioner = sarama.NewRandomPartitioner

	config.Producer.Return.Successes = true

	config.Producer.Return.Errors = true

	config.Version = sarama.V2_1_0_0

	log.Info().Msg("start make topic")
	err := createTopic(config, brokerUrls[0], topic)
	if err != nil {
		log.Error().Msgf("%v", err)
		return err
	}

	log.Info().Msg("start make producer")
	producer, err := sarama.NewAsyncProducer(brokerUrls, config)
	if err != nil {
		log.Error().Msgf("%v", err)
		return err
	}
	defer producer.AsyncClose()

	log.Info().Msgf("starting server at %s", listenAddr)
	errChan := make(chan error, 1)
	go func(p sarama.AsyncProducer) {
		errChan <- httpServer(p, topic, listenAddr)
	}(producer)

	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

Loop:
	for {
		select {
		case succ := <-producer.Successes():
			log.Info().Msgf("success: offset: %d, timestamp: %s, partitions: %d",
				succ.Offset, succ.Timestamp.String(), succ.Partition)
		case fail := <-producer.Errors():
			log.Error().Err(fail).Msg("fail while produce message, exiting...")
			break Loop
		case <-signalChan:
			log.Info().Msg("got an interrupt, exiting...")
			break Loop
		case err := <-errChan:
			log.Error().Err(err).Msg("error while runing api, exiting...")
			break Loop
		}
	}
	return nil
}

// createTopic 创建 Topic
func createTopic(config *sarama.Config, brokerURL, topicName string) error {
	broker := sarama.NewBroker(brokerURL)
	broker.Open(config)
	yes, err := broker.Connected()
	if err != nil {
		log.Error().Msgf("broker connect fail, %v", err)
		return err
	}
	log.Debug().Msgf("broker connect status: %v", yes)

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: 1,
		ConfigEntries:     make(map[string]*string),
	}

	topicDetails := make(map[string]*sarama.TopicDetail)
	topicDetails[topicName] = topicDetail

	request := sarama.CreateTopicsRequest{
		Timeout:      time.Second * 15,
		TopicDetails: topicDetails,
	}

	response, err := broker.CreateTopics(&request)
	if err != nil {
		log.Error().Msgf("create topics fail, %v", err)
		return err
	}
	log.Debug().Msgf("response length: %d", len(response.TopicErrors))
	for key, val := range response.TopicErrors {
		log.Debug().Msgf("Key is %s", key)
		log.Debug().Msgf("Val is %#v", val.Err.Error())
		log.Debug().Msgf("ValMsg is %#v", val.ErrMsg)
	}
	log.Info().Msgf("create topics response: %v", response)
	broker.Close()
	return nil
}

// httpServer 启动 HTTP 服务
func httpServer(producer sarama.AsyncProducer, topicName, listenAddr string) error {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.POST("/api/v1/data", func(ctx *gin.Context) {
		parent := context.Background()
		defer parent.Done()

		form := &struct {
			Text string `form:"text" json:"text"`
		}{}

		err := ctx.ShouldBindJSON(form)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": map[string]interface{}{
					"message": fmt.Sprintf("error while bind request param: %s", err.Error()),
				},
			})
			ctx.Abort()
			return
		}
		formInBytes, err := json.Marshal(form)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
				"error": map[string]interface{}{
					"message": fmt.Sprintf("error while marshalling json: %s", err.Error()),
				},
			})
			ctx.Abort()
			return
		}

		// send message to kafka
		msg := &sarama.ProducerMessage{
			Topic:     topicName,
			Key:       sarama.StringEncoder(MakeSha1(form.Text)),
			Value:     sarama.ByteEncoder(formInBytes),
			Timestamp: time.Now(),
		}
		producer.Input() <- msg

		ctx.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "success push data into kafka",
			"data":    form,
		})
	})
	return router.Run(listenAddr)
}

// MakeSha1 计算字符串的 sha1 hash 值
func MakeSha1(source string) string {
	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(source))
	return hex.EncodeToString(sha1Hash.Sum(nil))
}
