package action

import (
	"flag"
)

type popen struct {
	real *open
}

func init() {
	addProcessor(&popen{ real: &open{} }, "popen", "Shortcut of open -peco.")
}

func (self *popen) SetOption(flagset *flag.FlagSet) {
	self.real.SetOption(flagset)
}

func (self *popen) Do(args []string) error {
	self.real.pecolize = true
	return self.real.Do(args)
}
