package action

import (
	"fmt"
	"flag"
	"os"
	"io"
	"io/ioutil"
	"context"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/peco/peco"
)

type cd struct {
	pecolize bool
}

func init() {
	processors["cd"] = &cd{}
}

func (self *cd) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
}

func (self *cd) Do(args []string) error {
	var path string = ""
	var err error
	if len(args) > 0 { path = args[0] }

	var next_abs_path string = ""
	if self.pecolize {
		next_path, err := runPeco(path, AbsolutePathOf(path))
		if err != nil { return err }

		next_abs_path = AbsolutePathOf(next_path)
	} else {
		next_abs_path = AbsolutePathOf(path)
	}
	
	info, err := os.Stat(next_abs_path)
	if err != nil { return err }
	if ! info.IsDir() { return fmt.Errorf("Not directory : %s", next_abs_path) }

	Context.Cwd = next_abs_path
	log.WithFields(log.Fields{ "cwd": Context.Cwd }).Info("update cwd")

	return nil
}

func runPeco(path string, abs_path string) (string, error) {
	from_ls_reader, to_peco_writer := io.Pipe()

	go func() {
		ls := &ls{ writer: to_peco_writer }
		ls.printNodesIn(path, abs_path)
		
		to_peco_writer.Close()
	}()

	from_peco_reader, to_self_writer := io.Pipe()

	var peco_err error = nil
	go func() {
		peco := peco.New()
		peco.Argv	= []string{}
		peco.Stdin	= from_ls_reader
		peco.Stdout = to_self_writer

		ctx, cancel := context.WithCancel(context.Background())
		if err := peco.Run(ctx); err != nil {
			peco_err = err
		}

		peco.PrintResults()
		
		to_self_writer.Close()
		cancel()
	}()
	if peco_err != nil { return "", peco_err }
	
	bytes, err := ioutil.ReadAll(from_peco_reader)
	if err != nil { return "", err }

	return strings.TrimRight(string(bytes), "\n"), nil
}
