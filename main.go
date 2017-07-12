package main

import (
	"os"
	"flag"
	"fmt"
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

				repo := action.ProcessorRepository()
				processor := repo.GetProcessor("sync")
				
				flagset := flag.NewFlagSet("sync", flag.PanicOnError)

				var help_required bool = false
				flagset.BoolVar(&help_required, "h", false, "Show help.")
	
				processor.SetOption(flagset)

				err := flagset.Parse(ctx.Args()[1:])
				if err != nil {	return err }

				if help_required {
					fmt.Fprintf(os.Stderr, "%s: %s\n\nOptions:\n", "sync", repo.GetUsage("sync"))
					flagset.PrintDefaults()
					return nil
				}
				
				err = processor.Do(flagset.Args())
				if err != nil { return err }

				return nil
			},
		},
	}
	
	app.Run(os.Args)
}
