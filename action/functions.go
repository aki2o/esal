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
	"runtime"
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

func DeletePostData(number_as_string string) error {
	return os.Remove(GetPostDataPath(number_as_string))
}

func DeletePostBody(number_as_string string) error {
	return os.Remove(GetPostBodyPath(number_as_string))
}

func ExcludePostName(path string) string {
	re, _ := regexp.Compile("(/|^)([0-9]+):[^/]+$")
	matches	:= re.FindStringSubmatch(path)

	if len(matches) > 2 {
		return re.ReplaceAllString(path, matches[1])+matches[2]
	} else {
		return path
	}
}

func DirectoryFormat(category string) string {
	if category == "" { return category }
	if category == "/" { return category }

	re, _ := regexp.Compile("/$")
	return re.ReplaceAllString(category, "")+"/"
}

func ParentOf(category string) string {
	re, _ := regexp.Compile("/[^/]+/?$")
	ret := re.ReplaceAllString(category, "")

	if ret == "" {
		return "/"
	} else {
		return ret
	}
}

func CategoryOf(physical_path string) string {
	separator := string(os.PathSeparator)
	root_dirs := strings.Split(Context.Root(), separator)
	curr_dirs := strings.Split(util.DecodePath(physical_path), separator)[len(root_dirs):]

	return strings.Join(curr_dirs, "/")
}

func PhysicalPathOf(path string) string {
	path = ExcludePostName(path)
	if path == "" {	return Context.Cwd }
	
	categories_or_number := strings.Split(path, "/")
	physical_path := util.EncodePath(filepath.Join(categories_or_number...))
	
	if categories_or_number[0] == "" {
		return filepath.Join(Context.Root(), physical_path)
	} else {
		return filepath.Join(Context.Cwd, physical_path)
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

func FindPostDataPath(physical_path string, number_as_string string) []string {
	ret := []string{}
	
	for _, node := range util.GetNodes(physical_path) {
		if node.IsDir() {
			ret = append(ret, FindPostDataPath(filepath.Join(physical_path, node.Name()), number_as_string)...)
		} else if node.Name() == number_as_string {
			ret = append(ret, filepath.Join(physical_path, number_as_string))
		}
	}

	return ret
}

func BrowserCommand() string {
	if runtime.GOOS == "windows" {
		return "start"
	} else if runtime.GOOS == "darwin" {
		return "open"
	} else {
		return "xdg-open"
	}
}
