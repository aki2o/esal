package action

import (
	"flag"
	"net/url"
	"strconv"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/esa-cui/util"
	"github.com/aki2o/esa-cui/config"
)

type sync struct {}

func init() {
	processors["sync"] = &sync{}
}

func (self *sync) SetOption(flagset *flag.FlagSet) {
}

func (self *sync) Do(args []string) error {
	for _, query_config := range config.Current.Queries {
		fetched_count	:= 0
		total_count		:= 1
		page_index		:= 1
		
		log.Info("start to fetch "+Context.Team+":"+Context.Root())
		for fetched_count < total_count {
			query := url.Values{}

			for _, entry := range query_config.Entries {
				query.Add(entry.Key, entry.Value)
			}
			
			query.Add("page", strconv.Itoa(page_index))
			query.Add("per_page", "100")
			
			log.WithFields(log.Fields{ "page": page_index, "total_count": total_count }).Debug("get post")
			res, err := Context.Client.Post.GetPosts(Context.Team, query)
			if err != nil { return err }
			log.WithFields(log.Fields{ "next_page": res.NextPage }).Debug("got post")

			if err := util.EnsureDir(Context.Root()); err != nil { return err }

			for _, post := range res.Posts {
				SavePost(&post)
				fetched_count += 1
			}

			page_index += 1
			total_count = res.TotalCount
		}
	}

	return nil
}
