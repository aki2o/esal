package action

import (
	"os/exec"
	"fmt"
	"regexp"
	"io"
	"bufio"
	"errors"
	"encoding/json"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type ag struct {
	pecoable
}

func init() {
	registProcessor(func() util.Processable { return &ag{} }, "ag", "Execute ag command.", "[OPTIONS] PATTERN")
}

func (self *ag) Do(args []string) error {
	cmd_args := append([]string{"-G", ".md$"}, args...)
	cmd_args = append(cmd_args, Context.BodyRoot())

	path_re, _ := regexp.Compile("(?m)^"+Context.BodyRoot()+"/[0-9]+\\.md:[0-9]+:")
	
	var result string
	var out []byte
	var err error
	
	if self.PecoRequired() {
		provider := func(writer *io.PipeWriter) {
			defer writer.Close()

			out, err := exec.Command("ag", cmd_args...).Output()
			if err != nil { return }

			rich_writer := bufio.NewWriter(writer)

			fmt.Fprintf(rich_writer, path_re.ReplaceAllStringFunc(string(out), self.appendPostName))
			rich_writer.Flush()
		}

		result, _, err = pipePeco(provider, "Query")
	} else {
		out, err = exec.Command("ag", cmd_args...).Output()
		if err != nil { return err }
		
		result = path_re.ReplaceAllStringFunc(string(out), self.appendPostName)
	}
	
	if err != nil { return err }

	if self.PecoRequired() {
		if result == "" { return nil }
		
		re, _ := regexp.Compile("^([0-9]+):[0-9]+:")
		matches := re.FindStringSubmatch(result)

		if len(matches) > 1 {
			open := &open{}
			return open.Do([]string{matches[1]})
		} else {
			return errors.New("Can't find post from result!")
		}
	} else {
		fmt.Print(result)
		return nil
	}
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
