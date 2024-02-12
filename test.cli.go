package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/pcp/internal/log"
	"github.com/dennis-tra/pcp/pkg/receive"
	"github.com/dennis-tra/pcp/pkg/send"
)

var (
	RawVersion  = "dev"
	ShortCommit = "5f3759df"
)

func main() {
	// ShortCommit version tag
	verTag := fmt.Sprintf("v%s+%s", RawVersion, ShortCommit)

	app := &cli.App{
		Name: "LibP2P File Transfer",
		Authors: []*cli.Author{
			{
				Name:  "Dennis Trautwein",
				Email: "xyz@iitj.ac.in",
			},
		},
		Usage:                "Peer Copy, a peer-to-peer data transfer tool.",
		Version:              verTag,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			receive.Command,
			send.Command,
		},
		Before: func(c *cli.Context) error {
			if c.Bool("debug") {
				log.SetLevel(log.DebugLevel)
			}
			return nil
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "enables debug log output",
			},
			&cli.BoolFlag{
				Name:  "dht",
				Usage: "Only advertise via the DHT",
			},
			&cli.BoolFlag{
				Name:  "mdns",
				Usage: "Only advertise via multicast DNS",
			},
			&cli.BoolFlag{
				Name:   "homebrew",
				Usage:  "if set transfers a hard coded file with a hard coded word sequence",
				Hidden: true,
			},
		},
	}

	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		<-sigs
		log.Infoln("Stopping...")
		signal.Stop(sigs)
		cancel()
	}()

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Infof("error: %v\n", err)
		os.Exit(1)
	}
}
