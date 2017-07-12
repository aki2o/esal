package action

import (
	"flag"
	"errors"
	"net/url"
	"strconv"
	"time"
	log "github.com/sirupsen/logrus"
	"gopkg.in/cheggaaa/pb.v1"
	"github.com/aki2o/esa-cui/util"
	"github.com/aki2o/esa-cui/config"
)

type sync struct {
	all bool
	progress_bars map[string]*pb.ProgressBar
}

func init() {
	addProcessor(&sync{}, "sync", "Fetch posts.")
}

func (self *sync) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.all, "a", false, "Exec for all config.")
}

func (self *sync) Do(args []string) error {
	if len(args) == 0 && !self.all {
		return errors.New("Require query name!")
	}

	pool, err := self.setupProgressBars(args)
	if err != nil { return err }
	defer pool.Stop()
	
	query_configs := make([]config.Query, len(config.Team.Queries))
	
	for index, query_config := range config.Team.Queries {
		if !self.isTarget(query_config, args) { continue }

		err := self.processQuery(query_config)
		if err == nil {
			query_config.SynchronizedAt = time.Now()
		} else {
			util.PutError(err)
		}
		
		query_configs[index] = query_config
	}

	config.Team.Queries = query_configs
	config.Save()
	return nil
}

func (self *sync) isTarget(query config.Query, args []string) bool {
	if self.all { return true }
	
	for _, query_name := range args {
		if query_name == query.Name {
			return true
		}
	}
	
	return false
}

func (self *sync) setupProgressBars(args []string) (*pb.Pool, error) {
	self.progress_bars = make(map[string]*pb.ProgressBar)
	var bars = []*pb.ProgressBar{}
	
	for _, query_config := range config.Team.Queries {
		if ! self.isTarget(query_config, args) { continue }

		bar := pb.New(0).Prefix(query_config.Name)
		// bar.SetRefreshRate(time.Second)
		bar.ShowCounters = true
		bar.ShowTimeLeft = false
		bar.ShowSpeed = false
		bar.SetWidth(80)
		bar.SetMaxWidth(80)
		bar.Start()

		self.progress_bars[query_config.Name] = bar
		bars = append(bars, bar)
	}

	return pb.StartPool(bars...)
}

func (self *sync) getProgressBar(name string) *pb.ProgressBar {
	return self.progress_bars[name]
}

func (self *sync) processQuery(query_config config.Query) error {
	fetched_count	:= 0
	total_count		:= 1
	page_index		:= 1
	progress_bar	:= self.getProgressBar(query_config.Name)
	
	defer progress_bar.Finish()
	
	log.Info("start to fetch "+Context.Team+":"+query_config.Name)
	for fetched_count < total_count {
		query := url.Values{}

		for _, entry := range query_config.Entries {
			query.Add(entry.Key, entry.Value)
		}
		
		query.Add("page", strconv.Itoa(page_index))
		query.Add("per_page", "100")
		query.Add("updated", ">"+query_config.SynchronizedAt.Format("2006-01-02"))
		
		log.WithFields(log.Fields{ "page": page_index, "total_count": total_count }).Debug("get post")
		res, err := Context.Client.Post.GetPosts(Context.Team, query)
		if err != nil { return err }
		log.WithFields(log.Fields{ "next_page": res.NextPage }).Debug("got post")

		if err := util.EnsureDir(Context.Root()); err != nil { return err }

		if total_count == 1 {
			log.WithFields(log.Fields{ "total_count": res.TotalCount }).Debug("set total")
			
			total_count			= res.TotalCount
			progress_bar.Total	= int64(total_count)
		}
		
		for _, post := range res.Posts {
			fetched_count += 1
			
			if util.Exists(GetLocalPostPath(post.Category, strconv.Itoa(post.Number), "lock")) { continue }
			
			SavePost(&post)
			progress_bar.Increment()
		}

		page_index += 1
	}

	log.Info("finished to fetch "+Context.Team+":"+query_config.Name)
	return nil
}
