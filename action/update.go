package action

import (
	"errors"
	"encoding/json"
	"fmt"
	"os"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type update struct {
	util.ProcessIO
	pecoable
	uploadable
	WithoutBody bool `short:"B" long:"nobody" description:"Exec without body."`
	KeepLockRequired bool `short:"L" long:"keeplock" description:"Exec without unlock."`
}

func init() {
	registProcessor(func() util.Processable { return &update{} }, "update", "Update a post.", "[OPTIONS] POST...")
}

func (self *update) Do(args []string) error {
	if len(args) == 0 { args = self.ScanArgs() }

	if len(args) == 0 && self.PecoRequired() {
		var err error
		args, err = selectNodeByPeco("", false)
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
		OriginalRevision: esa.PostOriginalRevision {
			Number: post.RevisionNumber,
			User: post.UpdatedBy.ScreenName,
		},
	}

	self.setWip(&new_post, post.Wip)
	self.setTags(&new_post, post.Tags)
	self.setCategory(&new_post, post.Category)
	self.setName(&new_post, post.Name)
	self.setMessage(&new_post)
	if err := self.setBody(&new_post, post_number); err != nil { return err }
	
	self.Println("Start upload...")
	res, err := Context.Client.Post.Update(Context.Team, post.Number, new_post)
	if err != nil { return err }
	self.Println("Finished upload.")
	
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
