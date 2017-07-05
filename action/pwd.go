package action

import (
	"flag"
	"fmt"
)

type pwd struct {}

func init() {
	processors["pwd"] = &pwd{}
}

func (self *pwd) SetOption(flagset *flag.FlagSet) {
}

func (self *pwd) Do(args []string) error {
	fmt.Println(Context.Cwd)
	return nil
}
