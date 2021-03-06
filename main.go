package main

import (
	"os"
	"strings"
	"bufio"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/aki2o/esal/util"
	"github.com/aki2o/esal/action"
	"github.com/aki2o/esal/config"
)

func init() {
	fp, err := os.OpenFile("/tmp/esal.log", os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0666)
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
					Usage: "esa access_token for team.",
				},
				cli.BoolFlag{
					Name: "exec, e",
					Usage: "Execute the lines given as second argument or stdin.",
				},
				cli.BoolFlag{
					Name: "non-interactive",
					Usage: "Run process as not interactive shell.",
				},
				cli.BoolFlag{
					Name: "use-peco",
					Usage: "Prefer peco.",
				},
			},
			Action: func(ctx *cli.Context) error {
				team := ctx.Args().First()
				arg1 := ctx.Args().Get(1)
				code := ""

				if ctx.Bool("exec") {
					code = arg1
					if code == "" {
						scanner := bufio.NewScanner(os.Stdin)
						scanner.Scan()
						code = scanner.Text()
					}
				} else if arg1 != "" {
					bytes, err := ioutil.ReadFile(arg1)
					if err != nil {
						util.PutError(err)
						return nil
					}
					code = string(bytes)
				}
				
				config.Load(team)
				
				if err := action.SetupContext(team, detectAccessToken(ctx, team), true); err != nil { panic(err) }

				if code != "" {
					action.RegistProcessor(func() util.Processable { return &action.Exit{} }, "exit", "Exit a process.", "")

					util.ProcessWithString("action", action.ProcessorRepository(), code)
				} else if ctx.Bool("non-interactive") {
					action.RegistProcessor(func() util.Processable { return &action.Exit{} }, "exit", "Exit a process.", "")

					util.ProcessNonInteractive("action", action.ProcessorRepository())
				} else {
					action.SetupPeco(ctx.Bool("use-peco"))

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
	}
	
	app.Run(os.Args)
}

func detectAccessToken(ctx *cli.Context, team string) string {
	access_token := ctx.String("access-token")
	if access_token == "" { access_token = os.Getenv("ESAL_ACCESS_TOKEN_"+strings.ToUpper(team)) }
	if access_token == "" { access_token = util.ReadAccessToken() }

	return access_token
}
