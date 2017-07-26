package action

import (
	"errors"
	"github.com/aki2o/esal/util"
)

type lock struct {
	*pecoable
}

func init() {
	registProcessor(func() util.Processable { return &lock{} }, "lock", "Start to guard a post from updated by SYNC.", "[OPTIONS] POST...")
}

func (self *lock) Do(args []string) error {
	if self.PecoRequired() {
		var path string = ""
		var err error
	
		if len(args) > 0 { path = args[0] }

		args, err = selectNodeByPeco(path, false)
		if err != nil { return err }
	}

	for _, path := range args {
		if err := self.process(path); err != nil { return err }
	}
	return nil
}

func (self *lock) process(path string) error {
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
