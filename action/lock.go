package action

import (
	"flag"
	"errors"
	"github.com/aki2o/esa-cui/util"
)

type lock struct {
	pecolize bool
	recursive bool
}

func init() {
	addProcessor(&lock{}, "lock", "Start to guard a post from updated by SYNC.")
}

func (self *lock) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
	flagset.BoolVar(&self.recursive, "r", false, "Recursively for peco.")
}

func (self *lock) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if self.pecolize {
		next_path, err := selectNodeByPeco(path, false)
		if err != nil { return err }

		path = next_path
	}

	_, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_file_path := GetPostLockPath(post_number)
	if util.Exists(lock_file_path) { return nil }

	bytes, err := LoadPostBody(post_number)
	if err != nil { return err }

	err = util.CreateFile(lock_file_path, string(bytes))
	if err != nil { return err }

	return nil
}
