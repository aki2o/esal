package action

import (
	"fmt"
	"path/filepath"
	"strings"
	"encoding/json"
	"os"
	"time"
	"strconv"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type ls struct {
	util.ProcessIO
	LongFormatRequired bool `short:"l" long:"long" description:"Print long format."`
	Recursive bool `short:"r" long:"recursive" description:"Exec recursively."`
	CategoryOnly bool `short:"c" long:"category" description:"Print only category."`
	PostOnly bool `short:"p" long:"post" description:"Print only post."`
}

func init() {
	RegistProcessor(func() util.Processable { return &ls{} }, "ls", "Print a list of category and post information.", "[OPTIONS]")
}

func (self *ls) Do(args []string) error {
	if len(args) == 0 { args = self.ScanArgs() }

	var path string = ""
	if len(args) > 0 { path = args[0] }
	
	self.printNodesIn(path, PhysicalPathOf(path))
	return nil
}

func (self *ls) printNodesIn(path string, physical_path string) {
	path = DirectoryFormat(path)
	for _, node := range util.GetNodes(physical_path) {
		node_physical_path := filepath.Join(physical_path, node.Name())
		
		if node.IsDir() {
			node_path := path+util.DecodePath(node.Name())
			
			if ! self.PostOnly {
				self.Println(self.makeDirLine(node_path))
			}

			if self.Recursive { self.printNodesIn(node_path, node_physical_path) }
		} else if ! self.CategoryOnly {
			var post esa.PostResponse
			
			post_number := node.Name()
			bytes, err := LoadPostData(post_number)
			
			if err == nil { err = json.Unmarshal(bytes, &post) }

			if err != nil {
				log.WithFields(log.Fields{ "name": node.Name(), "path": node_physical_path }).Error("Failed to load post")
				util.PutError(errors.New("Failed to load post data of "+post_number+"!"))
			} else {
				self.Println(self.makeFileLine(path, &post))
			}
		}
	}
}

func (self *ls) makeDirLine(node_path string) string {
	return self.makePostStatPart(nil)+node_path+"/"
}

func (self *ls) makeFileLine(path string, post *esa.PostResponse) string {
	post_number := strconv.Itoa(post.Number)

	var name_part string
	if self.LongFormatRequired {
		var wip string = ""
		var lock string = ""
		var tag string = ""
		
		if post.Wip { wip = " [WIP]" }
		if _, err := os.Stat(GetPostLockPath(post_number)); err == nil { lock = " *Lock*" }
		if len(post.Tags) > 0 { tag = " #"+strings.Join(post.Tags, " #") }
		
		name_part = fmt.Sprintf("%s:%s%s %s%s", path+post_number, wip, lock, post.Name, tag)
	} else {
		name_part = fmt.Sprintf("%s: %s", path+post_number, post.Name)
	}

	return self.makePostStatPart(post)+name_part
}

func (self *ls) makePostStatPart(post *esa.PostResponse) string {
	if !self.LongFormatRequired { return "" }

	var create_user		string = ""
	var update_user		string = ""
	var post_size		string = ""
	var last_updated_at string = ""

	if post != nil {
		create_user = post.CreatedBy.ScreenName
		update_user = post.UpdatedBy.ScreenName

		file_info, err := os.Stat(GetPostBodyPath(strconv.Itoa(post.Number)))
		if err == nil {
			post_size = fmt.Sprintf("%d", file_info.Size())
		} else {
			post_size = "?"
		}
		
		updated_at, err := time.Parse("2006-01-02T15:04:05-07:00", post.UpdatedAt)
		if err != nil {
			last_updated_at = "** ** **:**"
		} else if updated_at.Year() == time.Now().Year() {
			last_updated_at = updated_at.Format("01 02 15:04")
		} else {
			last_updated_at = updated_at.Format("01 02  2006")
		}
	}
	
	return fmt.Sprintf(
		"%s %s %s %s ",
		fmt.Sprintf("%20s", create_user),
		fmt.Sprintf("%20s", update_user),
		fmt.Sprintf("%10s", post_size),
		fmt.Sprintf("%11s", last_updated_at),
	)
}
