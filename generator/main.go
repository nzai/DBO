package main

import (
	"context"
	"log"
	"os"

	"github.com/nzai/dbo/v2/generator/command"
	"github.com/urfave/cli/v3"
)

func main() {

	cmd := &cli.Command{
		Name:     "generator",
		Commands: command.Commands,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
