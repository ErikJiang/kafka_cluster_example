package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/ErikJiang/kafka_tutorial/src/apiServer/kafka"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "kafka Tutorial API Server Commandline"
	app.Usage = "Run API Server"
	app.Version = "1.0.0"
	app.Flags = args()
	sort.Sort(cli.FlagsByName(app.Flags))
	app.Action = action
	err := app.Run(os.Args)
	if err != nil {
		log.Error().Msgf("error: %v", err)
	}
	log.Debug().Msgf("args: %v", os.Args)
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
			Value: "localhost:19092,localhost:29092,localhost:39092",
			Usage: "Kafka brokers in comma separated value",
		},
		cli.BoolFlag{
			Name:  "kafka-verbose, kv",
			Usage: "Kafka verbose logging",
		},
		cli.StringFlag{
			Name:  "kafka-cli-id, kci",
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
	log.Info().Msg("kafka tutorial api server.")
	log.Info().Msg("(c) Erik 2019")

	listenAddr := c.String("listen-address")
	brokerUrls := c.String("kafka-brokers")
	verbose := c.Bool("kafka-verbose")
	clientID := c.String("kafka-cli-id")
	topic := c.String("kafka-topic")

	log.Info().Msgf("listen-address: %s", listenAddr)
	log.Info().Msgf("kafka-brokers: %s", brokerUrls)
	log.Info().Msgf("kafka-verbose: %t", verbose)
	log.Info().Msgf("kafka-cli-id: %s", clientID)
	log.Info().Msgf("kafka-topic: %s", topic)

	producer, err := kafka.Configure(strings.Split(brokerUrls, ","), clientID, topic)
	if err != nil {
		log.Error().Msgf("config kafka producer fail, err: %v", err)
		return err
	}
	defer producer.Close()

	errChan := make(chan error, 1)
	go func() {
		log.Info().Msgf("starting server at %s", listenAddr)
		errChan <- server(listenAddr)
	}()

	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <- signalChan:
		log.Info().Msg("got an interrupt, exiting...")
		return errors.New("Got an interrupt exiting")
	case err := <- errChan:
		if err != nil {
			log.Error().Err(err).Msg("error while runing api, exiting...")
			return err
		}
	}
	return nil
}

func server(listenAddr string) error {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.POST("/api/v1/data", postDataToKafka)
	for _, routeInfo := range router.Routes() {
		log.Debug().
			Str("path", routeInfo.Path).
			Str("handler", routeInfo.Handler).
			Str("method", routeInfo.Method).
			Msg("registered routes")
	}
	return router.Run(listenAddr)
}

func postDataToKafka(ctx *gin.Context) {
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

	err = kafka.PushMsg(parent, nil, formInBytes)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("error while push message into kafka: %s", err.Error()),
			},
		})
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success push data into kafka",
		"data": form,
	})
}
