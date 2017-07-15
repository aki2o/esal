package action

import (
	"flag"
	"errors"
	"net/url"
	"strconv"
	"time"
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"
	"gopkg.in/cheggaaa/pb.v1"
	"github.com/aki2o/esa-cui/util"
	"github.com/aki2o/esa-cui/config"
)

type sync struct {
	all bool
	force bool
	progress_bars map[string]*pb.ProgressBar
}

func init() {
	addProcessor(&sync{}, "sync", "Fetch posts.")
}

func (self *sync) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.all, "a", false, "Exec for all config.")
	flagset.BoolVar(&self.force, "f", false, "Exec with ignore last synchronized time.")
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
		bar.SetRefreshRate(time.Millisecond * 100)
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
	
	log.Info("start to fetch post "+Context.Team+":"+query_config.Name)
	for fetched_count < total_count {
		query := url.Values{}

		for _, entry := range query_config.Entries {
			query.Add(entry.Key, entry.Value)
		}
		
		query.Add("page", strconv.Itoa(page_index))
		query.Add("per_page", "100")
		if ! self.force {
			// 当日以降に更新されたものを取得するためには、esaのドキュメントには指定日以降と記述されているが、
			// 記号が不等号から判断すると、前日を指定しないとダメっぽい。
			// さらに、実際は前日でもダメで前々日を指定しないといけないが、これはesa APIのバグっぽい。TZが考慮できてないとかなのかな
			prev_day_duration, _ := time.ParseDuration("-48h")
			
			query.Add("updated", ">"+query_config.SynchronizedAt.Add(prev_day_duration).Format("2006-01-02"))
		}
		
		log.WithFields(log.Fields{ "page": page_index, "fetched_count": fetched_count }).Debug("get post")
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
			
			if util.Exists(GetPostLockPath(strconv.Itoa(post.Number))) {
				log.WithFields(log.Fields{ "number": post.Number }).Info("skip locked post")
				fmt.Printf("Skip a locked post %d: %s\n", post.Number, post.FullName)
				continue
			}
			
			if err = SavePost(&post); err != nil {
				log.WithFields(log.Fields{ "number": post.Number, "full_name": post.FullName, "error": err.Error() }).Error("failed to save post")
				fmt.Fprintf(os.Stderr, "Failed to save post %d: %s\n", post.Number, post.FullName)
			}
			progress_bar.Increment()
		}

		page_index += 1
	}

	log.Info("finished to fetch post "+Context.Team+":"+query_config.Name)
	return nil
}
