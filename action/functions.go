package action

import (
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"strings"
	"strconv"
	"regexp"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

func GetCategoryPostPath(category string, number_as_string string) string {
	return filepath.Join(PhysicalPathOf(category), number_as_string)
}

func GetPostBodyPath(number_as_string string) string {
	return filepath.Join(Context.BodyRoot(), fmt.Sprintf("%s.md", number_as_string))
}

func GetPostDataPath(number_as_string string) string {
	return filepath.Join(Context.BodyRoot(), fmt.Sprintf("%s.json", number_as_string))
}

func GetPostLockPath(number_as_string string) string {
	return filepath.Join(Context.BodyRoot(), fmt.Sprintf("%s.lock", number_as_string))
}

func SavePost(post *esa.PostResponse) error {
	log.WithFields(log.Fields{ "path": post.FullName }).Debug("start to save post")

	post_number := strconv.Itoa(post.Number)
	
	for _, file_path := range FindPostDataPath(Context.Root(), post_number) {
		if err := os.Remove(file_path); err != nil { return err }
	}
	
	err := util.EnsureDir(Context.BodyRoot())
	if err != nil { return err }
	err = util.CreateFile(GetPostBodyPath(post_number), post.BodyMd)
	if err != nil { return err }

	post.BodyMd = ""
	post.BodyHTML = ""

	post_json_data, err := json.MarshalIndent(post, "", "\t")
	if err != nil { return err }
	err = util.CreateFile(GetPostDataPath(post_number), string(post_json_data))
	if err != nil { return err }
	
	err = util.EnsureDir(PhysicalPathOf("/"+post.Category))
	if err != nil { return err }
	err = util.CreateFile(GetCategoryPostPath("/"+post.Category, post_number), "")
	if err != nil { return err }

	return nil
}

func LoadPostData(number_as_string string) ([]byte, error) {
	return ioutil.ReadFile(GetPostDataPath(number_as_string))
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

func CategoryOf(physical_path string) string {
	separator := string(os.PathSeparator)
	root_dirs := strings.Split(Context.Root(), separator)
	curr_dirs := strings.Split(physical_path, separator)[len(root_dirs):]

	return strings.Join(curr_dirs, "/")
}

func PhysicalPathOf(path string) string {
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

func FindPostDataPath(abs_path string, number_as_string string) []string {
	ret := []string{}
	
	for _, node := range util.GetNodes(abs_path) {
		if node.IsDir() {
			ret = append(ret, FindPostDataPath(filepath.Join(abs_path, node.Name()), number_as_string)...)
		} else if node.Name() == number_as_string {
			ret = append(ret, filepath.Join(abs_path, number_as_string))
		}
	}

	return ret
}
