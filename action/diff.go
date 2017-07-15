package action

import (
	"flag"
	"errors"
	"io"
	"strconv"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type diff struct {
	pecolize bool
	recursive bool
}

func init() {
	addProcessor(&diff{}, "diff", "Diff a post between upstream and local.")
}

func (self *diff) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
	flagset.BoolVar(&self.recursive, "r", false, "Recursively for peco.")
}

func (self *diff) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if self.pecolize {
		next_path, err := self.runPeco(path)
		if err != nil { return err }

		path = next_path
	}

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

func (self *diff) runPeco(path string) (string, error) {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()
		
		ls := &ls{ writer: writer, recursive: self.recursive, file_only: true }
		ls.printNodesIn(path, AbsolutePathOf(path))
	}

	return pipePeco(provider)
}