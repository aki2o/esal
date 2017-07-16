package action

import (
	"flag"
	"errors"
	"io/ioutil"
	"strings"
	"regexp"
	"path/filepath"
	"fmt"
	"github.com/aki2o/go-esa/esa"
)

type regist struct {
	wip bool
	ship bool
	tags string
	category string
	post_name string
	message string
}

func init() {
	addProcessor(&regist{}, "regist", "Regist a post.")
}

func (self *regist) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.wip, "wip", false, "Update the post as wip.")
	flagset.BoolVar(&self.ship, "ship", false, "Ship the post.")
	flagset.StringVar(&self.tags, "tags", "", "Tag names separated comma.")
	flagset.StringVar(&self.category, "category", "", "Category.")
	flagset.StringVar(&self.post_name, "name", "", "Name of the post.")
	flagset.StringVar(&self.message, "m", "Update post.", "Commit message.")
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
	if self.wip { wip = true }
	if self.ship { wip = false }

	tags := []string{}
	if self.tags != "" { tags = strings.Split(self.tags, ",") }

	category := Context.Cwd
	if self.category != "" {
		re, _ := regexp.Compile("^/")
		category = re.ReplaceAllString(self.category, "")
	}

	re, _ := regexp.Compile("\\.[^.]+$")
	post_name := re.ReplaceAllString(filepath.Base(regist_file_path), "")
	if self.post_name != "" { post_name = self.post_name }
	
	post := esa.Post{
		Name: post_name,
		BodyMd: string(body_bytes),
		Tags: tags,
		Category: category,
		Wip: wip,
		Message: self.message,
	}

	fmt.Println("Start upload...")
	res, err := Context.Client.Post.Create(Context.Team, post)
	if err != nil { return err }
	fmt.Println("Finished upload.")

	err = SavePost(res)
	if err != nil { return err }
	
	return nil
}
