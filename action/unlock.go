package action

import (
	"flag"
	"errors"
	"os"
	"github.com/aki2o/esa-cui/util"
)

type unlock struct {}

func init() {
	addProcessor(&unlock{}, "unlock", "Stop to guard a post from updated by SYNC.")
}

func (self *unlock) SetOption(flagset *flag.FlagSet) {
}

func (self *unlock) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	dir_path, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_file_path := AbsolutePathOf(dir_path)+"/"+GetLocalPostFileName(post_number, "lock")
	if ! util.Exists(lock_file_path) { return nil }
	
	if err := os.Remove(lock_file_path); err != nil { return err }
	
	return nil
}
