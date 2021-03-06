package action

import (
	osexec "os/exec"
	"fmt"
	"regexp"
	"io"
	"strings"
	"errors"
	"encoding/json"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type ag struct {
	util.ProcessIO
	pecoable
	matchable
	ByBrowser bool `short:"b" long:"browser" description:"Open peco results by browser."`
	EditRequired bool `short:"e" long:"edit" description:"Open peco results for edit."`
	category string
}

func init() {
	RegistProcessor(func() util.Processable { return &ag{} }, "ag", "Execute ag command.", "[OPTIONS] PATTERN [CATEGORY]")
}

func (self *ag) Do(args []string) error {
	if len(args) == 0 { return errors.New("Require pattern!") }
	
	cmd_args := []string{"-i", "-G", ".md$", args[0], Context.BodyRoot()}

	path := ""
	if len(args) > 1 {
		path, _ = DirectoryPathAndPostNumberOf(args[1])
	}
	self.category = CategoryOf(PhysicalPathOf(path))
	
	if self.PecoRequired() {
		return self.processByPeco(cmd_args)
	} else {
		return self.process(cmd_args)
	}
}

func (self *ag) process(cmd_args []string) error {
	out, err := osexec.Command("ag", cmd_args...).Output()
	if err != nil { return err }

	return self.printResult(string(out))
}

func (self *ag) processByPeco(cmd_args []string) error {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()

		out, err := osexec.Command("ag", cmd_args...).Output()
		if err != nil { return }

		self.SetWriter(writer)
		self.printResult(string(out))
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

func (self *ag) printResult(ret string) error {
	re, _ := regexp.Compile(fmt.Sprintf("^%s/([0-9]+)\\.md:([0-9]+):", Context.BodyRoot()))
	
	for _, line := range strings.Split(ret, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) <= 2 { continue }

		bytes, err := LoadPostData(matches[1])
		if err != nil { return err }

		var post esa.PostResponse
		if err := json.Unmarshal(bytes, &post); err != nil { return err }

		// 指定されたカテゴリ配下でない記事だったら弾く
		if self.category == "" && post.Category != "" { continue }
		if len(post.Category) < len(self.category) { continue }
		if ! strings.HasPrefix(post.FullName, self.category+"/") { continue }
		// 指定された条件にマッチしない記事だったら弾く
		if ! self.matchPost(&post) { continue }
		
		self.Println(re.ReplaceAllString(line, fmt.Sprintf("%s:%s:%s: ", matches[1], matches[2], post.FullName)))
	}
	
	return nil
}
