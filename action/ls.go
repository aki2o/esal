package action

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	log "github.com/sirupsen/logrus"
)

type ls struct {}

func init() {
	processors["ls"] = &ls{}
}

func (self *ls) SetOption(flagset *flag.FlagSet) {
}

func (self *ls) Do(args []string) error {
	for _, node := range GetNodes(Context.Cwd) {
		node_path := filepath.Join(Context.Cwd, node.Name())
		
		if node.IsDir() {
			fmt.Println(node.Name()+"/")
		} else if filepath.Ext(node_path) == ".md" {
			node_name_parts := strings.Split(node.Name(), ".")
			post_number		:= node_name_parts[len(node_name_parts) - 2]
			
			fmt.Printf("%s: %s\n", post_number, strings.Join(node_name_parts[0:len(node_name_parts) - 2], "."))
		} else if filepath.Ext(node_path) == ".json" {
		} else {
			log.WithFields(log.Fields{ "name": node.Name(), "path": node_path }).Error("Unknown node")
		}
	}
	
	return nil
}
