package action

import (
	"flag"
	"os/exec"
	"fmt"
	"regexp"
)

type ag struct {}

func init() {
	addProcessor(&ag{}, "ag", "Execute ag command.")
}

func (self *ag) SetOption(flagset *flag.FlagSet) {
}

func (self *ag) Do(args []string) error {
	cmd_args := append([]string{"-G", ".md$"}, args...)
	cmd_args = append(cmd_args, Context.Cwd)
	
	out, err := exec.Command("ag", cmd_args...).Output()
	if err != nil { return err }

	local_root_re, _ := regexp.Compile("(?m)^"+Context.Root())
	fmt.Print(local_root_re.ReplaceAllString(string(out), ""))
	return nil
}
