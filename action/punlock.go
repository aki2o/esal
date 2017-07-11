package action

import (
	"flag"
)

type punlock struct {
	real *unlock
}

func init() {
	addProcessor(&punlock{ real: &unlock{} }, "punlock", "Shortcut of unlock -peco.")
}

func (self *punlock) SetOption(flagset *flag.FlagSet) {
	self.real.SetOption(flagset)
}

func (self *punlock) Do(args []string) error {
	self.real.pecolize = true
	return self.real.Do(args)
}
