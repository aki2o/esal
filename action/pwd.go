package action

import (
	"flag"
	"fmt"
)

type pwd struct {}

func init() {
	addProcessor(&pwd{}, "pwd", "Print current work category.")
}

func (self *pwd) SetOption(flagset *flag.FlagSet) {
}

func (self *pwd) Do(args []string) error {
	fmt.Println(CategoryOf(Context.Cwd))
	return nil
}
