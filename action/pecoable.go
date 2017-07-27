package action

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"bytes"
	"reflect"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/peco/peco"
	"github.com/nsf/termbox-go"
	"github.com/aki2o/esal/config"
)

type pecoable struct {
	Pecolize bool `short:"p" long:"peco" description:"Exec with peco."`
	NoPecolize bool `short:"P" long:"nopeco" description:"Exec without peco."`
}

func (self *pecoable) PecoRequired() bool {
	if self.NoPecolize { return false }
	
	return self.Pecolize || Context.PecoPreferred
}

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

func SetupPeco(peco_preferred bool) {
	peco.ActionFunc(doBackwardNode).Register("EsaBackwardNode", termbox.KeyCtrlH)
	peco.ActionFunc(doForwardNode).Register("EsaForwardNode", termbox.KeyCtrlL)

	if config.Global.PecoPreferred || peco_preferred {
		Context.PecoPreferred = true
	}
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

func pipePeco(provider func(*io.PipeWriter), prompt string) (string, string, error) {
	from_provider_reader, to_peco_writer := io.Pipe()
	
	go provider(to_peco_writer)
	
	from_peco_reader, to_self_writer := io.Pipe()

	var status bytes.Buffer
	
	go func() {
		defer to_self_writer.Close()
		
		peco := peco.New()
		peco.Argv	= []string{"--on-cancel", "error", "--prompt", prompt+":"}
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

func selectNodeByPeco(path string, category_only bool) ([]string, error) {
	path = "/"+CategoryOf(PhysicalPathOf(path))
	
	for {
		provider := func(writer *io.PipeWriter) {
			defer writer.Close()
			
			ls_process := &ls{ writer: writer, CategoryOnly: category_only }
			ls_process.printNodesIn(path, PhysicalPathOf(path))
		}

		var prompt string
		if category_only {
			prompt = "Select category"
		} else {
			prompt = "Select post"
		}
		
		selected, status, err := pipePeco(provider, prompt)
		if err != nil { return []string{}, err }
		log.WithFields(log.Fields{ "selected": selected, "status": status }).Debug("return peco")
		
		switch status {
		case "forward":
			path = selected
		case "backward":
			path = ParentOf(path)
		default:
			return strings.Split(selected, "\n"), nil
		}
	}
}
