package action

import (
	"github.com/aki2o/esal/util"
)

type Exit struct {
	util.ProcessIO
}

func (self *Exit) Do(args []string) error {
	return &util.ProcessorExitRequired{}
}
