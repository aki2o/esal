package action

import (
	"fmt"
	"path/filepath"
	"strings"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type find struct {
	util.ProcessIO
	pecoable
	matchable
}

func init() {
	RegistProcessor(func() util.Processable { return &find{} }, "find", "Print categories and posts matches condition under CATEGORY.", "[OPTIONS] CATEGORY")
}

func (self *find) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if path == "" {
		return errors.New("Require category!")
	}
	
	found_paths, _, err := self.collectNodesIn(path)
	if err != nil { return err }
	
	self.Println(strings.Join(found_paths, "\n"))
	return nil
}

func (self *find) collectNodesIn(path string) ([]string, []*esa.PostResponse, error) {
	found_paths := []string{}
	found_posts := []*esa.PostResponse{}
	physical_path := PhysicalPathOf(path)
	path = DirectoryFormat(path)

	for _, node := range util.GetNodes(physical_path) {
		node_physical_path := filepath.Join(physical_path, node.Name())

		if node.IsDir() {
			decoded_name := util.DecodePath(node.Name())
			node_path := path+decoded_name

			if self.matchCategory(decoded_name) {
				found_paths = append(found_paths, node_path+"/")
			}

			found_child_paths, found_child_posts, err := self.collectNodesIn(node_path)
			if err != nil { return []string{}, []*esa.PostResponse{}, err }
			
			found_paths = append(found_paths, found_child_paths...)
			found_posts = append(found_posts, found_child_posts...)
		} else {
			var post esa.PostResponse

			post_number := node.Name()
			bytes, err := LoadPostData(post_number)
			if err == nil { err = json.Unmarshal(bytes, &post) }

			if err != nil {
				log.WithFields(log.Fields{ "name": node.Name(), "path": node_physical_path }).Error("Failed to load post")
				util.PutError(errors.New("Failed to load post data of "+post_number+"!"))
				continue
			}

			if self.matchPost(&post) {
				found_paths = append(found_paths, fmt.Sprintf("%s%s: %s", path, post_number, post.Name))
				found_posts = append(found_posts, &post)
			}
		}
	}

	return found_paths, found_posts, nil
}
