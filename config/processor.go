package config

import (
	"github.com/aki2o/esa-cui/util"
)

var processors = make(map[string]util.Processable)

func NewProcessor(name string) util.Processable {
	return processors[name]
}
