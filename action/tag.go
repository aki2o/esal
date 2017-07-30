package action

import (
	"os"
	"path/filepath"
	"encoding/json"
	"bufio"
	"io/ioutil"
	"strings"
	"errors"
	"github.com/aki2o/esal/util"
)

type tag struct {
	util.ProcessIO
	Separator string `short:"s" long:"separator" description:"Charactor to separate tags." default:"\n" value-name:"SEPARATOR"`
}

type tagStore struct {
	Values []string `json:"values"`
}

func init() {
	registProcessor(func() util.Processable { return &tag{} }, "tag", "Print/Add tag.", "")
}

func (self *tag) Do(args []string) error {
	var action_name string = ""

	if len(args) > 0 { action_name = args[0] }

	switch action_name {
	case "list":	return self.PrintTags()
	case "add":		return self.AddTags(args[1:])
	// case "remove":	return self.RemoveTags(args[1:])
	default:		return errors.New("Unknown action!")
	}
}

func (self *tag) PrintTags() error {
	tags, err := self.load()
	if err != nil { return err }

	self.Println(strings.Join(tags, self.Separator))
	return nil
}

func (self *tag) AddTags(tags []string) error {
	if len(tags) == 0 { tags = self.ScanArgs() }

	current_tags, err := self.load()
	if err != nil { return err }

	return self.StoreTags(append(current_tags, tags...))
}

func (self *tag) load() ([]string, error) {
	file_path := filepath.Join(self.GetLocalStragePath(), Context.Team+".json")
	
	if ! util.Exists(file_path) { return []string{}, nil }

	bytes, err := ioutil.ReadFile(file_path)
	if err != nil { return []string{}, err }

	var tag_store tagStore
	if err := json.Unmarshal(bytes, &tag_store); err != nil { return []string{}, err }

	return tag_store.Values, nil
}

func (self *tag) StoreTags(tags []string) error {
	if err := util.EnsureDir(self.GetLocalStragePath()); err != nil { return err }

	tag_store := &tagStore{ Values: util.RemoveDup(tags) }
	bytes, err := json.MarshalIndent(tag_store, "", "\t")
	if err != nil { return err }

	fp, err := os.Create(filepath.Join(self.GetLocalStragePath(), Context.Team+".json"))
	if err != nil { return err }
	defer fp.Close()
	
	writer := bufio.NewWriter(fp)
	_, err = writer.Write(bytes)
	if err != nil { return err }
	writer.Flush()
	return nil
}

func (self *tag) GetLocalStragePath() string {
	return filepath.Join(util.LocalRootPath(), "tags")
}
