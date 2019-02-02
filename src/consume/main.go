package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "kafka Tutorial Consume Commandline"
	app.Usage = "Run Consume"
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
			Name:   "kafka-brokers, kb",
			Value:  "kfk1:19092,kfk2:29092,kfk3:39092",
			Usage:  "Kafka brokers in comma separated value",
			EnvVar: "KAFKA_BROKERS",
		},
		cli.StringFlag{
			Name:   "kafka-consumer-group, kcg",
			Value:  "consumer-group",
			Usage:  "Kafka consumer group",
			EnvVar: "KAFKA_CONSUMER_GROUP_ID",
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
	log.Info().Msg("kafka tutorial consume.")
	log.Info().Msg("(c) Erik 2019")

	brokerUrls := c.String("kafka-brokers")
	topic := c.String("kafka-topic")
	consumerGroup := c.String("kafka-consumer-group")

	log.Info().Msgf("kafka-brokers: %s", brokerUrls)
	log.Info().Msgf("kafka-topic: %s", topic)
	log.Info().Msgf("kafka-consumer-group: %s", consumerGroup)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go clusterConsumer(wg, strings.Split(brokerUrls, ","), []string{topic}, consumerGroup)
	wg.Wait()

	return nil
}

// 支持brokers cluster的消费者
func clusterConsumer(wg *sync.WaitGroup, brokers, topics []string, groupID string) {
	defer wg.Done()
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 初始化消费者
	consumer, err := cluster.NewConsumer(brokers, groupID, topics, config)
	if err != nil {
		log.Debug().Msgf("%s: sarama.NewSyncProducer err, message=%s \n", groupID, err)
		return
	}
	defer consumer.Close()

	// 捕获终止中断信号触发程序退出
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// 消费错误信息
	go func() {
		for err := range consumer.Errors() {
			log.Debug().Msgf("%s:Error: %s\n", groupID, err.Error())
		}
	}()

	// 消费通知信息
	go func() {
		for ntf := range consumer.Notifications() {
			log.Debug().Msgf("%s:Rebalanced: %v \n", groupID, ntf)
		}
	}()

	// 消费信息及监听信号
	var successes int
Loop:
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				value := struct {
					Text string `form:"text" json:"text"`
				}{}
				err := json.Unmarshal(msg.Value, &value)
				if err != nil {
					log.Error().Msgf("consume message json format error, %v", err)
					break Loop
				}
				log.Debug().Msgf("GroupID: %s, Topic: %s, Partition: %d, Offset: %d, Key: %s, Value: %s",
					groupID, msg.Topic, msg.Partition, msg.Offset, msg.Key, value.Text)
				consumer.MarkOffset(msg, "") // 标记信息为已处理
				successes++
			}
		case <-signals:
			break Loop
		}
	}
	log.Debug().Msgf("%s consume %d messages", groupID, successes)
}
