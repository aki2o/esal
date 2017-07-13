package main

import (
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/aki2o/esa-cui/util"
	"github.com/aki2o/esa-cui/action"
	"github.com/aki2o/esa-cui/config"
)

func init() {
	fp, err := os.OpenFile("/tmp/esa.log", os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0666)
	if err != nil { panic(err) }
	
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(fp)
	log.SetLevel(log.DebugLevel)
}

func main() {
	app := cli.NewApp()
	app.Name = "esa"
	app.Commands = []cli.Command{
		{
			Name: "login",
			Usage: "Start prompt to access posts.",
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				access_token := os.Getenv("ESA_CUI_ACCESS_TOKEN")
				
				config.Load(team)
				
				if err := action.SetupContext(team, access_token); err != nil { panic(err) }

				util.ProcessInteractive("action", action.ProcessorRepository())

				return nil
			},
		},
		{
			Name: "config",
			Usage: "Start prompt to configure.",
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				
				config.Load(team)
				
				util.ProcessInteractive("config", config.ProcessorRepository())

				return nil
			},
		},
		{
			Name: "sync",
			Usage: "Synchronize posts.",
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				access_token := os.Getenv("ESA_CUI_ACCESS_TOKEN")
				
				config.Load(team)
				
				if err := action.SetupContext(team, access_token); err != nil { panic(err) }

				repo			:= action.ProcessorRepository()
				processor_name	:= "sync"

				adapter := &util.IshellAdapter{
					Processor: repo.GetProcessor(processor_name),
					ProcessorName: processor_name,
					ProcessorUsage: repo.GetUsage(processor_name),
				}

				adapter.Run(ctx.Args()[1:])
				return nil
			},
		},
		{
			Name: "members",
			Usage: "Print members.",
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				access_token := os.Getenv("ESA_CUI_ACCESS_TOKEN")
				
				config.Load(team)
				
				if err := action.SetupContext(team, access_token); err != nil { panic(err) }

				repo			:= action.ProcessorRepository()
				processor_name	:= "members"

				adapter := &util.IshellAdapter{
					Processor: repo.GetProcessor(processor_name),
					ProcessorName: processor_name,
					ProcessorUsage: repo.GetUsage(processor_name),
				}

				adapter.Run(ctx.Args()[1:])
				return nil
			},
		},
	}
	
	app.Run(os.Args)
}
