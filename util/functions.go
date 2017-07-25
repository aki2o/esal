package util

import (
	"os"
	"io/ioutil"
	"bufio"
	"fmt"
	"path/filepath"
	"syscall"
	"errors"
	"reflect"
	"golang.org/x/crypto/ssh/terminal"
	log "github.com/sirupsen/logrus"
	"github.com/abiosoft/ishell"
	flags "github.com/jessevdk/go-flags"
	"github.com/flynn-archive/go-shlex"
)

func LocalRootPath() string {
	return filepath.Join(os.Getenv("HOME"), ".esa")
}

func PutError(err error) {
	log.Error(err.Error())
	os.Stderr.Write([]byte(err.Error()+"\n"))
}

func GetNodes(path string) []os.FileInfo {
	nodes, err := ioutil.ReadDir(path)
	
	if err != nil {
		PutError(err)
		return []os.FileInfo{}
	}
	
	return nodes
}

func EnsureDir(path string) error {
	if Exists(path) { return nil }
	
	return os.MkdirAll(path, 0777)
}

func Exists(filename string) bool {
    _, err := os.Stat(filename)
	
    return err == nil
}

func Readln(r *bufio.Reader) (string, error) {
    var (
        isPrefix bool  = true
        err      error = nil
        line, ln []byte
    )
	
    for isPrefix && err == nil {
        line, isPrefix, err = r.ReadLine()
        ln = append(ln, line...)
    }
	
    return string(ln), err
}

func CreateFile(path string, body string) error {
	fp, err := os.Create(path)
	if err != nil { return err }
	defer fp.Close()
	
	writer := bufio.NewWriter(fp)
	_, err = writer.WriteString(body)
	if err != nil { return err }
	writer.Flush()

	return nil
}

func ReadAccessToken() string {
	fmt.Print("Enter access_token: ")
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil { panic(err) }

	return string(bytes)
}

func ProcessInteractive(name string, repo *ProcessorRepository) {
	shell := ishell.New()

	for _, processor_name := range repo.ProcessorNames() {
		processor	:= repo.NewProcessor(processor_name)
		description := repo.GetDescription(processor_name)
		usage		:= repo.GetUsage(processor_name)

		adapter := &IshellAdapter{
			Processor: processor,
			ProcessorName: processor_name,
			ProcessorDescription: description,
			ProcessorUsage: usage,
		}

		shell.AddCmd(&ishell.Cmd{
			Name: processor_name,
			Help: fmt.Sprintf("%s\nUsage:\n  %s %s\n", description, processor_name, usage),
			Func: adapter.Adapt,
		})
	}

	shell.Run()
}

func ProcessNonInteractive(name string, repo *ProcessorRepository) {
	for {
		fmt.Print("\u0003")
		
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		args, err := shlex.Split(scanner.Text())
		if err != nil {
			PutError(err)
			continue
		}

		if len(args) == 0 { continue }
		if args[0] == "exit" { break }

		processor_name := args[0]
		processor := repo.NewProcessor(processor_name)
		if processor == nil {
			PutError(errors.New("Unknown command!"))
			continue
		}

		parser := flags.NewParser(processor, flags.Default)
		parser.Name = processor_name
		parser.Usage = repo.GetUsage(processor_name)
		
		args, err = parser.ParseArgs(args)
		if err != nil {
			PutError(err)
			continue
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
