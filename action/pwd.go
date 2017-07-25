package action

import (
	"fmt"
	"github.com/aki2o/esa-cui/util"
)

type pwd struct {}

func init() {
	registProcessor(func() util.Processable { return &pwd{} }, "pwd", "Print current work category.", "")
}

func (self *pwd) Do(args []string) error {
	fmt.Println("/"+CategoryOf(Context.Cwd))
	return nil
}
