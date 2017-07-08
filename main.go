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
	// fp, err := os.OpenFile("/tmp/esa.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0666)
	// if err != nil { panic(err) }
	
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)
}

func main() {
	app := cli.NewApp()
	app.Name = "esa"
	app.Commands = []cli.Command{
		{
			Name: "login",
			Usage: "start prompt to access posts.",
			Action: func(c *cli.Context) error {
				team := c.Args().First()
				access_token := os.Getenv("ESA_CUI_ACCESS_TOKEN")
				
				config.Load(team)
				
				if err := action.SetupContext(team, access_token); err != nil { panic(err) }

				util.ProcessInteractive("action", action.NewProcessor)

				return nil
			},
		},
		{
			Name: "config",
			Usage: "configure.",
			Action: func(c *cli.Context) error {
				team := c.Args().First()
				
				config.Load(team)
				
				util.ProcessInteractive("config", config.NewProcessor)

				return nil
			},
		},
	}
	
	app.Run(os.Args)
}
