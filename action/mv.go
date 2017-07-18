package action

import (
	"flag"
	"errors"
	"os"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type mv struct {}

func init() {
	addProcessor(&mv{}, "mv", "Move post, category.")
}

func (self *mv) SetOption(flagset *flag.FlagSet) {
}

func (self *mv) Do(args []string) error {
	if len(args) < 2 { return errors.New("Require posts, categories and category as destination!") }

	dest_path := args[len(args)-1]
	dest_category := CategoryOf(PhysicalPathOf(dest_path))

	for _, path := range args[:len(args)-1] {
		dir_path, post_number := DirectoryPathAndPostNumberOf(path)
		if post_number == "" {
			category := CategoryOf(PhysicalPathOf(dir_path))

			log.WithFields(log.Fields{ "from": category, "to": dest_category }).Debug("request batch move")
			err := Context.Client.Category.BatchMove(Context.Team, category, dest_category)
			if err != nil {
				log.WithFields(log.Fields{ "from": category, "to": dest_category, "error": err.Error() }).Error("failed to batch move")
				fmt.Fprintf(os.Stderr, "Failed to move '%s' to '%s' : %s\n", category, dest_category, err.Error())
			}
		} else {
			update_process := &update{ category: dest_category, nobody: true, lock_keeping: true }
			err := update_process.Do([]string{ path })
			if err != nil {
				log.WithFields(log.Fields{ "from": path, "to": dest_category, "error": err.Error() }).Error("failed to batch move")
				fmt.Fprintf(os.Stderr, "Failed to move '%s' to '%s' : %s\n", path, dest_category, err.Error())
			}
		}
	}
	return nil
}
