package action

import (
	"regexp"
	"strconv"
	"time"
	"fmt"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type matchable struct {
	Type string `short:"t" long:"type" description:"Node type ident. one of c(ategory), p(ost), w(ip), s(hipped), l(ocked), u(nlocked)." value-name:"IDENT"`
	NamePattern string `short:"n" long:"name" description:"Node name pattern." value-name:"PATTERN"`
	Tags []string `short:"T" long:"tag" description:"Tag name labeling tha post." value-name:"TAG"`
	TagsByPecoRequired bool `short:"t" long:"tagp" description:"Choice tags by peco."`
	CreateUsers []string `short:"C" long:"cuser" description:"User screen name create the post." value-name:"USER"`
	CreateUsersByPecoRequired bool `short:"c" long:"cuserp" description:"Choice user screen name by peco."`
	UpdateUsers []string `short:"U" long:"uuser" description:"User screen name last update the post." value-name:"USER"`
	UpdateUsersByPecoRequired bool `short:"u" long:"uuserp" description:"Choice user screen name by peco."`
	CreateTimeCondition string `short:"t" long:"ctime" description:"Created time condition (same Linux find command) of the post." value-name:"VALUE"`
	UpdateTimeCondition string `short:"T" long:"utime" description:"Updated time condition (same Linux find command) of the post." value-name:"VALUE"`
	name_re *regexp.Regexp
}

func (self *matchable) matchCategory(name string) bool {
	if self.Type != "" && self.Type != "c" { return false }
	if ! self.matchName(name) { return false }
	
	return true
}

func (self *matchable) matchPost(post *esa.PostResponse) bool {
	if self.Type == "c" { return false }
	if self.Type == "w" && ! post.Wip { return false }
	if self.Type == "s" && post.Wip { return false }
	if self.Type == "l" && ! util.Exists(GetPostLockPath(strconv.Itoa(post.Number))) { return false }
	if self.Type == "u" && util.Exists(GetPostLockPath(strconv.Itoa(post.Number))) { return false }
	if ! self.matchName(post.Name) { return false }
	if ! self.matchUser(post) { return false }
	if ! self.matchTime(post) { return false }
	if ! self.matchTag(post) { return false }
	
	return true
}

func (self *matchable) matchName(name string) bool {
	if self.NamePattern == "" { return true }
	
	if self.name_re == nil {
		re, _ := regexp.Compile(self.NamePattern)
		self.name_re = re
	}

	return self.name_re.MatchString(name)
}

func (self *matchable) matchUser(post *esa.PostResponse) bool {
	if ! self.matchCreateUser(post) { return false }
	if ! self.matchUpdateUser(post) { return false }

	return true
}

func (self *matchable) matchCreateUser(post *esa.PostResponse) bool {
	if len(self.CreateUsers) == 0 && self.CreateUsersByPecoRequired {
		if users, err := selectUserByPeco("create user"); err == nil { self.CreateUsers = users }
		
		self.CreateUsersByPecoRequired = false
	}

	if len(self.CreateUsers) == 0 { return true }
	
	for _, user := range self.CreateUsers {
		if user == post.CreatedBy.ScreenName { return true }
	}

	return false
}

func (self *matchable) matchUpdateUser(post *esa.PostResponse) bool {
	if len(self.UpdateUsers) == 0 && self.UpdateUsersByPecoRequired {
		if users, err := selectUserByPeco("update user"); err == nil { self.UpdateUsers = users }
		
		self.UpdateUsersByPecoRequired = false
	}

	if len(self.UpdateUsers) == 0 { return true }
	
	for _, user := range self.UpdateUsers {
		if user == post.UpdatedBy.ScreenName { return true }
	}

	return false
}

func (self *matchable) matchTime(post *esa.PostResponse) bool {
	if ! self.matchCreatedTime(post) { return false }
	if ! self.matchUpdatedTime(post) { return false }

	return true
}

func (self *matchable) matchCreatedTime(post *esa.PostResponse) bool {
	if self.CreateTimeCondition == "" { return true }

	return self.matchTimeCondition(self.CreateTimeCondition, post.CreatedAt)
}

func (self *matchable) matchUpdatedTime(post *esa.PostResponse) bool {
	if self.UpdateTimeCondition == "" { return true }
	
	return self.matchTimeCondition(self.UpdateTimeCondition, post.UpdatedAt)
}

func (self *matchable) matchTimeCondition(cond string, tested_time_str string) bool {
	number_of_date_value := ""
	if len(cond) == 1 { number_of_date_value = string(cond[0]) }
	if len(cond) == 2 { number_of_date_value = string(cond[1]) }
	if number_of_date_value == "" { return false }
	number_of_date, err := strconv.Atoi(number_of_date_value)
	if err != nil { return false }

	range_type := "in"
	if len(cond) == 2 {
		switch string(cond[0]) {
		case "-":
			range_type = "since"
		case "+":
			range_type = "ago"
		default:
			return false
		}
	}

	tested_time, _ := time.Parse("2006-01-02T15:04:05-07:00", tested_time_str)
	
	var beginning_of_duration time.Duration
	var end_of_duration time.Duration
	switch range_type {
	case "in":
		beginning_of_duration, _	= time.ParseDuration(fmt.Sprintf("-%dh", number_of_date * 24))
		end_of_duration, _			= time.ParseDuration(fmt.Sprintf("-%dh", (number_of_date + 1) * 24))
	case "since":
		beginning_of_duration, _	= time.ParseDuration("-0h")
		end_of_duration, _			= time.ParseDuration(fmt.Sprintf("-%dh", number_of_date * 24))
	case "ago":
		beginning_of_duration, _	= time.ParseDuration(fmt.Sprintf("-%dh", number_of_date * 24))
	}

	curr_time := time.Now()
	if tested_time.After(curr_time.Add(beginning_of_duration)) { return false }
	
	if end_of_duration != 0 {
		if tested_time.Before(curr_time.Add(end_of_duration)) { return false }
		if tested_time.Equal(curr_time.Add(end_of_duration)) { return false }
	}

	return true
}

func (self *matchable) matchTag(post *esa.PostResponse) bool {
	if len(self.Tags) == 0 && self.TagsByPecoRequired {
		if tags, err := selectTagByPeco(); err == nil { self.Tags = tags }
		
		self.TagsByPecoRequired = false
	}

	if len(self.Tags) == 0 { return true }

	for _, tag := range self.Tags {
		found := false
		for _, post_tag := range post.Tags {
			if tag == post_tag {
				found = true
				break
			}
		}
		if ! found { return false }
	}

	return true
}
