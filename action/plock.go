package action

import (
	"flag"
)

type plock struct {
	real *lock
}

func init() {
	addProcessor(&plock{ real: &lock{} }, "plock", "Shortcut of lock -peco.")
}

func (self *plock) SetOption(flagset *flag.FlagSet) {
	self.real.SetOption(flagset)
}

func (self *plock) Do(args []string) error {
	self.real.pecolize = true
	return self.real.Do(args)
}
