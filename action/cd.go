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
	processors["cd"] = &cd{}
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
		next_path, err := runPeco(path, AbsolutePathOf(path))
		if err != nil { return err }

		next_abs_path = AbsolutePathOf(next_path)
	} else {
		next_abs_path = AbsolutePathOf(path)
	}
	
	info, err := os.Stat(next_abs_path)
	if err != nil { return err }
	if ! info.IsDir() { return fmt.Errorf("Not directory : %s", next_abs_path) }

	Context.Cwd = next_abs_path
	log.WithFields(log.Fields{ "cwd": Context.Cwd }).Info("update cwd")

	return nil
}

func runPeco(path string, abs_path string) (string, error) {
	provider := func(writer *io.PipeWriter) {
		ls := &ls{ writer: writer, recursive: true }
		ls.printNodesIn(path, abs_path)
		
		writer.Close()
	}

	return pipePeco(provider)
}

