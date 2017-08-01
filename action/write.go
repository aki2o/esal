package action

import (
	"strings"
	"errors"
	"regexp"
	"bytes"
	"fmt"
	"io/ioutil"
	"github.com/aki2o/esal/util"
)

type write struct {
	util.ProcessIO
	pecoable
	Appending bool `short:"a" long:"append" description:"Append into the post body."`
	InsertConditions []string `short:"i" long:"insert" description:"Insert into the end of SECTION." value-name:"SECTION"`
	ReplaceRegexp string `short:"r" long:"replace" description:"Replace the matched part to REGEXP." value-name:"REGEXP"`
}

func init() {
	RegistProcessor(func() util.Processable { return &write{} }, "write", "Write text into post.", "[OPTIONS] POST...")
}

func (self *write) Do(args []string) error {
	if len(args) == 0 && self.PecoRequired() {
		var err error
		args, err = selectNodeByPeco("", false)
		if err != nil { return err }
	}

	write_texts := self.ScanArgs()
	if len(write_texts) == 0 { return nil }

	for _, path := range args {
		if err := self.process(path, write_texts); err != nil { return err }
	}
	return nil
}

func (self *write) process(path string, write_texts []string) error {
	_, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_process := &lock{}
	if err := lock_process.Do([]string{ path }); err != nil { return err }

	new_body, err := self.makeBody(post_number, write_texts)
	if err != nil { return err }

	err = util.CreateFile(GetPostBodyPath(post_number), new_body)
	if err != nil { return err }

	return nil
}

func (self *write) makeBody(post_number string, write_texts []string) (string, error) {
	body_bytes, err := LoadPostBody(post_number)
	if err != nil { return "", err }

	if self.ReplaceRegexp != "" {
		re, err := regexp.Compile(self.ReplaceRegexp)
		if err != nil { return "", err }

		return re.ReplaceAllString(string(body_bytes), strings.Join(write_texts, "\n")), nil
	} else if len(self.InsertConditions) > 0 {
		buf := new(bytes.Buffer)
		block_beginning_re, _ := regexp.Compile("^(#|```)")
		cond_index := 0
		checked_next_block := ""
		
		for _, line := range strings.Split(string(body_bytes), "\n") {
			if checked_next_block != "" {
				if strings.HasPrefix(line, checked_next_block) {
					fmt.Fprint(buf, strings.Join(write_texts, "\n"))
					checked_next_block = ""
				}
			} else if cond_index <= len(self.InsertConditions) {
				if strings.HasPrefix(line, self.InsertConditions[cond_index]) {
					cond_index = cond_index + 1
				}
				if cond_index > len(self.InsertConditions) {
					checked_next_block = block_beginning_re.FindString(line)
				}
			}

			fmt.Fprintln(buf, line)
		}

		new_body_bytes, err := ioutil.ReadAll(buf)
		if err != nil { return "", err }

		return string(new_body_bytes), nil
	} else if self.Appending {
		return string(body_bytes)+strings.Join(write_texts, "\n"), nil
	} else {
		return strings.Join(write_texts, "\n"), nil
	}
}
