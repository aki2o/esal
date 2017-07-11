package action

import (
	"flag"
)

type pupload struct {
	real *upload
}

func init() {
	addProcessor(&pupload{ real: &upload{} }, "pupload", "Shortcut of upload -peco.")
}

func (self *pupload) SetOption(flagset *flag.FlagSet) {
	self.real.SetOption(flagset)
}

func (self *pupload) Do(args []string) error {
	self.real.pecolize = true
	return self.real.Do(args)
}
