package action

import (
	"flag"
	"os/exec"
	"fmt"
)

type ag struct {}

func init() {
	processors["ag"] = &ag{}
}

func (self *ag) SetOption(flagset *flag.FlagSet) {
}

func (self *ag) Do(args []string) error {
	cmd_args := append([]string{"-G", ".md$"}, args...)
	cmd_args = append(cmd_args, Context.Cwd)
	
	out, err := exec.Command("ag", cmd_args...).Output()
	if err != nil { return err }
	
	fmt.Print(string(out))
	return nil
}
