package action

import (
	"errors"
	"os"
	"os/exec"
	"encoding/json"
	"runtime"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

type open struct {
	*pecoable
	Editor string `short:"e" long:"editor" description:"Open with Editor." value-name:"EDITOR"`
}

func init() {
	registProcessor(func() util.Processable { return &open{} }, "open", "Open a post.", "[OPTIONS] POST")
}

func (self *open) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }
	
	if self.PecoRequired() {
		next_path, err := selectNodeByPeco(path, false)
		if err != nil { return err }

		path = next_path
	}

	_, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	editor := self.Editor
	if editor == "" { editor = os.Getenv("EDITOR") }

	if editor != "" {
		return self.openByEditor(path, post_number, editor)
	} else {
		return self.openByBrowser(post_number)
	}
}

func (self *open) openByEditor(path string, post_number string, editor string) error {
	real_path := GetPostBodyPath(post_number)
	
	before_file_info, err := os.Stat(real_path)
	if err != nil { return err }

	lock_process := &lock{}
	if err := lock_process.Do([]string{ path }); err != nil { return err }
	
	cmd := exec.Command(editor, real_path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil { return err }
	
	after_file_info, err := os.Stat(real_path)
	if err != nil { return err }

	if after_file_info.ModTime().After(before_file_info.ModTime()) {
		log.WithFields(log.Fields{ "path": path }).Info("Start update after open.")
		update_process := &update{}
		if err := update_process.Do([]string{ path }); err != nil { return err }
	} else {
		unlock_process := &unlock{}
		if err := unlock_process.Do([]string{ path }); err != nil { return err }
	}

	return nil
}

func (self *open) openByBrowser(post_number string) error {
	bytes, err := LoadPostData(post_number)
	if err != nil { return err }

	var post esa.PostResponse
	if err := json.Unmarshal(bytes, &post); err != nil { return err }

	cmd := ""
	if runtime.GOOS == "windows" {
		cmd = "start"
	} else if runtime.GOOS == "darwin" {
		cmd = "open"
	} else {
		cmd = "xdg-open"
	}

	if err := exec.Command(cmd, post.URL+"/edit").Run(); err != nil { return err }

	return nil
}
