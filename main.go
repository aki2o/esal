package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
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
				
				for {
					fmt.Print("> ")
					
					scanner := bufio.NewScanner(os.Stdin)
					scanner.Scan()
					
					input_tokens := strings.Fields(scanner.Text())
					if len(input_tokens) == 0 { continue }
					
					command_name	:= input_tokens[0]
					processor		:= action.NewProcessor(command_name)
					if processor == nil {
						log.WithFields(log.Fields{ "command": command_name }).Debug("unknown command")
						os.Stderr.Write([]byte("Unknown command!\n"))
						continue
					}
					
					flagset := flag.NewFlagSet("action", flag.PanicOnError)
					processor.SetOption(flagset)
					
					err := flagset.Parse(input_tokens[1:])
					if err != nil {
						action.PutError(err)
						continue
					}
					
					err = processor.Do(flagset.Args())
					if err != nil {
						action.PutError(err)
						continue
					}
				}
			},
		},
		{
			Name: "config",
			Action: func(c *cli.Context) error {
			},
		},
	}
	
	app.Run(os.Args)
}
