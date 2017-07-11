package action

import (
	"flag"
)

type pcd struct {
	real *cd
}

func init() {
	addProcessor(&pcd{ real: &cd{} }, "pcd", "Shortcut of cd -peco.")
}

func (self *pcd) SetOption(flagset *flag.FlagSet) {
	self.real.SetOption(flagset)
}

func (self *pcd) Do(args []string) error {
	self.real.pecolize = true
	return self.real.Do(args)
}
