package action

import (
	"errors"
	"os"
	"github.com/aki2o/esal/util"
)

type unlock struct {
	*pecoable
}

func init() {
	registProcessor(func() util.Processable { return &unlock{} }, "unlock", "Stop to guard a post from updated by SYNC.", "[OPTIONS] POST")
}

func (self *unlock) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if self.PecoRequired() {
		next_path, err := selectNodeByPeco(path, false)
		if err != nil { return err }

		path = next_path
	}

	_, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_file_path := GetPostLockPath(post_number)
	if ! util.Exists(lock_file_path) { return nil }
	
	if err := os.Remove(lock_file_path); err != nil { return err }
	
	return nil
}
