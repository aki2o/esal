package action

import (
	"flag"
	"errors"
	"fmt"
	"os"
	"github.com/aki2o/esa-cui/util"
)

type lock struct {}

func init() {
	processors["lock"] = &lock{}
}

func (self *lock) SetOption(flagset *flag.FlagSet) {
}

func (self *lock) Do(args []string) error {
	var post_number string = ""

	if len(args) > 0 { post_number = args[0] }
	
	if post_number == "" {
		return errors.New("Require post number!")
	}

	lock_file_path := fmt.Sprintf("%s/%s", Context.Cwd, GetLocalPostFileName(post_number, "lock"))
	if util.Exists(lock_file_path) { return nil }
	
	fp, err := os.Create(lock_file_path)
	if err != nil { return err }
	
	fp.Close()
	return nil
}
