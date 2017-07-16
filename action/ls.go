package action

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"encoding/json"
	"io"
	"bufio"
	"os"
	"time"
	"strconv"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

type ls struct {
	writer io.Writer
	long_format bool
	recursive bool
	directory_only bool
	file_only bool
}

func init() {
	addProcessor(&ls{ writer: os.Stdout }, "ls", "Print a list of category and post information.")
}

func (self *ls) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.long_format, "l", false, "Print long format.")
	flagset.BoolVar(&self.recursive, "r", false, "Recursively.")
	flagset.BoolVar(&self.directory_only, "d", false, "Directory only.")
	flagset.BoolVar(&self.file_only, "f", false, "File only.")
}

func (self *ls) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }
	
	self.printNodesIn(path, PhysicalPathOf(path))
	return nil
}

func (self *ls) printNodesIn(path string, abs_path string) {
	writer := bufio.NewWriter(self.writer)
	
	for _, node := range util.GetNodes(abs_path) {
		node_path		:= filepath.Join(path, node.Name())
		node_abs_path	:= filepath.Join(abs_path, node.Name())
		
		if node.IsDir() {
			if ! self.file_only {
				fmt.Fprintln(writer, self.makeDirLine(node_path))
				writer.Flush()
			}

			if self.recursive { self.printNodesIn(node_path, node_abs_path) }
		} else if ! self.directory_only {
			var post esa.PostResponse
			
			post_number := node.Name()
			bytes, err := LoadPostData(path, post_number)
			
			if err == nil { err = json.Unmarshal(bytes, &post) }

			if err != nil {
				log.WithFields(log.Fields{ "name": node.Name(), "path": node_abs_path }).Error("Failed to load post")
				util.PutError(errors.New("Failed to load post data of "+post_number+"!"))
			} else {
				fmt.Fprintln(writer, self.makeFileLine(path, &post))
			}
		}
	}

	writer.Flush()
}

func (self *ls) makeDirLine(path string) string {
	return self.makePostStatPart(path, nil)+path
}

func (self *ls) makeFileLine(path string, post *esa.PostResponse) string {
	post_number := strconv.Itoa(post.Number)

	var name_part string
	if self.long_format {
		var wip string = ""
		var lock string = ""
		var tag string = ""
		
		if post.Wip { wip = " [WIP]" }
		if _, err := os.Stat(GetPostLockPath(post_number)); err == nil { lock = " *Lock*" }
		if len(post.Tags) > 0 { tag = " #"+strings.Join(post.Tags, " #") }
		
		name_part = fmt.Sprintf("%s:%s%s %s%s", filepath.Join(path, post_number), wip, lock, post.Name, tag)
	} else {
		name_part = fmt.Sprintf("%s: %s", filepath.Join(path, post_number), post.Name)
	}

	return self.makePostStatPart(path, post)+name_part
}

func (self *ls) makePostStatPart(path string, post *esa.PostResponse) string {
	if !self.long_format { return "" }

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
