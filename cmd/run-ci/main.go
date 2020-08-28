package main

import (
	"context"
	"log"
	"os"

	"github.com/suzuki-shunsuke/run-ci/pkg/cli"
	"github.com/suzuki-shunsuke/run-ci/pkg/signal"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	runner := cli.Runner{}
	ctx, cancel := context.WithCancel(context.Background())
	go signal.Handle(os.Stderr, cancel)
	return runner.Run(ctx, os.Args...)
}
