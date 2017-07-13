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
	"strconv"
	"regexp"
	"reflect"
	"context"
	"golang.org/x/exp/utf8string"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
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

func GetLocalPostPath(category string, number_as_string string, extension string) string {
	return fmt.Sprintf("%s/%s/%s.%s", Context.Root(), category, number_as_string, extension)
}

func GetLocalPostFileName(number_as_string string, extension string) string {
	return fmt.Sprintf("%s.%s", number_as_string, extension)
}

func StorePostData(category string, number int, extension string, body string) {
	fp, err := os.Create(GetLocalPostPath(category, strconv.Itoa(number), extension))
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

func ExcludePostName(path string) string {
	number_re, _ := regexp.Compile("(/|^)([0-9]+):[^/]+$")
	matches	:= number_re.FindStringSubmatch(path)

	if len(matches) > 2 {
		return number_re.ReplaceAllString(path, matches[1])+matches[2]
	} else {
		return path
	}
}

func AbsolutePathOf(path string) string {
	path = ExcludePostName(path)
	
	if path == "" {
		return Context.Cwd
	} else if utf8string.NewString(path).Slice(0, 1) == "/" {
		return filepath.Join(Context.Root(), path)
	} else {
		return filepath.Join(Context.Cwd, path)
	}
}

func DirectoryPathAndPostNumberOf(path string) (string, string) {
	path = ExcludePostName(path)
	
	re, _	:= regexp.Compile("/?([0-9]*)$")
	matches := re.FindStringSubmatch(path)
	
	var post_number string = ""
	if len(matches) > 1 { post_number = matches[1] }
	
	return re.ReplaceAllString(path, ""), post_number
}

func pipePeco(provider func(*io.PipeWriter)) (string, error) {
	from_provider_reader, to_peco_writer := io.Pipe()
	
	go provider(to_peco_writer)
	
	from_peco_reader, to_self_writer := io.Pipe()

	go func() {
		defer to_self_writer.Close()
		
		peco := peco.New()
		peco.Argv	= []string{"--on-cancel", "error"}
		peco.Stdin	= from_provider_reader
		peco.Stdout = to_self_writer
		
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := peco.Run(ctx); err != nil {
			// peco の終了を判断する機能が公開されていないので、 reflect を使って、無理矢理実装
			err_type := reflect.ValueOf(err)
			switch fmt.Sprintf("%s", err_type.Type()) {
			case "peco.errCollectResults":
				peco.PrintResults()
				return
			case "*peco.errWithExitStatus":
				return
			default:
				log.Info(fmt.Sprintf("Peco return %s", err_type.Type()))
				return
			}
		}
	}()
	
	bytes, err := ioutil.ReadAll(from_peco_reader)
	if err != nil { return "", err }

	return strings.TrimRight(string(bytes), "\n"), nil
}
