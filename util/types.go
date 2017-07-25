package util

import (
	"os"
	"reflect"
	"fmt"
	"github.com/abiosoft/ishell"
	flags "github.com/jessevdk/go-flags"
)

type Processable interface {
	Do(args []string) error
}

type ProcessorHelpRequired struct {}

func (self *ProcessorHelpRequired) Error() string {
	return ""
}

type ProcessorRepository struct {
	processor_generators map[string]func() Processable
	descriptions map[string]string
	usages map[string]string
}

func (self *ProcessorRepository) ProcessorNames() []string {
	self.setup()
	
	var ret = make([]string, len(self.processor_generators))
	
	index := 0
	for name, _ := range self.processor_generators {
		ret[index] = name
		index += 1
	}
	
	return ret
}

func (self *ProcessorRepository) setup() {
	if self.processor_generators == nil { self.processor_generators = make(map[string]func() Processable) }
	if self.descriptions == nil { self.descriptions = make(map[string]string) }
	if self.usages == nil { self.usages = make(map[string]string) }
}

func (self *ProcessorRepository) SetProcessorGenerator(name string, processor func() Processable) {
	self.setup()
	self.processor_generators[name] = processor
}

func (self *ProcessorRepository) GetProcessorGenerator(name string) func() Processable {
	self.setup()
	return self.processor_generators[name]
}

func (self *ProcessorRepository) SetDescription(name string, description string) {
	self.setup()
	self.descriptions[name] = description
}

func (self *ProcessorRepository) GetDescription(name string) string {
	self.setup()
	return self.descriptions[name]
}

func (self *ProcessorRepository) SetUsage(name string, usage string) {
	self.setup()
	self.usages[name] = usage
}

func (self *ProcessorRepository) GetUsage(name string) string {
	self.setup()
	return self.usages[name]
}

type IshellAdapter struct {
	ProcessorGenerator func() Processable
	ProcessorName string
	ProcessorDescription string
	ProcessorUsage string
}

func (self *IshellAdapter) Adapt(ctx *ishell.Context) {
	self.Run(ctx.Args)
}

func (self *IshellAdapter) Run(args []string) {
	processor := self.ProcessorGenerator()
	parser := flags.NewParser(processor, flags.Default)
	parser.Name = self.ProcessorName
	parser.Usage = self.ProcessorUsage
	
	args, err := parser.ParseArgs(args)
	if err != nil {
		PutError(err)
		return
	}
	
	err = processor.Do(args)
	if err != nil {
		err_type := reflect.ValueOf(err)
		switch fmt.Sprintf("%s", err_type.Type()) {
		case "util.ProcessorHelpRequired":
			parser.WriteHelp(os.Stderr)
		default:
			PutError(err)
		}
		return
	}
}
