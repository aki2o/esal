package action

import (
	"errors"
	"encoding/json"
	"strconv"
	"os/exec"
	"github.com/aki2o/esal/util"
)

type cat struct {
	util.ProcessIO
	pecoable
	ByBrowser bool `short:"b" long:"browser" description:"Open by browser."`
	JsonRequired bool `short:"j" long:"json" description:"Show properties as json."`
	WithoutIndent bool `short:"n" long:"noindent" description:"For json option, show without indent."`
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
	registProcessor(func() util.Processable { return &cat{} }, "cat", "Print a post.", "[OPTIONS] POST...")
}

func (self *cat) Do(args []string) error {
	if len(args) == 0 && self.PecoRequired() {
		var err error
		args, err = selectNodeByPeco("", false)
		if err != nil { return err }
	}

	for _, path := range args {
		if err := self.process(path); err != nil { return err }
	}
	return nil
}

func (self *cat) process(path string) error {
	_, post_number := DirectoryPathAndPostNumberOf(path)
	if post_number == "" {
		return errors.New("Require post number!")
	}

	if self.ByBrowser {
		bytes, err := LoadPostData(post_number)
		if err != nil { return err }

		var post postProperty
		if err := json.Unmarshal(bytes, &post); err != nil { return err }

		if err := exec.Command(BrowserCommand(), post.URL).Run(); err != nil { return err }
	} else if self.JsonRequired {
		bytes, err := LoadPostData(post_number)
		if err != nil { return err }

		var post postProperty
		if err := json.Unmarshal(bytes, &post); err != nil { return err }

		post.LocalPath	= GetPostBodyPath(strconv.Itoa(post.Number))
		post.Locked		= util.Exists(GetPostLockPath(strconv.Itoa(post.Number)))

		var json_bytes []byte
		if self.WithoutIndent {
			json_bytes, _ = json.Marshal(post)
		} else {
			json_bytes, _ = json.MarshalIndent(post, "", "\t")
		}
		
		self.Println(string(json_bytes))
	} else {
		bytes, err := LoadPostBody(post_number)
		if err != nil { return err }
		
		self.Println(string(bytes))
	}
	return nil
}
