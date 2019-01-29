package main

import (
	"os"
	"sort"

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

func action(c *cli.Context) error {
	log.Info().Msg("kafka tutorial api server.")
	log.Info().Msg("(c) Erik 2019")
	log.Info().Msgf("listen-address : %s \n", c.String("listen-address"))
	log.Info().Msgf("kafka-brokers : %s \n", c.String("kafka-brokers"))
	log.Info().Msgf("kafka-verbose : %t \n", c.Bool("kafka-verbose"))
	log.Info().Msgf("kafka-cli-id : %s \n", c.String("kafka-cli-id"))
	log.Info().Msgf("kafka-topic : %s \n", c.String("kafka-topic"))
	return nil
}
