package action

import (
	"errors"
	"flag"
	"fmt"
	"encoding/json"
	"strconv"
	"io"
	"github.com/aki2o/esa-cui/util"
)

type cat struct {
	json_format bool
	pecolize bool
	recursive bool
}

type postProperty struct {
	// esa.PostResponse の中で必要そうなやつだけに絞る
	Category      string `json:"category"`
	CommentsCount int    `json:"comments_count"`
	CreatedAt     string `json:"created_at"`
	CreatedBy     struct {
		Icon       string `json:"icon"`
		Name       string `json:"name"`
		ScreenName string `json:"screen_name"`
	} `json:"created_by"`
	DoneTasksCount  int      `json:"done_tasks_count"`
	FullName        string   `json:"full_name"`
	Kind            string   `json:"kind"`
	Message         string   `json:"message"`
	Name            string   `json:"name"`
	Number          int      `json:"number"`
	OverLapped      bool     `json:"overlapped"`
	RevisionNumber  int      `json:"revision_number"`
	Star            bool     `json:"star"`
	StargazersCount int      `json:"stargazers_count"`
	Tags            []string `json:"tags"`
	TasksCount      int      `json:"tasks_count"`
	UpdatedAt       string   `json:"updated_at"`
	UpdatedBy       struct {
		Icon       string `json:"icon"`
		Name       string `json:"name"`
		ScreenName string `json:"screen_name"`
	} `json:"updated_by"`
	URL           string `json:"url"`
	Watch         bool   `json:"watch"`
	WatchersCount int    `json:"watchers_count"`
	Wip           bool   `json:"wip"`
	
	LocalPath string `json:"local_path"`
	Locked bool `json:"locked"`
}

func init() {
	addProcessor(&cat{}, "cat", "Print a post body text as markdown.")
}

func (self *cat) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.json_format, "json", false, "Show properties as json.")
	flagset.BoolVar(&self.pecolize, "peco", false, "Exec with peco.")
	flagset.BoolVar(&self.recursive, "r", false, "Recursively for peco.")
}

func (self *cat) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }

	if self.pecolize {
		next_path, err := self.runPeco(path)
		if err != nil { return err }

		path = next_path
	}

	dir_path, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	if self.json_format {
		bytes := LoadPostData(AbsolutePathOf(dir_path), post_number, "json")

		var post postProperty
		if err := json.Unmarshal(bytes, &post); err != nil { return err }

		post.LocalPath	= GetLocalPostPath(post.Category, strconv.Itoa(post.Number), "md")
		post.Locked		= util.Exists(GetLocalPostPath(post.Category, strconv.Itoa(post.Number), "lock"))

		json_bytes, _ := json.MarshalIndent(post, "", "\t")
		
		fmt.Println(string(json_bytes))
	} else {
		bytes := LoadPostData(AbsolutePathOf(dir_path), post_number, "md")
		
		fmt.Println(string(bytes))
	}
	return nil
}

func (self *cat) runPeco(path string) (string, error) {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()
		
		ls := &ls{ writer: writer, recursive: self.recursive, file_only: true }
		ls.printNodesIn(path, AbsolutePathOf(path))
	}

	return pipePeco(provider)
}
