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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "access-token, a",
					Usage: "esa access_token for team",
				},
				cli.BoolFlag{
					Name: "non-interactive",
				},
			},
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				
				config.Load(team)
				
				access_token := ctx.String("access-token")
				if access_token == "" { access_token = util.ReadAccessToken() }

				if err := action.SetupContext(team, access_token, true); err != nil { panic(err) }

				action.SetupPeco()
				
				if ctx.Bool("non-interactive") {
					util.ProcessNonInteractive("action", action.ProcessorRepository())
				} else {
					util.ProcessInteractive("action", action.ProcessorRepository())
				}

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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "access-token, a",
					Usage: "esa access_token for team",
				},
			},
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				
				config.Load(team)
				
				access_token := ctx.String("access_token")
				if access_token == "" { access_token = util.ReadAccessToken() }

				if err := action.SetupContext(team, access_token, false); err != nil { panic(err) }

				repo			:= action.ProcessorRepository()
				processor_name	:= "sync"

				adapter := &util.IshellAdapter{
					Processor: repo.NewProcessor(processor_name),
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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "access-token, a",
					Usage: "esa access_token for team",
				},
			},
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				
				config.Load(team)
				
				access_token := ctx.String("access_token")
				if access_token == "" { access_token = util.ReadAccessToken() }

				if err := action.SetupContext(team, access_token, false); err != nil { panic(err) }

				repo			:= action.ProcessorRepository()
				processor_name	:= "members"

				adapter := &util.IshellAdapter{
					Processor: repo.NewProcessor(processor_name),
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
