package action

import (
	"net/url"
	"strconv"
	"time"
	"fmt"
	"os"
	"io"
	"strings"
	log "github.com/sirupsen/logrus"
	"gopkg.in/cheggaaa/pb.v1"
	"github.com/aki2o/esal/util"
	"github.com/aki2o/esal/config"
)

type sync struct {
	util.ProcessIO
	pecoable
	// AllRequired bool `short:"a" long:"all" description:"Exec for all queries."`
	Force bool `short:"f" long:"force" description:"Exec with ignore last synchronized time."`
	ByNumber bool `short:"n" long:"number" description:"Exec for only posts of numbers given as arguments."`
	Quiet bool `short:"q" long:"quiet" description:"Exec quietly."`
	RefreshedCategory string `short:"R" long:"refresh" description:"Remove category before exec." value-name:"CATEGORY"`
	progress_bars map[string]*pb.ProgressBar
	found_tags []string
}

func init() {
	RegistProcessor(func() util.Processable { return &sync{} }, "sync", "Fetch posts.", "[OPTIONS] QUERY_OR_POST...")
}

func (self *sync) Do(args []string) error {
	// if len(args) == 0 && ! self.AllRequired && self.PecoRequired() {
	// 	var err error
	// 	if self.ByNumber {
	// 		args, err = selectNodeByPeco("", false)
	// 	} else {
	// 		args, err = self.selectQuery()
	// 	}
	// 	if err != nil { return err }
	// }
	
	// if len(args) == 0 && !self.AllRequired {
	// 	return errors.New("Require query name!")
	// }
	if len(args) == 0 && !self.ByNumber {
		for _, query := range config.Team.Queries {
			args = append(args, query.Name)
		}
	}
	
	self.found_tags = []string{}

	if self.RefreshedCategory != "" {
		if err := self.Refresh(self.RefreshedCategory); err != nil {
			return err
		}
	}
	
	var err error
	if self.ByNumber {
		err = self.DoByNumber(args)
	} else {
		err = self.DoByQuery(args)
	}

	self.saveTags()
	
	return err
}

func(self *sync) Refresh(path string) error {
	find_process := &find{}
	find_process.Type = "u"
	node_paths, err := find_process.collectNodesIn(path)
	if err != nil { return err }

	for _, node_path := range node_paths {
		_, post_number := DirectoryPathAndPostNumberOf(node_path)
		
		err := DeletePostData(post_number)
		if err == nil { err = DeletePostBody(post_number) }
		if err != nil { return err }
	}

	return os.RemoveAll(PhysicalPathOf(path))
}

func (self *sync) DoByNumber(args []string) error {
	for _, path := range args {
		_, post_number := DirectoryPathAndPostNumberOf(path)
		if post_number == "" {
			fmt.Fprintf(os.Stderr, "Unknown post number of '%s'!", path)
			continue
		}

		post_number_as_int, _ := strconv.Atoi(post_number)
		post, err := Context.Client.Post.GetPost(Context.Team, post_number_as_int)
		if err != nil {
			log.WithFields(log.Fields{ "path": path, "number": post_number, "error": err.Error() }).Error("failed to fetch post")
			fmt.Fprintf(os.Stderr, "Failed to fetch post '%s' : %s\n", path, err.Error())
			continue
		}
		
		if err = SavePost(post); err != nil {
			log.WithFields(log.Fields{ "number": post.Number, "full_name": post.FullName, "error": err.Error() }).Error("failed to save post")
			fmt.Fprintf(os.Stderr, "Failed to save post '%d: %s' : %s\n", post.Number, post.FullName, err.Error())
			continue
		}

		self.found_tags = append(self.found_tags, post.Tags...)
	}
	return nil
}

func (self *sync) DoByQuery(args []string) error {
	pool, err := self.setupProgressBars(args)
	if err != nil { return err }
	if pool != nil { defer pool.Stop() }
	
	query_configs := make([]config.Query, len(config.Team.Queries))
	
	for index, query_config := range config.Team.Queries {
		if self.isTarget(query_config, args) {
			if self.processQuery(query_config) {
				query_config.SynchronizedAt = time.Now()
			}
		}

		query_configs[index] = query_config
	}

	config.Team.Queries = query_configs
	config.Save()
	return nil
}

func (self *sync) isTarget(query config.Query, args []string) bool {
	// if self.AllRequired { return true }
	
	for _, query_name := range args {
		if query_name == query.Name {
			return true
		}
	}
	
	return false
}

func (self *sync) setupProgressBars(args []string) (*pb.Pool, error) {
	if self.Quiet { return nil, nil }
	
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

func (self *sync) processQuery(query_config config.Query) bool {
	fetched_count	:= 0
	total_count		:= 1
	page_index		:= 1
	progress_bar	:= self.getProgressBar(query_config.Name)
	success			:= true
	
	if progress_bar != nil { defer progress_bar.Finish() }
	
	log.Info("start to fetch post "+Context.Team+":"+query_config.Name)
	for fetched_count < total_count {
		query := url.Values{}

		for _, entry := range query_config.Entries {
			query.Add(entry.Key, entry.Value)
		}
		
		query.Add("page", strconv.Itoa(page_index))
		query.Add("per_page", "100")
		if ! self.Force {
			// 当日以降に更新されたものを取得するためには、esaのドキュメントには指定日以降と記述されているが、
			// 記号が不等号から判断すると、前日を指定しないとダメっぽい。
			// さらに、実際は前日でもダメで前々日を指定しないといけないが、これはesa APIのバグっぽい。TZが考慮できてないとかなのかな
			prev_day_duration, _ := time.ParseDuration("-48h")
			
			query.Add("updated", ">"+query_config.SynchronizedAt.Add(prev_day_duration).Format("2006-01-02"))
		}
		if self.RefreshedCategory != "" {
			category := CategoryOf(PhysicalPathOf(self.RefreshedCategory))
			if category != "" {
				query.Add("in", category)
			}
		}
		
		log.WithFields(log.Fields{ "page": page_index, "fetched_count": fetched_count }).Debug("get post")
		res, err := Context.Client.Post.GetPosts(Context.Team, query)
		if err != nil {
			log.WithFields(log.Fields{ "page": page_index, "fetched_count": fetched_count, "error": err.Error() }).Error("failed to fetch post")
			fmt.Fprintf(os.Stderr, "Failed to fetch posts : %s\n", err.Error())
			return false
		}
		log.WithFields(log.Fields{ "next_page": res.NextPage }).Debug("got post")

		if err := util.EnsureDir(Context.Root()); err != nil {
			log.WithFields(log.Fields{ "path": Context.Root(), "error": err.Error() }).Error("failed to ensure dir")
			fmt.Fprintf(os.Stderr, "Failed to ensure local directory : %s\n", err.Error())
			return false
		}

		if total_count == 1 {
			log.WithFields(log.Fields{ "total_count": res.TotalCount }).Debug("set total")
			
			total_count = res.TotalCount
			
			if progress_bar != nil { progress_bar.Total = int64(total_count) }
		}
		
		for _, post := range res.Posts {
			fetched_count += 1
			
			if util.Exists(GetPostLockPath(strconv.Itoa(post.Number))) {
				log.WithFields(log.Fields{ "number": post.Number }).Info("skip locked post")
				fmt.Fprintf(os.Stderr, "Skip a locked post '%d: %s'\n", post.Number, post.FullName)
				continue
			}
			
			if err = SavePost(&post); err != nil {
				log.WithFields(log.Fields{ "number": post.Number, "full_name": post.FullName, "error": err.Error() }).Error("failed to save post")
				fmt.Fprintf(os.Stderr, "Failed to save post '%d: %s' : %s\n", post.Number, post.FullName, err.Error())
				success = false
			}
			
			self.found_tags = append(self.found_tags, post.Tags...)
			
			if progress_bar != nil { progress_bar.Increment() }
		}

		page_index += 1
	}

	log.Info("finished to fetch post "+Context.Team+":"+query_config.Name)
	return success
}

func (self *sync) saveTags() {
	tag_process := &tag{}
	err := tag_process.AddTags(self.found_tags)
	if err != nil {
		log.WithFields(log.Fields{ "tags": self.found_tags, "error": err.Error() }).Error("failed to save tags")
	}
}

func (self *sync) selectQuery() ([]string, error) {
	provider := func(writer *io.PipeWriter) {
		defer writer.Close()

		for _, query := range config.Team.Queries {
			fmt.Fprintln(writer, query.Name)
		}
	}

	selected, _, err := pipePeco(provider, "Query")
	if err != nil { return []string{}, err }

	return strings.Split(selected, "\n"), nil
}
