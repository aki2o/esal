package action

import (
	"errors"
	"strconv"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/aki2o/esal/util"
)

type diff struct {
	pecoable
}

func init() {
	registProcessor(func() util.Processable { return &diff{} }, "diff", "Diff a post between upstream and local.", "[OPTIONS] POST...")
}

func (self *diff) Do(args []string) error {
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

func (self *diff) process(path string) error {
	_, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	body_bytes, err := LoadPostBody(post_number)
	if err != nil { return err }

	post_number_as_int, _ := strconv.Atoi(post_number)
	latest_postres, err := Context.Client.Post.GetPost(Context.Team, post_number_as_int)
	if err != nil { return err }

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(body_bytes), latest_postres.BodyMd, false)
	fmt.Println(dmp.DiffPrettyText(diffs))

	return nil
}
