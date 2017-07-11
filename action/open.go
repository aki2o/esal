package action

import (
	"flag"
	"errors"
	"os"
	"os/exec"
	"io"
	log "github.com/sirupsen/logrus"
)

type open struct {
	pecolize bool
	recursive bool
}

func init() {
	addProcessor(&open{}, "open", "Open a post.")
}

func (self *open) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
	flagset.BoolVar(&self.recursive, "r", false, "Recursively for peco.")
}

func (self *open) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }
	
	if self.pecolize {
		next_path, err := self.runPeco(path)
		if err != nil { return err }

		path = next_path
	}

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

func (self *open) runPeco(path string) (string, error) {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()
		
		ls := &ls{ writer: writer, recursive: self.recursive, file_only: true }
		ls.printNodesIn(path, AbsolutePathOf(path))
	}

	return pipePeco(provider)
}
