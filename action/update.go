package action

import (
	"flag"
	"errors"
	"strconv"
	"encoding/json"
	"fmt"
	"io"
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

	var postres esa.PostResponse
	bytes := LoadPostData(AbsolutePathOf(dir_path), post_number, "json")
	if err := json.Unmarshal(bytes, &postres); err != nil { return err }

	bytes = LoadPostData(AbsolutePathOf(dir_path), post_number, "md")
	
	post := esa.Post{
		Name: postres.Name,
		BodyMd: string(bytes),
		Tags: postres.Tags,
		Category: postres.Category,
		Wip: postres.Wip,
		// UpdatedBy: "",
		OriginalRevision: esa.PostOriginalRevision {
			BodyMd: postres.BodyMd,
			Number: postres.RevisionNumber,
			User: postres.UpdatedBy.ScreenName,
		},
	}

	fmt.Println("Start upload...")
	post_number_i, _ := strconv.Atoi(post_number)
	res, err := Context.Client.Post.Update(Context.Team, post_number_i, post)
	if err != nil { return err }
	fmt.Println("Finished upload.")
	
	SavePost(res)
	if res.OverLapped {
		fmt.Printf("Conflict happened!!!\nFor resolving that, you should do `open %s`.\n", path)
	}

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
