package action

import (
	"errors"
	"flag"
	"fmt"
	"encoding/json"
	"github.com/upamune/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

type cat struct {
	json_format bool
}

type postProperty struct {
	esa.PostResponse
	LocalPath string `json:"local_path"`
	Locked bool `json:"locked"`
}

func init() {
	processors["cat"] = &cat{}
}

func (self *cat) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.json_format, "json", false, "Show properties as json.")
}

func (self *cat) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	dir_path, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	if self.json_format {
		bytes := LoadPostData(AbsolutePathOf(dir_path), post_number, "json")

		var post postProperty
		if err := json.Unmarshal(bytes, &post); err != nil { return err }
		
		post.LocalPath	= GetLocalPostPath(post.Category, post.Number, "md")
		post.Locked		= util.Exists(GetLocalPostPath(post.Category, post.Number, "lock"))

		json_bytes, _ := json.MarshalIndent(post, "", "\t")
		
		fmt.Println(string(json_bytes))
	} else {
		bytes := LoadPostData(AbsolutePathOf(dir_path), post_number, "md")
		
		fmt.Println(string(bytes))
	}
	return nil
}
