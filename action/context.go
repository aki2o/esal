package action

import (
	"errors"
	"path/filepath"
	log "github.com/sirupsen/logrus"
	"github.com/upamune/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

type EsaCuiActionContext struct {
	local_strage_path string
	Team              string
	Cwd               string
	Client            *esa.Client
}

func (c *EsaCuiActionContext) Root() string {
	return filepath.Join(c.local_strage_path, c.Team)
}

var Context *EsaCuiActionContext

func SetupContext(team string, access_token string) error {
	Context = &EsaCuiActionContext{}

	if team == "" {
		return errors.New("Invalid Team!")
	}
	
	Context.local_strage_path	= filepath.Join(util.LocalRootPath(), "posts")
	Context.Team				= team
	Context.Cwd					= Context.Root()
	Context.Client				= esa.NewClient(access_token)
	
	log.WithFields(log.Fields{ "team": Context.Team, "cwd": Context.Cwd }).Debug("setup Context")
	
	return nil
}
