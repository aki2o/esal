package action

import (
	"flag"
	"os"
)

type exit struct {}

func init() {
	processors["exit"] = &exit{}
	processors["quit"] = &exit{}
}

func (self *exit) SetOption(flagset *flag.FlagSet) {
}

func (self *exit) Do(args []string) error {
	os.Exit(0)
	return nil
}
