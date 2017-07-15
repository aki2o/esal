package action

import (
	"fmt"
	"io"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"strings"
	"strconv"
	"regexp"
	"reflect"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/peco/peco"
	"github.com/aki2o/esa-cui/util"
)

func GetPostDataPath(category string, number_as_string string) string {
	return filepath.Join(AbsolutePathOf(category), number_as_string)
}

func GetPostBodyPath(number_as_string string) string {
	return filepath.Join(Context.BodyRoot(), fmt.Sprintf("%s.md", number_as_string))
}

func GetPostLockPath(number_as_string string) string {
	return filepath.Join(Context.BodyRoot(), fmt.Sprintf("%s.lock", number_as_string))
}

func SavePost(post *esa.PostResponse) error {
	log.WithFields(log.Fields{ "path": post.FullName }).Debug("start to save post")

	util.EnsureDir(AbsolutePathOf("/"+post.Category))
	err := util.CreateFile(GetPostBodyPath(strconv.Itoa(post.Number)), post.BodyMd)
	if err != nil { return err }

	post.BodyMd = ""
	post.BodyHTML = ""
	
	post_json_data, err := json.MarshalIndent(post, "", "\t")
	if err != nil { return err }
	
	err = util.CreateFile(GetPostDataPath(post.Category, strconv.Itoa(post.Number)), string(post_json_data))
	if err != nil { return err }

	return nil
}

func LoadPostData(category string, number_as_string string) ([]byte, error) {
	return ioutil.ReadFile(GetPostDataPath(category, number_as_string))
}

func LoadPostBody(number_as_string string) ([]byte, error) {
	return ioutil.ReadFile(GetPostBodyPath(number_as_string))
}

func LoadPostLock(number_as_string string) ([]byte, error) {
	return ioutil.ReadFile(GetPostLockPath(number_as_string))
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
	if path == "" {	return Context.Cwd }
	
	dir_names := strings.Split(path, "/")
	
	if dir_names[0] == "" {
		return filepath.Join(Context.Root(), filepath.Join(dir_names...))
	} else {
		return filepath.Join(Context.Cwd, filepath.Join(dir_names...))
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
