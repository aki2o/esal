package action

import (
	osexec "os/exec"
	"github.com/aki2o/esal/util"
)

type exec struct {
	util.ProcessIO
}

func init() {
	registProcessor(func() util.Processable { return &exec{} }, "exec", "Execute os command.", "STRING")
}

func (self *exec) Do(args []string) error {
	cmd := osexec.Command(args[0], args[1:]...)
	cmd.Stdout = self.Writer
	cmd.Stdin  = self.Reader
	cmd.Start()
	cmd.Wait()
	return nil
}
