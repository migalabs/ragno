package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cortze/ragno/cmd"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)


var (
	AppName = "Ragno"
	AppVersion = "v0.0.1-alpha"
)

func main() {
	printVersion()

	app := &cli.App{
		Name: AppName,
		Usage: "Light software to crawl and identify peers in Ethereum's Execution Layer", 
		UsageText: "ragno [commands] [options...]",
		Authors: []*cli.Author{
			{
				Name: "@cortze",
				Email: "cortze@protonmail.com",
			},
		},
		EnableBashCompletion: true, 
		Commands: []*cli.Command{
			cmd.RunCommand,
			cmd.Discv4Cmd,
			cmd.ConnectCmd,
		},
	}
	
	// run the tool and check if there is any error reported
	err := app.RunContext(context.Background(), os.Args)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}


func printVersion() {
	fmt.Println(AppName+"-"+AppVersion)
}
