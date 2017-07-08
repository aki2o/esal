package action

import (
	"errors"
	"fmt"
	"flag"
	"os"
	log "github.com/sirupsen/logrus"
)

type cd struct {}

func init() {
	processors["cd"] = &cd{}
}

func (self *cd) SetOption(flagset *flag.FlagSet) {
}

func (self *cd) Do(args []string) error {
	var path string = ""

	if len(args) > 0 { path = args[0] }

	if path == "" {
		return errors.New("Require path")
	}

	next_path := AbsolutePathOf(path)

	info, err := os.Stat(next_path)
	if err != nil { return err }
	if ! info.IsDir() { return fmt.Errorf("Not directory : %s", next_path) }

	Context.Cwd = next_path
	log.WithFields(log.Fields{ "cwd": Context.Cwd }).Info("update cwd")

	return nil
}
