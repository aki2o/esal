package action

import (
	"flag"
	"errors"
	"os"
	"os/exec"
	log "github.com/sirupsen/logrus"
)

type open struct {}

func init() {
	addProcessor(&open{}, "open", "Open a post.")
}

func (self *open) SetOption(flagset *flag.FlagSet) {
}

func (self *open) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }
	
	if path == "" {
		return errors.New("Require path!")
	}

	real_path := AbsolutePathOf(path)+".md"
	before_file_info, err := os.Stat(real_path)
	if err != nil { return err }

	lock_process := &lock{}
	if err := lock_process.Do([]string{ path }); err != nil { return err }
	
	cmd := exec.Command(os.Getenv("EDITOR"), real_path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil { return err }
	
	after_file_info, err := os.Stat(real_path)
	if err != nil { return err }

	if after_file_info.ModTime().After(before_file_info.ModTime()) {
		log.WithFields(log.Fields{ "path": path }).Info("Start upload after open.")
		upload_process := &upload{}
		if err := upload_process.Do([]string{ path }); err != nil { return err }
	}

	unlock_process := &unlock{}
	if err := unlock_process.Do([]string{ path }); err != nil { return err }
	
	return nil
}
