package action

import (
	"fmt"
	"path/filepath"
	"strings"
	"encoding/json"
	"bufio"
	"errors"
	"regexp"
	"os"
	"strconv"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type find struct {
	pecoable
	Type string `short:"t" long:"type" description:"Node type ident. one of c(ategory), p(ost), w(ip), s(hipped), l(ocked), u(nlocked)." value-name:"IDENT"`
	NamePattern string `short:"n" long:"name" description:"Node name pattern." value-name:"PATTERN"`
	name_re *regexp.Regexp
}

func init() {
	registProcessor(func() util.Processable { return &find{} }, "find", "Print categories and posts matches condition under CATEGORY.", "[OPTIONS] CATEGORY")
}

func (self *find) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if path == "" {
		return errors.New("Require category!")
	}
	
	writer := bufio.NewWriter(os.Stdout)

	found_paths, err := self.collectNodesIn(path)
	if err != nil { return err }
	
	fmt.Fprintln(writer, strings.Join(found_paths, "\n"))
	
	writer.Flush()
	return nil
}

func (self *find) collectNodesIn(path string) ([]string, error) {
	founds := []string{}
	physical_path := PhysicalPathOf(path)
	path = DirectoryFormat(path)

	if self.NamePattern != "" {
		re, err := regexp.Compile(self.NamePattern)
		if err != nil { return []string{}, err }
		self.name_re = re
	}
	
	for _, node := range util.GetNodes(physical_path) {
		node_physical_path := filepath.Join(physical_path, node.Name())

		if node.IsDir() {
			decoded_name := util.DecodePath(node.Name())
			node_path := path+decoded_name

			if self.matchesCategory(decoded_name) {
				founds = append(founds, node_path+"/")
			}

			found_paths, err := self.collectNodesIn(node_path)
			if err != nil { return []string{}, err }
			
			founds = append(founds, found_paths...)
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

			if self.matchesPost(&post) {
				founds = append(founds, fmt.Sprintf("%s%s: %s", path, post_number, post.Name))
			}
		}
	}

	return founds, nil
}

func (self *find) matchesCategory(node_name string) bool {
	if self.Type != "" && self.Type != "c" { return false }
	if self.name_re != nil && ! self.name_re.MatchString(node_name) { return false }
	
	return true
}

func (self *find) matchesPost(post *esa.PostResponse) bool {
	if self.Type == "c" { return false }
	if self.Type == "w" && ! post.Wip { return false }
	if self.Type == "s" && post.Wip { return false }
	if self.Type == "l" && ! util.Exists(GetPostLockPath(strconv.Itoa(post.Number))) { return false }
	if self.Type == "u" && util.Exists(GetPostLockPath(strconv.Itoa(post.Number))) { return false }
	if self.name_re != nil && ! self.name_re.MatchString(post.Name) { return false }
	
	return true
}
