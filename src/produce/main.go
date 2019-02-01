package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

var (
	ListenAddr string
	BrokerUrls string
	ClientID   string
	Topic      string
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
			Name:  "listen-address, la",
			Value: "0.0.0.0:9000",
			Usage: "Listen address for api",
		},
		cli.StringFlag{
			Name:  "kafka-brokers, kb",
			Value: "kfk1:19092,kfk2:29092,kfk3:39092",
			Usage: "Kafka brokers in comma separated value",
		},
		cli.StringFlag{
			Name:  "kafka-client-id, kci",
			Value: "kafka-client",
			Usage: "Kafka client id to connect",
		},
		cli.StringFlag{
			Name:  "kafka-topic, kt",
			Value: "hello",
			Usage: "Kafka topic to push",
		},
	}
}

// action 创建 Kafka 生产者并启动路由服务
func action(c *cli.Context) error {
	log.Info().Msg("kafka tutorial produce.")
	log.Info().Msg("(c) Erik 2019")

	ListenAddr = c.String("listen-address")
	BrokerUrls = c.String("kafka-brokers")
	ClientID = c.String("kafka-client-id")
	Topic = c.String("kafka-topic")

	log.Info().Msgf("listen-address: %s", ListenAddr)
	log.Info().Msgf("kafka-brokers: %s", BrokerUrls)
	log.Info().Msgf("kafka-client-id: %s", ClientID)
	log.Info().Msgf("kafka-topic: %s", Topic)

	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForAll

	config.Producer.Partitioner = sarama.NewRandomPartitioner

	config.Producer.Return.Successes = true

	config.Producer.Return.Errors = true

	config.Version = sarama.V2_1_0_0

	log.Info().Msg("start make producer")

	producer, err := sarama.NewAsyncProducer(strings.Split(BrokerUrls, ","), config)
	if err != nil {
		log.Error().Msgf("%v", err)
		return err
	}
	defer producer.AsyncClose()

	log.Info().Msgf("starting server at %s", ListenAddr)
	errChan := make(chan error, 1)
	go func(p sarama.AsyncProducer) {
		errChan <- httpServer(p)
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

func httpServer(producer sarama.AsyncProducer) error {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.POST("/api/v1/data", func(ctx *gin.Context) {
		parent := context.Background()
		defer parent.Done()

		form := &struct {
			Text string `form:"text" json:"text"`
		}{}

		ctx.Bind(form)
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
			Topic: Topic,
		}
		msg.Value = sarama.ByteEncoder(formInBytes)
		producer.Input() <- msg

		ctx.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "success push data into kafka",
			"data":    form,
		})
	})

	return router.Run(ListenAddr)
}
