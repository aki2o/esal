package action

import (
	"flag"
	"errors"
	"os"
	"io"
	"github.com/aki2o/esa-cui/util"
)

type unlock struct {
	pecolize bool
}

func init() {
	addProcessor(&unlock{}, "unlock", "Stop to guard a post from updated by SYNC.")
}

func (self *unlock) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
}

func (self *unlock) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if self.pecolize {
		next_path, err := self.runPeco(path)
		if err != nil { return err }

		path = next_path
	}

	dir_path, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_file_path := AbsolutePathOf(dir_path)+"/"+GetLocalPostFileName(post_number, "lock")
	if ! util.Exists(lock_file_path) { return nil }
	
	if err := os.Remove(lock_file_path); err != nil { return err }
	
	return nil
}

func (self *unlock) runPeco(path string) (string, error) {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()
		
		ls := &ls{ writer: writer, recursive: true, file_only: true }
		ls.printNodesIn(path, AbsolutePathOf(path))
	}

	return pipePeco(provider)
}
