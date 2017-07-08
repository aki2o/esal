package action

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/upamune/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

type ls struct {}

func init() {
	processors["ls"] = &ls{}
}

func (self *ls) SetOption(flagset *flag.FlagSet) {
}

func (self *ls) Do(args []string) error {
	for _, node := range util.GetNodes(Context.Cwd) {
		node_path := filepath.Join(Context.Cwd, node.Name())
		
		if node.IsDir() {
			fmt.Println(node.Name()+"/")
		} else {
			node_name_parts := strings.Split(node.Name(), ".")
			if len(node_name_parts) == 2 {
				post_number := node_name_parts[0]
				post_data_type := node_name_parts[1]
				if post_data_type == "json" {
					bytes := LoadPostData(Context.Cwd, post_number, "json")

					var post esa.PostResponse
					if err := json.Unmarshal(bytes, &post); err != nil { return err }

					fmt.Printf("%s: %s\n", post_number, post.Name)
				}
			} else {
				log.WithFields(log.Fields{ "name": node.Name(), "path": node_path }).Error("Unknown node")
			}
		}
	}
	
	return nil
}
