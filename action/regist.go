package action

import (
	"errors"
	"io/ioutil"
	"regexp"
	"path/filepath"
	"fmt"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

type regist struct {
	Wip bool `short:"w" long:"wip" description:"Update the post as wip."`
	Shipping bool `short:"s" long:"ship" description:"Ship the post."`
	Tags []string `short:"t" long:"tag" description:"Tag name labeling tha post."`
	Category string `short:"c" long:"category" description:"Category of the post."`
	PostName string `short:"n" long:"name" description:"Name of the post."`
	Message string `short:"m" long:"message" description:"Commit message."`
}

func init() {
	registProcessor(func() util.Processable { return &regist{} }, "regist", "Regist a post.", "[OPTIONS] POST")
}

func (self *regist) Do(args []string) error {
	var regist_file_path = ""
	if len(args) > 0 { regist_file_path = args[0] }

	if regist_file_path == "" {
		return errors.New("Require regist file path!")
	}

	body_bytes, err := ioutil.ReadFile(regist_file_path)
	if err != nil { return err }

	wip := true
	if self.Wip { wip = true }
	if self.Shipping { wip = false }

	tags := self.Tags

	category := Context.Cwd
	if self.Category != "" {
		re, _ := regexp.Compile("^/")
		category = re.ReplaceAllString(self.Category, "")
	}

	re, _ := regexp.Compile("\\.[^.]+$")
	post_name := re.ReplaceAllString(filepath.Base(regist_file_path), "")
	if self.PostName != "" { post_name = self.PostName }
	
	post := esa.Post{
		Name: post_name,
		BodyMd: string(body_bytes),
		Tags: tags,
		Category: category,
		Wip: wip,
		Message: self.Message,
	}

	fmt.Println("Start upload...")
	res, err := Context.Client.Post.Create(Context.Team, post)
	if err != nil { return err }
	fmt.Println("Finished upload.")

	err = SavePost(res)
	if err != nil { return err }
	
	return nil
}
