package action

import (
	"errors"
	"github.com/aki2o/esa-cui/util"
)

type lock struct {
	Pecolize bool `short:"p" long:"peco" description:"Exec with peco."`
}

func init() {
	registProcessor(func() util.Processable { return &lock{} }, "lock", "Start to guard a post from updated by SYNC.", "[OPTIONS] POST")
}

func (self *lock) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if self.Pecolize {
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
