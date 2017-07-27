package action

import (
	"regexp"
	"fmt"
	"io"
	"strings"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type uploadable struct {
	Wip bool `short:"w" long:"wip" description:"Update the post as wip."`
	Shipping bool `short:"s" long:"ship" description:"Ship the post."`
	Tags []string `short:"t" long:"tag" description:"Tag name labeling tha post." value-name:"TAG"`
	Category string `short:"c" long:"category" description:"Category of the post." value-name:"CATEGORY"`
	PostName string `short:"n" long:"name" description:"Name of the post." value-name:"NAME"`
	Message string `short:"m" long:"message" description:"Commit message." default:"Update post." value-name:"MESSAGE"`
	TagsByPecoRequired bool `short:"T" long:"tag-by-peco" description:"Choice tags by peco."`
	CategoryByPecoRequired bool `short:"C" long:"category-by-peco" description:"Choice category by peco."`
	MessageByScan bool `short:"M" long:"message-by-scan" description:"Input commit message."`
}

func (self *uploadable) setWip(post *esa.Post, default_value bool) {
	wip := default_value
	
	if self.Wip { wip = true }
	if self.Shipping { wip = false }

	post.Wip = wip
}

func (self *uploadable) setTags(post *esa.Post, default_value []string) {
	tags := default_value
	
	if len(self.Tags) > 0 {
		tags = self.Tags
	} else if self.TagsByPecoRequired {
		provider := func(writer *io.PipeWriter) {
			defer writer.Close()

			tag_process := &tag{ writer: writer, Separator: "\n" }
			tag_process.PrintTags()
		}

		selected, _, err := pipePeco(provider, "Select tag")
		if err == nil {
			tags = strings.Split(selected, "\n")
		}
	}

	post.Tags = tags
}

func (self *uploadable) setCategory(post *esa.Post, default_value string) {
	category := default_value
	re, _ := regexp.Compile("^/")
	
	if self.Category != "" {
		category = re.ReplaceAllString(self.Category, "")
	} else if self.CategoryByPecoRequired {
		categories, err := selectNodeByPeco(CategoryOf(Context.Cwd), true)
		if err == nil {
			child_category := util.ScanString(fmt.Sprintf("Child Category of (%s): ", categories[0]))
			category = re.ReplaceAllString(categories[0], "")+"/"+re.ReplaceAllString(child_category, "")
		}
	}

	post.Category = category
}

func (self *uploadable) setName(post *esa.Post, default_value string) {
	post_name := default_value
	
	if self.PostName != "" { post_name = self.PostName }

	post.Name = post_name
}

func (self *uploadable) setMessage(post *esa.Post) {
	if self.MessageByScan {
		post.Message = util.ScanString("Commit Message: ")
	}

	if post.Message == "" {
		post.Message = self.Message
	}
}
