package util

import (
	"os"
	"io/ioutil"
	"bufio"
	"fmt"
	"strings"
	"flag"
	log "github.com/sirupsen/logrus"
)

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

func ProcessInteractive(name string, processor_generator func(string) Processable) {
	for {
		fmt.Print("> ")
		
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		
		input_tokens := strings.Fields(scanner.Text())
		if len(input_tokens) == 0 { continue }

		command_name := input_tokens[0]

		if command_name == "exit" || command_name == "quit" {
			os.Exit(0)
			return
		}
		
		processor := processor_generator(command_name)
		if processor == nil {
			log.WithFields(log.Fields{ "command": command_name }).Debug("unknown command")
			os.Stderr.Write([]byte("Unknown command!\n"))
			continue
		}
		
		flagset := flag.NewFlagSet(name, flag.PanicOnError)
		processor.SetOption(flagset)
		
		err := flagset.Parse(input_tokens[1:])
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
