package config

import (
	"github.com/aki2o/esal/util"
)

var repo *util.ProcessorRepository

func ProcessorRepository() *util.ProcessorRepository {
	return repo
}

func RegistProcessor(processor_generator func() util.Processable, name string, description string, usage string) {
	if repo == nil { repo = &util.ProcessorRepository{} }
	
	repo.SetProcessorGenerator(name, processor_generator)
	repo.SetDescription(name, description)
	repo.SetUsage(name, usage)
}
