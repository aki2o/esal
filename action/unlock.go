package action

import (
	"flag"
	"errors"
	"fmt"
	"os"
	"github.com/aki2o/esa-cui/util"
)

type unlock struct {}

func init() {
	addProcessor(&unlock{}, "unlock", "Stop to guard a post from updated by SYNC.")
}

func (self *unlock) SetOption(flagset *flag.FlagSet) {
}

func (self *unlock) Do(args []string) error {
	var post_number string = ""

	if len(args) > 0 { post_number = args[0] }
	
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_file_path := fmt.Sprintf("%s/%s", Context.Cwd, GetLocalPostFileName(post_number, "lock"))
	if ! util.Exists(lock_file_path) { return nil }
	
	if err := os.Remove(lock_file_path); err != nil { return err }
	
	return nil
}
