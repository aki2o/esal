package action

import (
	"errors"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type update struct {
	*pecoable
	Wip bool `short:"w" long:"wip" description:"Update the post as wip."`
	Shipping bool `short:"s" long:"ship" description:"Ship the post."`
	Tags []string `short:"t" long:"tag" description:"Tag name labeling tha post."`
	Category string `short:"c" long:"category" description:"Category of the post."`
	PostName string `short:"n" long:"name" description:"Name of the post."`
	Message string `short:"m" long:"message" description:"Commit message."`
	WithoutBody bool `short:"n" long:"nobody" description:"Exec without body."`
	KeepLockRequired bool `short:"k" long:"keeplock" description:"Exec without unlock."`
}

func init() {
	registProcessor(func() util.Processable { return &update{} }, "update", "Update a post.", "[OPTIONS] POST...")
}

func (self *update) Do(args []string) error {
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

func (self *update) process(path string) error {
	_, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	post, err := self.loadPostDataWithVerify(post_number)
	if err != nil { return err }
	
	new_post := esa.Post{
		Message: self.Message,
		OriginalRevision: esa.PostOriginalRevision {
			Number: post.RevisionNumber,
			User: post.UpdatedBy.ScreenName,
		},
	}

	self.setWip(&new_post, post)
	self.setTags(&new_post, post)
	self.setCategory(&new_post, post)
	self.setName(&new_post, post)
	if err := self.setBody(&new_post, post_number); err != nil { return err }
	
	fmt.Println("Start upload...")
	res, err := Context.Client.Post.Update(Context.Team, post.Number, new_post)
	if err != nil { return err }
	fmt.Println("Finished upload.")
	
	if res.OverLapped { fmt.Fprintf(os.Stderr, "Conflict happened in update '%s'!!!\n", path) }

	err = SavePost(res)
	if err != nil { return err }
	
	if ! self.KeepLockRequired {
		unlock_process := &unlock{}
		if err := unlock_process.Do([]string{ path }); err != nil { return err }
	}
	
	return nil
}

func (self *update) loadPostDataWithVerify(post_number string) (*esa.PostResponse, error) {
	json_bytes, err := LoadPostData(post_number)
	if err != nil { return nil, err }

	var post esa.PostResponse
	if err = json.Unmarshal(json_bytes, &post); err != nil { return nil, err }

	latest_post, err := Context.Client.Post.GetPost(Context.Team, post.Number)
	if err != nil { return nil, err }

	if latest_post.RevisionNumber != post.RevisionNumber {
		return nil, errors.New("Post has been updated by other user!")
	}

	return &post, nil
}

func (self *update) setWip(new_post *esa.Post, post *esa.PostResponse) {
	wip := post.Wip
	
	if self.Wip { wip = true }
	if self.Shipping { wip = false }

	new_post.Wip = wip
}

func (self *update) setTags(new_post *esa.Post, post *esa.PostResponse) {
	tags := post.Tags
	
	if len(self.Tags) > 0 { tags = self.Tags }

	new_post.Tags = tags
}

func (self *update) setCategory(new_post *esa.Post, post *esa.PostResponse) {
	category := post.Category
	
	if self.Category != "" {
		re, _ := regexp.Compile("^/")
		category = re.ReplaceAllString(self.Category, "")
	}

	new_post.Category = category
}

func (self *update) setName(new_post *esa.Post, post *esa.PostResponse) {
	post_name := post.Name
	
	if self.PostName != "" { post_name = self.PostName }

	new_post.Name = post_name
}

func (self *update) setBody(new_post *esa.Post, post_number string) error {
	body_bytes, err := LoadPostBody(post_number)
	if err != nil { return err }
	
	lock_bytes, err := LoadPostLock(post_number)
	if err != nil { lock_bytes = body_bytes }

	if self.WithoutBody {
		new_post.BodyMd = string(lock_bytes)
	} else {
		new_post.BodyMd = string(body_bytes)
	}
	new_post.OriginalRevision.BodyMd = string(lock_bytes)
	return nil
}
