package action

import (
	"flag"
	"fmt"
	"strings"
	"os"
)

type pwd struct {}

func init() {
	addProcessor(&pwd{}, "pwd", "Print current work category.")
}

func (self *pwd) SetOption(flagset *flag.FlagSet) {
}

func (self *pwd) Do(args []string) error {
	separator := string(os.PathSeparator)
	root_dirs := strings.Split(Context.Root(), separator)
	curr_dirs := strings.Split(Context.Cwd, separator)[len(root_dirs):]

	fmt.Println("/"+strings.Join(curr_dirs, "/"))
	return nil
}
