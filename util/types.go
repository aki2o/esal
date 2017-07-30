package util

import (
	"os"
	"reflect"
	"fmt"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"bytes"
	"github.com/abiosoft/ishell"
	flags "github.com/jessevdk/go-flags"
)

type Processable interface {
	Do(args []string) error
	SetWriter(writer io.Writer)
	SetReader(reader io.Reader)
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
	// 引数が無い場合、パイプで渡された出力を引数として受け取る
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
	var buf bytes.Buffer

	args_list := SplitByPipe(args)
	for idx, args := range args_list {
		processor_name := ""
		if idx == 0 {
			processor_name = self.ProcessorName
		} else if len(args) > 0 {
			processor_name = args[0]

			args = args[1:]
		}

		if processor_name == "" {
			PutError(errors.New("Input is invalid!"))
			return
		}

		processor_generator := self.ProcessorRepository.GetProcessorGenerator(processor_name)
		if processor_generator == nil {
			PutError(errors.New("Unknown command : "+processor_name))
			return
		}

		processor := processor_generator()
		parser := flags.NewParser(processor, flags.Default)
		parser.Name = processor_name
		parser.Usage = self.ProcessorRepository.GetUsage(processor_name)
		
		args, err := parser.ParseArgs(args)
		if err != nil {
			PutError(err)
			return
		}

		if idx == 0 {
			processor.SetReader(os.Stdin)
		} else {
			processor.SetReader(&buf)
		}

		if idx + 1 == len(args_list) {
			// 最後のコマンドは、標準出力へ出力
			processor.SetWriter(os.Stdout)
		} else {
			// 後続のコマンドがある場合は、出力をバッファに溜める
			//
			// NOTE:
			//   io.Pipe() を使いたかったけど、同じ関数内で Writer/Reader を使うと
			//   デッドロックしてしまい、それを回避する実装が思いつかなかった
			//
			//   http://christina04.hatenablog.com/entry/2017/01/06/190000
			//
			buf = bytes.Buffer{}
			processor.SetWriter(&buf)
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
}
