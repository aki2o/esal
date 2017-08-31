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
		head_beginning_re, _ := regexp.Compile("^#+ +")
		appended := false
		cond_index := 0
		
		for index, line := range strings.Split(string(body_bytes), "\r\n") {
			if ! appended && head_beginning_re.MatchString(line) {
				// まだ追加していなくて、見出し行が見つかったら、そこが追加すべき見出しかどうか判定する
				
				if cond_index >= len(self.InsertConditions) {
					// 指定された条件を全てクリアしているなら、現在の見出しの前（目的の見出しの最後）に追加する
					fmt.Fprint(buf, strings.Join(write_texts, "\r\n")+"\r\n")
					appended = true
				} else {
					// まだ見つかっていない条件が残っているなら、現在の見出しとその条件がマッチするか調べる
					curr_head := head_beginning_re.ReplaceAllString(line, "")
					cond_head := self.InsertConditions[cond_index]
					if curr_head == cond_head {
						// 条件に合った見出しが見つかったので、次の条件に移る
						cond_index = cond_index + 1
					}
				}
			}

			if index > 0 { fmt.Fprint(buf, "\r\n") }
			fmt.Fprint(buf, line)
		}

		if ! appended && cond_index >= len(self.InsertConditions) {
			// 指定された条件を全てクリアしているのに、まだ追加していない場合は、
			// 目的の見出しが記事中の最後の見出しのはずなので、記事の最後に追加する
			fmt.Fprint(buf, strings.Join(write_texts, "\r\n")+"\r\n")
		}
		
		new_body_bytes, err := ioutil.ReadAll(buf)
		if err != nil { return "", err }

		return string(new_body_bytes), nil
	} else if self.Appending {
		return string(body_bytes)+strings.Join(write_texts, "\r\n"), nil
	} else {
		return strings.Join(write_texts, "\r\n"), nil
	}
}
