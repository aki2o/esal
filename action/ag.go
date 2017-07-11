package action

import (
	"flag"
	"os/exec"
	"fmt"
	"regexp"
	"io"
	"bufio"
	"errors"
)

type ag struct {
	pecolize bool
}

func init() {
	addProcessor(&ag{}, "ag", "Execute ag command.")
}

func (self *ag) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
}

func (self *ag) Do(args []string) error {
	cmd_args := append([]string{"-G", ".md$"}, args...)
	cmd_args = append(cmd_args, Context.Cwd)

	local_root_re, _ := regexp.Compile("(?m)^"+Context.Root())
	
	var result string
	var out []byte
	var err error
	
	if self.pecolize {
		provider := func(writer *io.PipeWriter) {
			defer writer.Close()

			out, err = exec.Command("ag", cmd_args...).Output()
			if err != nil { return }

			rich_writer := bufio.NewWriter(writer)

			fmt.Fprintf(rich_writer, local_root_re.ReplaceAllString(string(out), ""))
			rich_writer.Flush()
		}

		result, err = pipePeco(provider)
	} else {
		out, err = exec.Command("ag", cmd_args...).Output()
		if err != nil { return err }
		
		result = local_root_re.ReplaceAllString(string(out), "")
	}
	
	if err != nil { return err }

	if self.pecolize {
		open := &open{}

		path_re, _ := regexp.Compile("^(.+?)\\.md:[0-9]+:")
		matches := path_re.FindStringSubmatch(result)

		if len(matches) > 1 {
			return open.Do([]string{matches[1]})
		} else {
			return errors.New("Can't find post from result!")
		}
	} else {
		fmt.Print(result)
		return nil
	}
}
