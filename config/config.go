package config

import (
	"os"
	"io/ioutil"
	"bufio"
	"path/filepath"
	"encoding/json"
	"github.com/aki2o/esa-cui/util"
)

type Query struct {
	Name string `json:"name"`
	Entries []struct {
		Key string `json:"key"`
		Value string `json:"value"`
	} `json:"entries"`
}

type TeamConfig struct {
	Name string `json:"name"`
	Queries []Query `json:"queries"`
}

type Config struct {
	Teams []TeamConfig `json:"teams"`
}

var config_path string = filepath.Join(os.Getenv("HOME"), ".esa", "config.json")

var Current *TeamConfig

func Load(team string) {
	for _, team_config := range loadConfig().Teams {
		if team_config.Name != team { continue }

		Current = &team_config
		return
	}

	Current = &TeamConfig{}
}

func Save() {
	config := &Config{}
	teams := []TeamConfig{}

	teams = append(teams, *Current)
	for _, team_config := range loadConfig().Teams {
		if team_config.Name == Current.Name { continue }

		teams = append(teams, team_config)
	}
	config.Teams = teams
	
	bytes, err := json.MarshalIndent(config, "", "\t")
	if err != nil { panic(err) }
	fp, err := os.Create(config_path)
	if err != nil { panic(err) }
	writer := bufio.NewWriter(fp)
	_, err = writer.WriteString(string(bytes))
	if err != nil { panic(err) }
	writer.Flush()
}

func loadConfig() *Config {
	if ! util.Exists(config_path) { return &Config{} }

	var config Config
	bytes, err := ioutil.ReadFile(config_path)
	if err != nil { panic(err) }

	err = json.Unmarshal(bytes, &config)
	if err != nil { panic(err) }

	return &config
}
