package main

import (
	// "context"
	// "time"
	"os"
	"os/signal"
	"sort"
	"strings"
	// "syscall"
	// "errors"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
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
			Name:  "kafka-brokers, kb",
			Value: "kfk1:19092,kfk2:29092,kfk3:39092",
			Usage: "Kafka brokers in comma separated value",
		},
		cli.BoolFlag{
			Name:  "kafka-verbose, kv",
			Usage: "Kafka verbose logging",
		},
		cli.StringFlag{
			Name:  "kafka-consumer-group, kcg",
			Value: "consumer-group",
			Usage: "Kafka consumer group",
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
	log.Info().Msg("kafka tutorial consume.")
	log.Info().Msg("(c) Erik 2019")

	brokerUrls := c.String("kafka-brokers")
	verbose := c.Bool("kafka-verbose")
	clientID := c.String("kafka-client-id")
	topic := c.String("kafka-topic")
	consumerGroup := c.String("kafka-consumer-group")

	log.Info().Msgf("kafka-brokers: %s", brokerUrls)
	log.Info().Msgf("kafka-verbose: %t", verbose)
	log.Info().Msgf("kafka-client-id: %s", clientID)
	log.Info().Msgf("kafka-topic: %s", topic)
	log.Info().Msgf("kafka-consumer-group: %s", consumerGroup)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go clusterConsumer(wg, strings.Split(brokerUrls, ","), []string{"hello"}, "consume-group")
	wg.Wait();

	return nil
}


// 支持brokers cluster的消费者
func clusterConsumer(wg *sync.WaitGroup, brokers, topics []string, groupID string)  {
    defer wg.Done()
    config := cluster.NewConfig()
    config.Consumer.Return.Errors = true
    config.Group.Return.Notifications = true
    config.Consumer.Offsets.Initial = sarama.OffsetNewest
 
    // init consumer
    consumer, err := cluster.NewConsumer(brokers, groupID, topics, config)
    if err != nil {
        log.Debug().Msgf("%s: sarama.NewSyncProducer err, message=%s \n", groupID, err)
        return
    }
    defer consumer.Close()
 
    // trap SIGINT to trigger a shutdown
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Interrupt)
 
    // consume errors
    go func() {
        for err := range consumer.Errors() {
            log.Debug().Msgf("%s:Error: %s\n", groupID, err.Error())
        }
    }()
 
    // consume notifications
    go func() {
        for ntf := range consumer.Notifications() {
            log.Debug().Msgf("%s:Rebalanced: %+v \n", groupID, ntf)
        }
    }()
 
    // consume messages, watch signals
    var successes int
    Loop:
    for {
        select {
        case msg, ok := <-consumer.Messages():
            if ok {
                log.Debug().Msgf("%s:%s/%d/%d\t%s\t%s\n", groupID, msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
                consumer.MarkOffset(msg, "")  // mark message as processed
                successes++
            }
        case <-signals:
            break Loop
        }
    }
    log.Debug().Msgf("%s consume %d messages \n", groupID, successes)
}
