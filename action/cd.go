package action

import (
	"errors"
	"fmt"
	"flag"
	"path/filepath"
	"os"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/utf8string"
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

	var next_path string
	if utf8string.NewString(path).Slice(0, 1) == "/" {
		next_path = filepath.Join(Context.Root(), path)
	} else {
		next_path = filepath.Join(Context.Cwd, path)
	}

	info, err := os.Stat(next_path)
	if err != nil { return err }
	if ! info.IsDir() { return fmt.Errorf("Not directory : %s", next_path) }

	Context.Cwd = next_path
	log.WithFields(log.Fields{ "cwd": Context.Cwd }).Info("update cwd")

	return nil
}
