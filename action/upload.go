package action

import (
	"flag"
	"errors"
	"strconv"
	"encoding/json"
	"fmt"
	"github.com/upamune/go-esa/esa"
)

type upload struct {}

func init() {
	addProcessor(&upload{}, "upload", "Upload a post.")
}

func (self *upload) SetOption(flagset *flag.FlagSet) {
}

func (self *upload) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	dir_path, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	var post esa.Post
	bytes := LoadPostData(AbsolutePathOf(dir_path), post_number, "json")
	if err := json.Unmarshal(bytes, &post); err != nil { return err }
	
	bytes = LoadPostData(AbsolutePathOf(dir_path), post_number, "md")
	post.BodyMd = string(bytes)

	fmt.Println("Start upload...")
	post_number_i, _ := strconv.Atoi(post_number)
	res, err := Context.Client.Post.Update(Context.Team, post_number_i, post)
	if err != nil { return err }
	SavePost(res)

	fmt.Println("Finished upload.")
	return nil
}
