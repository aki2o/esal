package action

import (
	"fmt"
	"flag"
	"os"
	log "github.com/sirupsen/logrus"
)

type cd struct {
	pecolize bool
}

func init() {
	addProcessor(&cd{}, "cd", "Change working category.")
}

func (self *cd) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
}

func (self *cd) Do(args []string) error {
	var path string = ""
	var err error
	if len(args) > 0 { path = args[0] }

	var next_abs_path string = ""
	if self.pecolize {
		next_path, err := selectNodeByPeco(path, true)
		if err != nil { return err }

		next_abs_path = PhysicalPathOf(next_path)
	} else {
		next_abs_path = PhysicalPathOf(path)
	}
	
	info, err := os.Stat(next_abs_path)
	if err != nil { return fmt.Errorf("Failed to get stat of '%s' : %s", next_abs_path, err.Error()) }
	if ! info.IsDir() { return fmt.Errorf("Not directory : %s", next_abs_path) }

	Context.Cwd = next_abs_path
	log.WithFields(log.Fields{ "cwd": Context.Cwd }).Info("update cwd")

	return nil
}
