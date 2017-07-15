package action

import (
	"flag"
	"errors"
	"strconv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"github.com/aki2o/go-esa/esa"
)

type update struct {
	pecolize bool
	recursive bool
}

func init() {
	addProcessor(&update{}, "update", "Update a post.")
}

func (self *update) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
	flagset.BoolVar(&self.recursive, "r", false, "Recursively for peco.")
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
	
	post := esa.Post{
		Name: postres.Name,
		BodyMd: string(body_bytes),
		Tags: postres.Tags,
		Category: postres.Category,
		Wip: postres.Wip,
		// UpdatedBy: "",
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
