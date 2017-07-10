package util

import (
	"os"
	"io/ioutil"
	"bufio"
	"flag"
	"path/filepath"
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

func ProcessInteractive(name string, repo *ProcessorRepository) {
	shell := ishell.New()

	for _, processor_name := range repo.ProcessorNames() {
		processor := repo.GetProcessor(processor_name)
		
		shell.AddCmd(&ishell.Cmd{
			Name: processor_name,
			Func: func (c *ishell.Context) {
				flagset := flag.NewFlagSet(name, flag.PanicOnError)
				processor.SetOption(flagset)
				
				err := flagset.Parse(c.Args)
				if err != nil {
					PutError(err)
					return
				}
				
				err = processor.Do(flagset.Args())
				if err != nil {
					PutError(err)
					return
				}
			},
		})
	}

	shell.Run()
}
