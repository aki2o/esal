package action

import (
	"flag"
)

type pupdate struct {
	real *update
}

func init() {
	addProcessor(&pupdate{ real: &update{} }, "pupdate", "Shortcut of update -peco.")
}

func (self *pupdate) SetOption(flagset *flag.FlagSet) {
	self.real.SetOption(flagset)
}

func (self *pupdate) Do(args []string) error {
	self.real.pecolize = true
	return self.real.Do(args)
}
