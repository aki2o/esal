package action

import (
	"flag"
	"errors"
	"encoding/json"
	"fmt"
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
	nobody bool
	lock_keeping bool
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
	flagset.BoolVar(&self.nobody, "nobody", false, "Exec without body.")
	flagset.BoolVar(&self.lock_keeping, "keeplock", false, "Exec without unlock.")
}

func (self *update) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if self.pecolize {
		next_path, err := selectNodeByPeco(path, false)
		if err != nil { return err }

		path = next_path
	}

	dir_path, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	post, err := self.loadPostDataWithVerify(dir_path, post_number)
	if err != nil { return err }
	
	new_post := esa.Post{
		Message: self.message,
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
	
	if ! self.lock_keeping {
		unlock_process := &unlock{}
		if err := unlock_process.Do([]string{ path }); err != nil { return err }
	}
	
	return nil
}

func (self *update) loadPostDataWithVerify(dir_path string, post_number string) (*esa.PostResponse, error) {
	json_bytes, err := LoadPostData(dir_path, post_number)
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
	
	if self.wip { wip = true }
	if self.ship { wip = false }

	new_post.Wip = wip
}

func (self *update) setTags(new_post *esa.Post, post *esa.PostResponse) {
	tags := post.Tags
	
	if self.tags != "" { tags = strings.Split(self.tags, ",") }

	new_post.Tags = tags
}

func (self *update) setCategory(new_post *esa.Post, post *esa.PostResponse) {
	category := post.Category
	
	if self.category != "" {
		re, _ := regexp.Compile("^/")
		category = re.ReplaceAllString(self.category, "")
	}

	new_post.Category = category
}

func (self *update) setName(new_post *esa.Post, post *esa.PostResponse) {
	post_name := post.Name
	
	if self.post_name != "" { post_name = self.post_name }

	new_post.Name = post_name
}

func (self *update) setBody(new_post *esa.Post, post_number string) error {
	body_bytes, err := LoadPostBody(post_number)
	if err != nil { return err }
	
	lock_bytes, err := LoadPostLock(post_number)
	if err != nil { lock_bytes = body_bytes }

	if self.nobody {
		new_post.BodyMd = string(lock_bytes)
	} else {
		new_post.BodyMd = string(body_bytes)
	}
	new_post.OriginalRevision.BodyMd = string(lock_bytes)
	return nil
}
