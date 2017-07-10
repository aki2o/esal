package action

import (
	"os"
	"fmt"
	"io"
	"io/ioutil"
	"bufio"
	"encoding/json"
	"path/filepath"
	"strings"
	"regexp"
	"context"
	"golang.org/x/exp/utf8string"
	log "github.com/sirupsen/logrus"
	"github.com/upamune/go-esa/esa"
	"github.com/peco/peco"
	"github.com/aki2o/esa-cui/util"
)

func SavePost(post *esa.PostResponse) {
	log.WithFields(log.Fields{ "path": post.FullName }).Debug("start to save post")
	
	util.EnsureDir(Context.Root()+"/"+post.Category)
	StorePostData(post.Category, post.Number, "md", post.BodyMd)

	post.BodyMd = ""
	post.BodyHTML = ""
	
	post_json_data, err := json.MarshalIndent(post, "", "\t")
	if err != nil {
		util.PutError(err)
		return
	}
	StorePostData(post.Category, post.Number, "json", string(post_json_data))
}

func GetLocalPostPath(category string, number int, extension string) string {
	return fmt.Sprintf("%s/%s/%d.%s", Context.Root(), category, number, extension)
}

func GetLocalPostFileName(number_as_string string, extension string) string {
	return fmt.Sprintf("%s.%s", number_as_string, extension)
}

func StorePostData(category string, number int, extension string, body string) {
	fp, err := os.Create(GetLocalPostPath(category, number, extension))
	if err != nil { panic(err) }
	defer fp.Close()
	writer := bufio.NewWriter(fp)
	_, err = writer.WriteString(body)
	if err != nil { panic(err) }
	writer.Flush()
}

func LoadPostData(path string, number_as_string string, extension string) []byte {
	file_path := fmt.Sprintf("%s/%s", path, GetLocalPostFileName(number_as_string, extension))
	bytes, err := ioutil.ReadFile(file_path)
	if err != nil { panic(err) }

	return bytes
}

func AbsolutePathOf(path string) string {
	if path == "" {
		return Context.Cwd
	} else if utf8string.NewString(path).Slice(0, 1) == "/" {
		return filepath.Join(Context.Root(), path)
	} else {
		return filepath.Join(Context.Cwd, path)
	}
}

func DirectoryPathAndPostNumberOf(path string) (string, string) {
	re, _ := regexp.Compile("/?([0-9]*)$")
	matches := re.FindStringSubmatch(path)
	
	var post_number string = ""
	if len(matches) > 1 { post_number = matches[1] }
	
	return re.ReplaceAllString(path, ""), post_number
}

func pipePeco(provider func(*io.PipeWriter)) (string, error) {
	from_provider_reader, to_peco_writer := io.Pipe()
	
	go provider(to_peco_writer)
	
	from_peco_reader, to_self_writer := io.Pipe()

	var peco_err error = nil
	go func() {
		defer to_self_writer.Close()
		
		peco := peco.New()
		peco.Argv	= []string{}
		peco.Stdin	= from_provider_reader
		peco.Stdout = to_self_writer

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		if err := peco.Run(ctx); err != nil {
			peco_err = err
		}
		
		peco.PrintResults()
	}()
	if peco_err != nil { return "", peco_err }
	
	bytes, err := ioutil.ReadAll(from_peco_reader)
	if err != nil { return "", err }

	return strings.TrimRight(string(bytes), "\n"), nil
}
