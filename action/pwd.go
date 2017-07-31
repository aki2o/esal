package action

import (
	"github.com/aki2o/esal/util"
)

type pwd struct {
	util.ProcessIO
}

func init() {
	RegistProcessor(func() util.Processable { return &pwd{} }, "pwd", "Print current work category.", "")
}

func (self *pwd) Do(args []string) error {
	self.Println("/"+CategoryOf(Context.Cwd))
	return nil
}
