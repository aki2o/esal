package action

import (
	"fmt"
	"flag"
	"os"
	"io"
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
		next_path, err := self.runPeco(path)
		if err != nil { return err }

		next_abs_path = PhysicalPathOf(next_path)
	} else {
		next_abs_path = PhysicalPathOf(path)
	}
	
	info, err := os.Stat(next_abs_path)
	if err != nil { return err }
	if ! info.IsDir() { return fmt.Errorf("Not directory : %s", next_abs_path) }

	Context.Cwd = next_abs_path
	log.WithFields(log.Fields{ "cwd": Context.Cwd }).Info("update cwd")

	return nil
}

func (self *cd) runPeco(path string) (string, error) {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()
		
		ls := &ls{ writer: writer, recursive: true, directory_only: true }
		ls.printNodesIn(path, PhysicalPathOf(path))
	}

	return pipePeco(provider)
}

