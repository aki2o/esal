package action

import (
	"flag"
	"errors"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"regexp"
	"github.com/aki2o/go-esa/esa"
)

type update struct {
	pecolize bool
	recursive bool
	wip bool
	ship bool
	tags string
	category string
	post_name string
	message string
}

func init() {
	addProcessor(&update{}, "update", "Update a post.")
}

func (self *update) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
	flagset.BoolVar(&self.recursive, "r", false, "Recursively for peco.")
	flagset.BoolVar(&self.wip, "wip", false, "Update the post as wip.")
	flagset.BoolVar(&self.ship, "ship", false, "Ship the post.")
	flagset.StringVar(&self.tags, "tags", "", "Tag names separated comma.")
	flagset.StringVar(&self.category, "category", "", "Category.")
	flagset.StringVar(&self.post_name, "name", "", "Name of the post.")
	flagset.StringVar(&self.message, "m", "Update post.", "Commit message.")
}

func (self *update) Do(args []string) error {
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

	json_bytes, err := LoadPostData(dir_path, post_number)
	if err != nil { return err }

	var postres esa.PostResponse
	if err = json.Unmarshal(json_bytes, &postres); err != nil { return err }

	latest_postres, err := Context.Client.Post.GetPost(Context.Team, postres.Number)
	if err != nil { return err }

	if latest_postres.RevisionNumber != postres.RevisionNumber {
		return errors.New("Post has been updated by other user!")
	}
	
	body_bytes, err := LoadPostBody(post_number)
	if err != nil { return err }
	
	lock_bytes, err := LoadPostLock(post_number)
	if err != nil { lock_bytes = body_bytes }

	wip := postres.Wip
	if self.wip { wip = true }
	if self.ship { wip = false }

	tags := postres.Tags
	if self.tags != "" { tags = strings.Split(self.tags, ",") }

	category := postres.Category
	if self.category != "" {
		re, _ := regexp.Compile("^/")
		category = re.ReplaceAllString(self.category, "")
	}

	post_name := postres.Name
	if self.post_name != "" { post_name = self.post_name }
	
	post := esa.Post{
		Name: post_name,
		BodyMd: string(body_bytes),
		Tags: tags,
		Category: category,
		Wip: wip,
		Message: self.message,
		OriginalRevision: esa.PostOriginalRevision {
			BodyMd: string(lock_bytes),
			Number: postres.RevisionNumber,
			User: postres.UpdatedBy.ScreenName,
		},
	}

	fmt.Println("Start upload...")
	res, err := Context.Client.Post.Update(Context.Team, postres.Number, post)
	if err != nil { return err }
	fmt.Println("Finished upload.")
	
	if res.OverLapped { fmt.Fprintln(os.Stderr, "Conflict happened!") }

	err = SavePost(res)
	if err != nil { return err }

	unlock_process := &unlock{}
	if err := unlock_process.Do([]string{ path }); err != nil { return err }
	
	return nil
}

func (self *update) runPeco(path string) (string, error) {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()
		
		ls := &ls{ writer: writer, recursive: self.recursive, file_only: true }
		ls.printNodesIn(path, AbsolutePathOf(path))
	}

	return pipePeco(provider)
}
