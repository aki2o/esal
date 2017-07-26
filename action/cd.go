package action

import (
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/esal/util"
)

type cd struct {
	*pecoable
}

func init() {
	registProcessor(func() util.Processable { return &cd{} }, "cd", "Change working category.", "[OPTIONS] CATEGORY")
}

func (self *cd) Do(args []string) error {
	var path string = ""
	var err error
	if len(args) > 0 { path = args[0] }

	var next_abs_path string = ""
	if self.PecoRequired() {
		next_path, err := selectNodeByPeco(path, true)
		if err != nil { return err }

		next_abs_path = PhysicalPathOf(next_path[0])
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
