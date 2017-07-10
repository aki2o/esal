package util

import (
	"flag"
)

type Processable interface {
	SetOption(flagset *flag.FlagSet)
	Do(args []string) error
}

type ProcessorRepository struct {
	processors map[string]Processable
}

func (self *ProcessorRepository) ProcessorNames() []string {
	self.setup()
	
	var ret = make([]string, len(self.processors))
	
	index := 0
	for name, _ := range self.processors {
		ret[index] = name
		index += 1
	}
	
	return ret
}

func (self *ProcessorRepository) SetProcessor(name string, processor Processable) {
	self.setup()
	self.processors[name] = processor
}

func (self *ProcessorRepository) GetProcessor(name string) Processable {
	self.setup()
	return self.processors[name]
}

func (self *ProcessorRepository) setup() {
	if self.processors == nil { self.processors = make(map[string]Processable) }
}
