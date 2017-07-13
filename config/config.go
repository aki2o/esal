package config

import (
	"os"
	"io/ioutil"
	"bufio"
	"path/filepath"
	"encoding/json"
	"time"
	"github.com/aki2o/esa-cui/util"
)

type QueryEntry struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

type Query struct {
	Name string `json:"name"`
	Entries []QueryEntry `json:"entries"`
	SynchronizedAt time.Time `json:"synchronized_at"`
}

type TeamConfig struct {
	Name string `json:"name"`
	Queries []Query `json:"queries"`
}

type Config struct {
	Teams []TeamConfig `json:"teams"`
}

var config_path string = filepath.Join(util.LocalRootPath(), "config.json")

var Global *Config
var Team *TeamConfig

func Load(team string) {
	Global = loadConfig()
	
	for _, team_config := range Global.Teams {
		if team_config.Name != team { continue }

		Team = &team_config
		return
	}

	Team = &TeamConfig{ Name: team }
	Global.Teams = append(Global.Teams, *Team)
}

func Save() {
	teams := make([]TeamConfig, len(Global.Teams))
	for index, team_config := range Global.Teams {
		if team_config.Name == Team.Name {
			teams[index] = *Team
		} else {
			teams[index] = team_config
		}
	}
	Global.Teams = teams
	
	bytes, err := json.MarshalIndent(Global, "", "\t")
	if err != nil { panic(err) }
	
	fp, err := os.Create(config_path)
	if err != nil { panic(err) }
	defer fp.Close()
	
	writer := bufio.NewWriter(fp)
	_, err = writer.WriteString(string(bytes))
	if err != nil { panic(err) }
	
	writer.Flush()
}

func loadConfig() *Config {
	if ! util.Exists(config_path) {
		return &Config{}
	}

	var config Config
	bytes, err := ioutil.ReadFile(config_path)
	if err != nil { panic(err) }

	err = json.Unmarshal(bytes, &config)
	if err != nil { panic(err) }

	return &config
}
