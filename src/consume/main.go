package main

import (
	"context"
	"time"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
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

	// sigchan := make(chan os.Signal, 1)
	// signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// config := kafka.ReaderConfig{
	// 	Brokers: strings.Split(brokerUrls, ","),
	// 	GroupID: clientID,
	// 	Topic: topic,
	// 	MinBytes: 10e3,	// 10KB
	// 	MaxBytes: 10e6,	// 10MB
	// 	MaxWait: 1 * time.Second,
	// 	ReadLagInterval: -1,
	// }

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   strings.Split(brokerUrls, ","),
		// Brokers:   []string{"kfk1:19092"},
		// GroupID: "group-id",
		Topic:     topic,
		MinBytes:  1, 	// 1B
		MaxBytes:  10e6, // 10MB
		MaxWait:  1 * time.Second,
		QueueCapacity:    1024,
		SessionTimeout:   10 * time.Second,
		RebalanceTimeout: 5 * time.Second,
		
	})
	defer r.Close()
	r.SetOffset(30)
	errChan := make(chan error, 1)
	go func() {
		errChan <- loop(r)
	}()	

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	select {
		case <- sigchan:
			log.Info().Msgf("sign chan exit")
			return errors.New("sign chan exit")
		case err := <- errChan:
			if err != nil {
				log.Error().Err(err).Msg("error while runing api, exiting...")
				return err
			}
	}

	return nil

}

func loop(r *kafka.Reader) error {
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Error().Msgf("ReadMessage err: %v", err)
			return err
		}
		log.Debug().Msgf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}
}


// 支持brokers cluster的消费者
func clusterConsumer(wg *sync.WaitGroup, brokers, topics []string, groupId string)  {
    defer wg.Done()
    config := cluster.NewConfig()
    config.Consumer.Return.Errors = true
    config.Group.Return.Notifications = true
    config.Consumer.Offsets.Initial = sarama.OffsetNewest
 
    // init consumer
    consumer, err := cluster.NewConsumer(brokers, groupId, topics, config)
    if err != nil {
        log.Printf("%s: sarama.NewSyncProducer err, message=%s \n", groupId, err)
        return
    }
    defer consumer.Close()
 
    // trap SIGINT to trigger a shutdown
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Interrupt)
 
    // consume errors
    go func() {
        for err := range consumer.Errors() {
            log.Printf("%s:Error: %s\n", groupId, err.Error())
        }
    }()
 
    // consume notifications
    go func() {
        for ntf := range consumer.Notifications() {
            log.Printf("%s:Rebalanced: %+v \n", groupId, ntf)
        }
    }()
 
    // consume messages, watch signals
    var successes int
    Loop:
    for {
        select {
        case msg, ok := <-consumer.Messages():
            if ok {
                fmt.Fprintf(os.Stdout, "%s:%s/%d/%d\t%s\t%s\n", groupId, msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
                consumer.MarkOffset(msg, "")  // mark message as processed
                successes++
            }
        case <-signals:
            break Loop
        }
    }
    fmt.Fprintf(os.Stdout, "%s consume %d messages \n", groupId, successes)
}
