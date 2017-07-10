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
	log "github.com/sirupsen/logrus"
	"github.com/upamune/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

type ls struct {
	writer io.Writer
	recursive bool
	directory_only bool
}

func init() {
	addProcessor(&ls{ writer: os.Stdout }, "ls", "Print a list of category and post information.")
}

func (self *ls) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.recursive, "r", false, "Recursive.")
	flagset.BoolVar(&self.directory_only, "d", false, "Directory only.")
}

func (self *ls) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }
	
	self.printNodesIn(path, AbsolutePathOf(path))
	return nil
}

func (self *ls) printNodesIn(path string, abs_path string) {
	writer := bufio.NewWriter(self.writer)
	
	for _, node := range util.GetNodes(abs_path) {
		node_path		:= filepath.Join(path, node.Name())
		node_abs_path	:= filepath.Join(abs_path, node.Name())
		
		if node.IsDir() {
			fmt.Fprintln(writer, node_path)
			writer.Flush()

			if self.recursive { self.printNodesIn(node_path, node_abs_path) }
		} else if ! self.directory_only {
			node_name_parts := strings.Split(node.Name(), ".")
			
			if len(node_name_parts) == 2 {
				post_number := node_name_parts[0]
				post_data_type := node_name_parts[1]
				if post_data_type == "json" {
					bytes := LoadPostData(abs_path, post_number, "json")

					var post esa.PostResponse
					if err := json.Unmarshal(bytes, &post); err != nil {
						log.WithFields(log.Fields{ "name": node.Name(), "path": node_abs_path }).Error("Failed to load path")
					}

					fmt.Fprintf(writer, "%s: %s\n", filepath.Join(path, post_number), post.Name)
				}
			} else {
				log.WithFields(log.Fields{ "name": node.Name(), "path": node_abs_path }).Error("Unknown node")
			}
		}
	}

	writer.Flush()
}
