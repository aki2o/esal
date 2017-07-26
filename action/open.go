package action

import (
	"errors"
	"os"
	"os/exec"
	"encoding/json"
	"runtime"
	"io/ioutil"
	"net/url"
	"fmt"
	"bufio"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type open struct {
	*pecoable
	Editor string `short:"e" long:"editor" description:"Open by editor." value-name:"EDITOR"`
	Browser bool `short:"b" long:"browser" description:"Open by browser."`
	NewPost bool `short:"n" long:"new" description:"Open new post."`
}

func init() {
	registProcessor(func() util.Processable { return &open{} }, "open", "Open a post.", "[OPTIONS] POST")
}

func (self *open) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }
	
	if ! self.NewPost && self.PecoRequired() {
		next_path, err := selectNodeByPeco(path, false)
		if err != nil { return err }

		path = next_path
	}

	dir_path, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" && ! self.NewPost {
		return errors.New("Require post number!")
	}
	if post_number != "" && self.NewPost {
		return errors.New("Forbid post number with new option!")
	}

	editor := self.Editor
	if editor == "" { editor = os.Getenv("EDITOR") }

	if self.Browser {
		return self.openByBrowser(dir_path, post_number)
	} else if editor != "" {
		return self.openByEditor(path, dir_path, post_number, editor)
	} else {
		return self.openByBrowser(dir_path, post_number)
	}
}

func (self *open) openByEditor(path string, dir_path string, post_number string, editor string) error {
	real_path, err := self.getBufferPath(post_number)
	if err != nil { return err }
	
	before_file_info, err := os.Stat(real_path)
	if err != nil { return err }

	if ! self.NewPost {
		lock_process := &lock{}
		if err := lock_process.Do([]string{ path }); err != nil { return err }
	}
	
	cmd := exec.Command(editor, real_path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil { return err }
	
	after_file_info, err := os.Stat(real_path)
	if err != nil { return err }

	if after_file_info.ModTime().After(before_file_info.ModTime()) {
		if self.NewPost {
			log.WithFields(log.Fields{ "path": path }).Info("Start regist after open.")
			post_name := self.scanString("Post Name: ")
			
			category, err := selectNodeByPeco("/", true)
			if err != nil { category = CategoryOf(Context.Cwd) }
			
			regist_process := &regist{
				PostName: post_name,
				Category: category,
				Message: self.scanString("Commit Message: "),
			}
			if err := regist_process.Do([]string{ real_path }); err != nil { return err }
		} else {
			log.WithFields(log.Fields{ "path": path }).Info("Start update after open.")
			update_process := &update{
				Message: self.scanString("Commit Message: "),
			}
			if err := update_process.Do([]string{ path }); err != nil { return err }
		}
	}

	if ! self.NewPost {
		unlock_process := &unlock{}
		if err := unlock_process.Do([]string{ path }); err != nil { return err }
	}
	
	return nil
}

func (self *open) openByBrowser(dir_path string, post_number string) error {
	url, err := self.getURL(dir_path, post_number)
	if err != nil { return err }
	
	cmd := ""
	if runtime.GOOS == "windows" {
		cmd = "start"
	} else if runtime.GOOS == "darwin" {
		cmd = "open"
	} else {
		cmd = "xdg-open"
	}

	if err := exec.Command(cmd, url).Run(); err != nil { return err }

	return nil
}

func (self *open) getBufferPath(post_number string) (string, error) {
	if self.NewPost {
		fp, err := ioutil.TempFile("", Context.Team)
		if err != nil { return "", err }

		file_path := fp.Name()
		fp.Close()

		if err := os.Rename(file_path, file_path+".md"); err != nil { return "", err }
		
		return file_path+".md", nil
	} else {
		return GetPostBodyPath(post_number), nil
	}
}

func (self *open) getURL(dir_path string, post_number string) (string, error) {
	if self.NewPost {
		v := url.Values{}
		v.Add("category_path", "/"+CategoryOf(PhysicalPathOf(dir_path))+"/")
		
		return fmt.Sprintf("https://%s.esa.io/posts/new?%s", Context.Team, v.Encode()), nil
	} else {
		bytes, err := LoadPostData(post_number)
		if err != nil { return "", err }

		var post esa.PostResponse
		if err := json.Unmarshal(bytes, &post); err != nil { return "", err }

		return post.URL+"/edit", nil
	}
}

func (self *open) scanString(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
