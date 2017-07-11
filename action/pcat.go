package action

import (
	"flag"
)

type pcat struct {
	real *cat
}

func init() {
	addProcessor(&pcat{ real: &cat{} }, "pcat", "Shortcut of cat -peco.")
}

func (self *pcat) SetOption(flagset *flag.FlagSet) {
	self.real.SetOption(flagset)
}

func (self *pcat) Do(args []string) error {
	self.real.pecolize = true
	return self.real.Do(args)
}
