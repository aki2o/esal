package action

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"bytes"
	"regexp"
	"reflect"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/peco/peco"
	"github.com/nsf/termbox-go"
)

type errBackwardNode struct{}

func (err errBackwardNode) Error() string {
	return "backward"
}

type errForwardNode struct{
	node_path string
}

func (err errForwardNode) Error() string {
	return "forward"
}

func SetupPeco() {
	peco.ActionFunc(doBackwardNode).Register("EsaBackwardNode", termbox.KeyCtrlH)
	peco.ActionFunc(doForwardNode).Register("EsaForwardNode", termbox.KeyCtrlL)
}

func doBackwardNode(ctx context.Context, state *peco.Peco, e termbox.Event) {
	state.Exit(errBackwardNode{})
}

func doForwardNode(ctx context.Context, state *peco.Peco, e termbox.Event) {
	node_path := ""
	
	if l, err := state.CurrentLineBuffer().LineAt(state.Location().LineNumber()); err == nil {
		node_path = l.Buffer()
	}
	
	state.Exit(errForwardNode{ node_path: node_path})
}

func pipePeco(provider func(*io.PipeWriter)) (string, string, error) {
	from_provider_reader, to_peco_writer := io.Pipe()
	
	go provider(to_peco_writer)
	
	from_peco_reader, to_self_writer := io.Pipe()

	var status bytes.Buffer
	
	go func() {
		defer to_self_writer.Close()
		
		peco := peco.New()
		peco.Argv	= []string{"--on-cancel", "error"}
		peco.Stdin	= from_provider_reader
		peco.Stdout = to_self_writer
		
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := peco.Run(ctx); err != nil {
			// peco の終了を判断する機能が公開されていないので、 reflect を使って、無理矢理実装
			err_type := reflect.ValueOf(err)
			switch fmt.Sprintf("%s", err_type.Type()) {
			case "peco.errCollectResults":
				peco.PrintResults()
			case "action.errBackwardNode":
				status.WriteString(err.Error())
				log.Debug("Peco requires backward node")
			case "action.errForwardNode":
				status.WriteString(err.Error())

				peco.PrintResults()
				log.Debug("Peco requires forward node")
			case "*peco.errWithExitStatus":
				return
			default:
				log.Warn(fmt.Sprintf("Peco return %s", err_type.Type()))
			}
			return
		}
	}()
	
	bytes, err := ioutil.ReadAll(from_peco_reader)
	if err != nil { return "", "", err }

	return strings.TrimRight(string(bytes), "\n"), status.String(), nil
}

func selectNodeByPeco(path string, directory_only bool) (string, error) {
	path = "/"+CategoryOf(PhysicalPathOf(path))
	
	for {
		provider := func(writer *io.PipeWriter) {
			defer writer.Close()
			
			ls := &ls{ writer: writer, directory_only: directory_only }
			ls.printNodesIn(path, PhysicalPathOf(path))
		}

		selected, status, err := pipePeco(provider)
		if err != nil { return "", err }
		log.WithFields(log.Fields{ "selected": selected, "status": status }).Debug("return peco")
		
		switch status {
		case "forward":
			path = selected
		case "backward":
			if path != "/" {
				re, _ := regexp.Compile("/[^/]*$")
				
				path = re.ReplaceAllString(path, "")
				path = re.ReplaceAllString(path, "")

				if path == "" { path = "/" }
			}
		default:
			return selected, nil
		}
	}
}
