package util

import (
	"os"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"github.com/abiosoft/ishell"
)

type Processable interface {
	Do(args []string) error
	SetWriter(writer io.Writer)
	SetReader(reader io.Reader)
}

type ProcessorExitRequired struct {}

func (self *ProcessorExitRequired) Error() string {
	return ""
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

type ProcessIO struct {
	Writer io.Writer
	Reader io.Reader
}

func (self *ProcessIO) SetWriter(writer io.Writer) {
	self.Writer = writer
}

func (self *ProcessIO) SetReader(reader io.Reader) {
	self.Reader = reader
}

func (self *ProcessIO) Println(a ...interface{}) {
	fmt.Fprintln(self.Writer, a...)
}

func (self *ProcessIO) Printf(format string, a ...interface{}) {
	fmt.Fprintf(self.Writer, format, a...)
}

func (self *ProcessIO) ScanArgs() []string {
	if self.Reader == os.Stdin { return []string{} }
	
	bytes, err := ioutil.ReadAll(self.Reader)
	if err != nil { return []string{} }
	
	args := strings.Split(string(bytes), "\n")
	// 改行で終わっている場合は、それは含めない
	if len(args) > 0 && args[len(args)-1] == "" {
		args = args[0:len(args)-1]
	}

	return args
}

type IshellAdapter struct {
	ProcessorRepository *ProcessorRepository
	ProcessorName string
}

func (self *IshellAdapter) Adapt(ctx *ishell.Context) {
	self.Run(ctx.Args)
}

func (self *IshellAdapter) Run(args []string) {
	RunProcessorWithPipe(self.ProcessorRepository, self.ProcessorName, args)
}
