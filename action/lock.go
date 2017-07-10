package action

import (
	"flag"
	"errors"
	"os"
	"github.com/aki2o/esa-cui/util"
)

type lock struct {}

func init() {
	addProcessor(&lock{}, "lock", "Start to guard a post from updated by SYNC.")
}

func (self *lock) SetOption(flagset *flag.FlagSet) {
}

func (self *lock) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	dir_path, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_file_path := AbsolutePathOf(dir_path)+"/"+GetLocalPostFileName(post_number, "lock")
	if util.Exists(lock_file_path) { return nil }
	
	fp, err := os.Create(lock_file_path)
	if err != nil { return err }
	
	fp.Close()
	return nil
}
