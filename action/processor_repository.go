package action

import (
	"github.com/aki2o/esa-cui/util"
)

var repo *util.ProcessorRepository

func ProcessorRepository() *util.ProcessorRepository {
	return repo
}

func addProcessor(processor util.Processable, name string, usage string) {
	if repo == nil { repo = &util.ProcessorRepository{} }
	
	repo.SetProcessor(name, processor)
	repo.SetUsage(name, usage)
}
