package action

import (
	"errors"
	"io/ioutil"
	"regexp"
	"path/filepath"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type regist struct {
	util.ProcessIO
	uploadable
}

func init() {
	registProcessor(func() util.Processable { return &regist{} }, "regist", "Regist a post.", "[OPTIONS] FILE_PATH")
}

func (self *regist) Do(args []string) error {
	var regist_file_path = ""
	if len(args) > 0 { regist_file_path = args[0] }

	if regist_file_path == "" {
		return errors.New("Require regist file path!")
	}

	re, _ := regexp.Compile("\\.[^.]+$")
	post_name := re.ReplaceAllString(filepath.Base(regist_file_path), "")
	
	post := esa.Post{}
	self.setWip(&post, true)
	self.setTags(&post, []string{})
	self.setCategory(&post, CategoryOf(Context.Cwd))
	self.setName(&post, post_name)
	self.setMessage(&post)
	if err := self.setBody(&post, regist_file_path); err != nil { return err }
	
	self.Println("Start upload...")
	res, err := Context.Client.Post.Create(Context.Team, post)
	if err != nil { return err }
	self.Println("Finished upload.")

	err = SavePost(res)
	if err != nil { return err }

	self.Printf("Registed %d: %s.", res.Number, res.FullName)
	return nil
}

func (self *regist) setBody(post *esa.Post, file_path string) error {
	body_bytes, err := ioutil.ReadFile(file_path)
	if err != nil { return err }

	post.BodyMd = string(body_bytes)
	return nil
}
