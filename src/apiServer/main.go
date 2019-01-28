package main

import (
	// "fmt"
	"os"
	"sort"

	// "github.com/namsral/flag"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// var (
// 	listenAddrApi string
// )

func main() {
	app := cli.NewApp()
	app.Name = "API Server Commandline"
	app.Usage = "Run API Server"
	app.Version = "1.0.0"
	app.Flags = args()
	app.Commands = commands()
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

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
			Name:  "config, c",
			Value: "balbalbal",
			Usage: "test config ...",
		},
	}
}

func commands() []cli.Command {
	return []cli.Command{
		{
			Name:    "complete",
			Aliases: []string{"c"},
			Usage:   "complete a task on the list",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}
}

func action(c *cli.Context) error {
	log.Info().Msg("Commandline Sample.")
	log.Info().Msg("(c) erik 2019")
	log.Info().Msgf("listen-address : %s \n", c.String("listen-address"))
	log.Info().Msgf("config : %s \n", c.String("config"))
	return nil
}
