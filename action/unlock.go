package action

import (
	"errors"
	"os"
	"github.com/aki2o/esal/util"
)

type unlock struct {
	util.ProcessIO
	pecoable
	ListRequired bool `short:"l" long:"list" description:"List unlocked all posts."`
}

func init() {
	registProcessor(func() util.Processable { return &unlock{} }, "unlock", "Stop to guard a post from updated by SYNC.", "[OPTIONS] POST...")
}

func (self *unlock) Do(args []string) error {
	if len(args) == 0 { args = self.ScanArgs() }

	if len(args) == 0 && self.PecoRequired() {
		var category_only bool
		if self.ListRequired {
			category_only = true
		} else {
			category_only = false
		}
		
		var err error
		args, err = selectNodeByPeco("", category_only)
		if err != nil { return err }
	}

	for _, path := range args {
		if err := self.process(path); err != nil { return err }
	}
	return nil
}

func (self *unlock) process(path string) error {
	if self.ListRequired {
		return self.printPosts(path)
	} else {
		return self.unlockPost(path)
	}
}

func (self *unlock) printPosts(path string) error {
	find_process := &find{}
	find_process.Type = "u"
	node_paths, err := find_process.collectNodesIn(path)
	if err != nil { return err }

	for _, node_path := range node_paths {
		self.Println(node_path)
	}

	return nil
}

func (self *unlock) unlockPost(path string) error {
	_, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_file_path := GetPostLockPath(post_number)
	if ! util.Exists(lock_file_path) { return nil }
	
	if err := os.Remove(lock_file_path); err != nil { return err }
	
	return nil
}
