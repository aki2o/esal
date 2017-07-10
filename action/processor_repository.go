package action

import (
	"github.com/aki2o/esa-cui/util"
)

var processors = make(map[string]util.Processable)

func NewProcessorRepository() *util.ProcessorRepository {
	repo := &util.ProcessorRepository{}

	for name, processor := range processors {
		repo.SetProcessor(name, processor)
	}
	
	return repo
}
