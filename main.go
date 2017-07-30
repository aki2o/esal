package main

import (
	"os"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/aki2o/esal/util"
	"github.com/aki2o/esal/action"
	"github.com/aki2o/esal/config"
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
				cli.BoolFlag{
					Name: "use-peco",
				},
			},
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				
				config.Load(team)
				
				if err := action.SetupContext(team, detectAccessToken(ctx, team), true); err != nil { panic(err) }

				action.SetupPeco(ctx.Bool("use-peco"))

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
				
				if err := action.SetupContext(team, detectAccessToken(ctx, team), false); err != nil { panic(err) }

				adapter := &util.IshellAdapter{
					ProcessorRepository: action.ProcessorRepository(),
					ProcessorName: "sync",
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
				
				if err := action.SetupContext(team, detectAccessToken(ctx, team), false); err != nil { panic(err) }

				adapter := &util.IshellAdapter{
					ProcessorRepository: action.ProcessorRepository(),
					ProcessorName: "members",
				}

				adapter.Run(ctx.Args()[1:])
				return nil
			},
		},
	}
	
	app.Run(os.Args)
}

func detectAccessToken(ctx *cli.Context, team string) string {
	access_token := ctx.String("access-token")
	if access_token == "" { access_token = os.Getenv("ESAL_ACCESS_TOKEN_"+strings.ToUpper(team)) }
	if access_token == "" { access_token = util.ReadAccessToken() }

	return access_token
}
