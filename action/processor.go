package action

import (
	"flag"
)

type Processable interface {
	SetOption(flagset *flag.FlagSet)
	Do(args []string) error
}

var processors = make(map[string]Processable)

func NewProcessor(name string) Processable {
	return processors[name]
}
