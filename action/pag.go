package action

import (
	"flag"
)

type pag struct {
	real *ag
}

func init() {
	addProcessor(&pag{ real: &ag{} }, "pag", "Shortcut of ag -peco.")
}

func (self *pag) SetOption(flagset *flag.FlagSet) {
	self.real.SetOption(flagset)
}

func (self *pag) Do(args []string) error {
	self.real.pecolize = true
	return self.real.Do(args)
}
