package action

import (
	osexec "os/exec"
	"strings"
	"github.com/aki2o/esal/util"
)

type exec struct {
	util.ProcessIO
	ShellCommand string `short:"s" long:"shell" description:"Execute by SHELL." value-name:"SHELL"`
}

func init() {
	RegistProcessor(func() util.Processable { return &exec{} }, "exec", "Execute os command.", "STRING")
}

func (self *exec) Do(args []string) error {
	var cmd *osexec.Cmd
	
	if self.ShellCommand != "" {
		cmd = osexec.Command(self.ShellCommand, "-c", strings.Join(args, " "))
	} else {
		cmd = osexec.Command(args[0], args[1:]...)
	}
	
	cmd.Stdout = self.Writer
	cmd.Stdin  = self.Reader
	cmd.Start()
	cmd.Wait()
	
	return nil
}
