package util

import (
	"os"
	"io/ioutil"
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"errors"
	"reflect"
	"encoding/base64"
	"bytes"
	"golang.org/x/crypto/ssh/terminal"
	log "github.com/sirupsen/logrus"
	"github.com/abiosoft/ishell"
	flags "github.com/jessevdk/go-flags"
	"github.com/flynn-archive/go-shlex"
)

func LocalRootPath() string {
	return filepath.Join(os.Getenv("HOME"), ".esal")
}

func PutError(err error) {
	log.Error(err.Error())
	os.Stderr.Write([]byte(err.Error()+"\n"))
}

func RemoveDup(args []string) []string {
    results := make([]string, 0, len(args))
    encountered := map[string]bool{}
    for i := 0; i < len(args); i++ {
        if !encountered[args[i]] {
            encountered[args[i]] = true
            results = append(results, args[i])
        }
    }
    return results
}

func EncodePath(path string) string {
	separator := string(os.PathSeparator)
	r := strings.NewReplacer("=", "-", "/", "_", "+", ".")
	
	nodes := []string{}
	for _, node := range strings.Split(path, separator) {
		switch node {
		case ".", "..":
			nodes = append(nodes, node)
		default:
			enc_node := base64.StdEncoding.EncodeToString([]byte(node))
			nodes = append(nodes, r.Replace(enc_node))
		}
	}

	return strings.Join(nodes, separator)
}

func DecodePath(path string) string {
	separator := string(os.PathSeparator)
	r := strings.NewReplacer("-", "=",  "_", "/",".","+")
	
	nodes := []string{}
	for _, node := range strings.Split(path, separator) {
		switch node {
		case ".", "..":
			nodes = append(nodes, node)
		default:
			bytes, err := base64.StdEncoding.DecodeString(r.Replace(node))
			if err != nil {
				nodes = append(nodes, node)
			} else {
				nodes = append(nodes, string(bytes))
			}
		}
	}

	return strings.Join(nodes, separator)
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

func ScanString(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
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
		adapter := &IshellAdapter{
			ProcessorRepository: repo,
			ProcessorName: processor_name,
		}

		shell.AddCmd(&ishell.Cmd{
			Name: processor_name,
			Help: repo.GetDescription(processor_name),
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

		err = RunProcessorWithPipe(repo, args[0], args[1:])
		if err != nil { break }
	}
}

func ProcessWithString(name string, repo *ProcessorRepository, code string) {
	for _, line := range strings.Split(code, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#") { continue }
		if line == "" { continue }

		args, err := shlex.Split(line)
		if err != nil {
			PutError(err)
			return
		}
		if len(args) == 0 { continue }

		err = RunProcessorWithPipe(repo, args[0], args[1:])
		if err != nil { break }
	}
}

func RunProcessorWithPipe(repo *ProcessorRepository, first_processor_name string, args []string) error {
	var buf *bytes.Buffer

	args_list := SplitByPipe(args)
	for idx, args := range args_list {
		processor_name := ""
		if idx == 0 {
			processor_name = first_processor_name
		} else if len(args) > 0 {
			processor_name = args[0]

			args = args[1:]
		}

		if processor_name == "" {
			PutError(errors.New("Input is invalid!"))
			return nil
		}

		processor_generator := repo.GetProcessorGenerator(processor_name)
		if processor_generator == nil {
			PutError(errors.New("Unknown command : "+processor_name))
			return nil
		}

		processor := processor_generator()
		parser := flags.NewParser(processor, flags.Default)
		parser.Name = processor_name
		parser.Usage = repo.GetUsage(processor_name)
		
		args, err := parser.ParseArgs(args)
		if err != nil {
			return nil
		}

		if idx == 0 {
			processor.SetReader(os.Stdin)
		} else {
			processor.SetReader(buf)
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
			buf = new(bytes.Buffer)
			processor.SetWriter(buf)
		}

		err = processor.Do(args)
		if err != nil {
			err_type := reflect.ValueOf(err)
			switch fmt.Sprintf("%s", err_type.Type()) {
			case "*util.ProcessorExitRequired":
				return err
			case "*util.ProcessorHelpRequired":
				parser.WriteHelp(os.Stderr)
			default:
				PutError(err)
			}
			return nil
		}
	}

	return nil
}

func SplitByPipe(args []string) [][]string {
	ret := [][]string{}
	nargs := []string{}
	
	for _, arg := range args {
		if arg == "|" {
			ret = append(ret, nargs)
			nargs = []string{}
		} else {
			nargs = append(nargs, arg)
		}
	}

	return append(ret, nargs)
}
