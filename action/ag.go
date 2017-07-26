package action

import (
	"os/exec"
	"fmt"
	"regexp"
	"io"
	"bufio"
	"strings"
	"errors"
	"encoding/json"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type ag struct {
	pecoable
	ByBrowser bool `short:"b" long:"browser" description:"Open peco results by browser."`
	EditRequired bool `short:"e" long:"edit" description:"Open peco results for edit."`
}

func init() {
	registProcessor(func() util.Processable { return &ag{} }, "ag", "Execute ag command.", "[OPTIONS] PATTERN")
}

func (self *ag) Do(args []string) error {
	cmd_args := append([]string{"-G", ".md$"}, args...)
	cmd_args = append(cmd_args, Context.BodyRoot())

	if self.PecoRequired() {
		return self.processByPeco(cmd_args)
	} else {
		return self.process(cmd_args)
	}
}

func (self *ag) process(cmd_args []string) error {
	out, err := exec.Command("ag", cmd_args...).Output()
	if err != nil { return err }
	
	fmt.Print(self.pathRegexp().ReplaceAllStringFunc(string(out), self.appendPostName))
	return nil
}

func (self *ag) processByPeco(cmd_args []string) error {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()

		out, err := exec.Command("ag", cmd_args...).Output()
		if err != nil { return }

		rich_writer := bufio.NewWriter(writer)

		fmt.Fprintf(rich_writer, self.pathRegexp().ReplaceAllStringFunc(string(out), self.appendPostName))
		rich_writer.Flush()
	}

	selected, _, err := pipePeco(provider, "Query")
	if err != nil { return err }

	re, _ := regexp.Compile("^([0-9]+):[0-9]+:")
	for _, line := range strings.Split(selected, "\n") {
		if len(line) == 0 { continue }
		
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			args := []string{ matches[1] }
			
			if self.EditRequired {
				open_process := &open{ ByBrowser: self.ByBrowser }
				open_process.NoPecolize = true
				open_process.Do(args)
			} else {
				cat_process := &cat{ ByBrowser: self.ByBrowser }
				cat_process.NoPecolize = true
				cat_process.Do(args)
			}
		} else {
			return errors.New("Can't find post from result!")
		}
	}

	return nil
}

func (self *ag) pathRegexp() *regexp.Regexp {
	re, _ := regexp.Compile("(?m)^"+Context.BodyRoot()+"/[0-9]+\\.md:[0-9]+:")
	return re
}

func (self *ag) appendPostName(path string) string {
	re, _ := regexp.Compile("([0-9]+)\\.md:([0-9]+):$")
	matches := re.FindStringSubmatch(path)
	if len(matches) <= 2 { return path }

	bytes, err := LoadPostData(matches[1])
	if err != nil { return path }

	var post esa.PostResponse
	if err := json.Unmarshal(bytes, &post); err != nil { return path }

	return fmt.Sprintf("%s:%s:%s: ", matches[1], matches[2], post.FullName)
}
