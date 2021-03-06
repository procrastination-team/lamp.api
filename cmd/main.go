package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/procrastination-team/lamp.api/pkg/api"
	"github.com/procrastination-team/lamp.api/pkg/config"
)

var cfgFile string

func init() {
	flag.StringVar(&cfgFile, "config", "", "path to config file")
}

func main() {
	flag.Parse()
	conf, err := config.Init(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lamp, err := api.New(conf, ctx)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go lamp.Run()

	<-done
	signal.Stop(done)
}
