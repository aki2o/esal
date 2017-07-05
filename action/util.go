package action

import (
	"os"
	"fmt"
	"io/ioutil"
	"bufio"
	"encoding/json"
	"path/filepath"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/upamune/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

func SavePost(post *esa.PostResponse) {
	log.WithFields(log.Fields{ "path": post.FullName }).Debug("start to save post")
	
	util.EnsureDir(Context.Root()+"/"+post.Category)
	StorePostData(post.FullName, post.Number, "md", post.BodyMd)

	post.BodyMd = ""
	post.BodyHTML = ""
	
	post_json_data, err := json.MarshalIndent(post, "", "\t")
	if err != nil {
		util.PutError(err)
		return
	}
	StorePostData(post.FullName, post.Number, "json", string(post_json_data))
}

func GetLocalPostPath(full_name string, number int, extension string) string {
	return fmt.Sprintf("%s/%s.%d.%s", Context.Root(), full_name, number, extension)
}

func StorePostData(full_name string, number int, extension string, body string) {
	fp, err := os.Create(GetLocalPostPath(full_name, number, extension))
	if err != nil { panic(err) }
	defer fp.Close()
	writer := bufio.NewWriter(fp)
	_, err = writer.WriteString(body)
	if err != nil { panic(err) }
	writer.Flush()
}

func LoadPostData(path string, number string, extension string) []byte {
	for _, node := range util.GetNodes(path) {
		if node.IsDir() { continue }
		
		node_path			:= filepath.Join(path, node.Name())
		node_name_parts		:= strings.Split(node.Name(), ".")
		curr_post_number	:= node_name_parts[len(node_name_parts) - 2]
		
		if filepath.Ext(node_path) != "."+extension { continue }
		if curr_post_number != number { continue }

		bytes, err := ioutil.ReadFile(node_path)
		if err != nil { panic(err) }

		return bytes
	}

	return []byte{}
}
