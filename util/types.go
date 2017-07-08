package util

import (
	"flag"
)

type Processable interface {
	SetOption(flagset *flag.FlagSet)
	Do(args []string) error
}
