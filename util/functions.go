package util

import (
	"os"
	"io/ioutil"
	"bufio"
	"fmt"
	"path/filepath"
	"syscall"
	"strings"
	"errors"
	"flag"
	"golang.org/x/crypto/ssh/terminal"
	log "github.com/sirupsen/logrus"
	"github.com/abiosoft/ishell"
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
		processor	:= repo.GetProcessor(processor_name)
		usage		:= repo.GetUsage(processor_name)

		adapter := &IshellAdapter{
			Processor: processor,
			ProcessorName: processor_name,
			ProcessorUsage: usage,
		}
		
		shell.AddCmd(&ishell.Cmd{
			Name: processor_name,
			Help: usage,
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
		
		args := strings.Split(strings.TrimSpace(scanner.Text()), " ")

		if len(args) == 0 { continue }
		if args[0] == "exit" { break }
		
		processor := repo.GetProcessor(args[0])
		if processor == nil {
			PutError(errors.New("Unknown command!"))
			continue
		}

		flagset := flag.NewFlagSet(args[0], flag.PanicOnError)
		processor.SetOption(flagset)

		err := flagset.Parse(args[1:])
		if err != nil {
			PutError(err)
			continue
		}

		err = processor.Do(flagset.Args())
		if err != nil {
			PutError(err)
			continue
		}
	}
}
