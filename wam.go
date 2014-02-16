// Copyright 2014 Unknown
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	rss "github.com/haarts/go-pkg-rss"
	"github.com/huichen/gobo"
)

const (
	TIME_OUT         = 30
	FETCH_INTERVAL   = 15
	PROCESS_INTERVAL = 5
)

type config struct {
	TimeOut         int    `json:"time_out"`
	FetchInterval   int    `json:"fetch_interval"`
	ProcessInterval int    `json:"process_interval"`
	AccessToken     string `json:"access_token"`
	AppKey          string `json:"app_key"`
}

var Cfg config

func init() {
	f, err := os.Open("conf.json")
	if err != nil {
		log.Fatalf("[ERROR] Fail to open config file: %s", err)
	}

	if err = json.NewDecoder(f).Decode(&Cfg); err != nil {
		log.Fatalf("[ERROR] Fail to decode config file: %s", err)
	}

	if Cfg.TimeOut == 0 {
		Cfg.TimeOut = TIME_OUT
	}
	if Cfg.FetchInterval == 0 {
		Cfg.FetchInterval = FETCH_INTERVAL
	}
	if Cfg.ProcessInterval == 0 {
		Cfg.ProcessInterval = PROCESS_INTERVAL
	}
}

var (
	Archived        = map[string]bool{}
	ArchivedMetions = make([]string, 0, 10)
	MsgQueue        = make([]interface{}, 0, 10)

	weibo = gobo.Weibo{}
)

func main() {
	go PollFeed("http://blog.golang.org/feed.atom", itemHandlerGoBlog)
	go PollFeed("http://blog.gopheracademy.com/feed.atom", itemHandlerGaBlog)
	go PollFeed("http://blog.go-china.org/feed.atom", itemHandlerGcBlog)
	go PollFeed("https://news.ycombinator.com/rss", itemHandlerHackerNews)
	go PollFeed("http://www.reddit.com/r/golang.rss", itemHandlerReddit)
	go FetchMentions()
	ProcessMsgQueue()
}

func PollFeed(uri string, itemHandler rss.ItemHandler) {
	feed := rss.New(Cfg.TimeOut, true, chanHandler, itemHandler)
	for {
		log.Println("Fetching from", uri, "...")
		if err := feed.Fetch(uri, nil); err != nil {
			log.Printf("[ERROR] %s: %s\n", uri, err)
			return
		}
		log.Println("Waiting updates from", uri, "...")
		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
	}
}

func genericItemHandler(feed *rss.Feed, ch *rss.Channel, newItems []*rss.Item, individualItemHandler func(*rss.Item)) {
	log.Printf("%d new item(s) in %s\n", len(newItems), feed.Url)
	for _, item := range newItems {
		individualItemHandler(item)
	}
}

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	//noop
}

func itemHandlerGoBlog(feed *rss.Feed, ch *rss.Channel, newItems []*rss.Item) {
	f := func(item *rss.Item) {
		short_title := item.Title
		if len(short_title) > 100 {
			short_title = short_title[:99] + "…"
		}
		log.Println(short_title + " " + item.Links[0].Href)
		MsgQueue = append(MsgQueue, short_title+" "+item.Links[0].Href)
	}

	if _, ok := Archived["go"]; !ok {
		Archived["go"] = false
	} else {
		genericItemHandler(feed, ch, newItems, f)
	}
}

func itemHandlerGaBlog(feed *rss.Feed, ch *rss.Channel, newItems []*rss.Item) {
	f := func(item *rss.Item) {
		short_title := item.Title
		if len(short_title) > 100 {
			short_title = short_title[:99] + "…"
		}
		log.Println(short_title + " " + item.Links[0].Href)
		MsgQueue = append(MsgQueue, short_title+" "+item.Links[0].Href)
	}

	if _, ok := Archived["ga"]; !ok {
		Archived["ga"] = false
	} else {
		genericItemHandler(feed, ch, newItems, f)
	}
}

func itemHandlerGcBlog(feed *rss.Feed, ch *rss.Channel, newItems []*rss.Item) {
	f := func(item *rss.Item) {
		short_title := item.Title
		if len(short_title) > 100 {
			short_title = short_title[:99] + "…"
		}
		log.Println(short_title + " " + item.Links[0].Href)
		MsgQueue = append(MsgQueue, short_title+" "+item.Links[0].Href)
	}

	if _, ok := Archived["gc"]; !ok {
		Archived["gc"] = false
	} else {
		genericItemHandler(feed, ch, newItems, f)
	}
}

func itemHandlerHackerNews(feed *rss.Feed, ch *rss.Channel, newItems []*rss.Item) {
	f := func(item *rss.Item) {
		if match, _ := regexp.MatchString(`\w Go( |$|\.)`, item.Title); match {
			short_title := item.Title
			if len(short_title) > 100 {
				short_title = short_title[:99] + "…"
			}
			log.Println(short_title + " " + item.Links[0].Href)
			MsgQueue = append(MsgQueue, short_title+" "+item.Links[0].Href)
		}
	}

	if _, ok := Archived["hn"]; !ok {
		Archived["hn"] = false
	} else {
		genericItemHandler(feed, ch, newItems, f)
	}
}

func itemHandlerReddit(feed *rss.Feed, ch *rss.Channel, newItems []*rss.Item) {
	f := func(item *rss.Item) {
		re := regexp.MustCompile(`([^"]+)">\[link\]`)
		matches := re.FindStringSubmatch(item.Description)
		if len(matches) == 2 {
			short_title := item.Title
			if len(short_title) > 100 {
				short_title = short_title[:99] + "…"
			}
			log.Println(short_title + " " + item.Links[0].Href)
			MsgQueue = append(MsgQueue, short_title+" "+item.Links[0].Href)
		}
	}

	if _, ok := Archived["reddit"]; !ok {
		Archived["reddit"] = false
	} else {
		genericItemHandler(feed, ch, newItems, f)
	}
}

func PostWeibo(content string) {
	params := gobo.Params{"source": Cfg.AppKey, "status": "#golang# " + content}
	if err := weibo.Call("statuses/update", "post", Cfg.AccessToken, params, nil); err != nil {
		log.Printf("[ERROR] PostWeibo: %s\n", err)
	}
}

func isMehtionExist(id string) bool {
	for _, str := range ArchivedMetions {
		if str == id {
			return true
		}
	}
	ArchivedMetions = append(ArchivedMetions, id)
	return false
}

func RepostWeibo(id int64) {
	params := gobo.Params{"id": id}
	if err := weibo.Call("statuses/repost", "post", Cfg.AccessToken, params, nil); err != nil {
		log.Printf("[ERROR] Repost: %s\n", err)
	}
}

func FetchMentions() {
	for {
		log.Println("Fetching mention list...")
		// Fetch list of mentions.
		var statuses gobo.Statuses
		if err := weibo.Call("statuses/mentions", "get", Cfg.AccessToken, nil, &statuses); err != nil {
			log.Printf("[ERROR] FetchMentions: %s\n", err)
		}

		// Filter original mentions.
		for _, status := range statuses.Statuses {
			if status.Retweeted_Status == nil &&
				strings.Contains(status.Text, "#golang#") {
				if _, ok := Archived["mention"]; !ok {
					continue
				} else if isMehtionExist(fmt.Sprint(status.Id)) {
					continue
				}
				log.Printf("Mention: %s\n", status.Text)
				MsgQueue = append(MsgQueue, status.Id)
			}
		}
		Archived["mention"] = false

		log.Println("Waiting for new mention...")
		time.Sleep(time.Duration(Cfg.FetchInterval) * time.Minute)
	}
}

func ProcessMsgQueue() {
	for {
		log.Println("Processing the request...")
		if len(MsgQueue) > 0 {
			switch MsgQueue[0].(type) {
			case string:
				content := MsgQueue[0].(string)
				PostWeibo(content)
			case int64:
				id := MsgQueue[0].(int64)
				RepostWeibo(id)
			}
			MsgQueue = MsgQueue[1:]
		}
		log.Printf("Waiting for process next request... %d left\n", len(MsgQueue))
		time.Sleep(time.Duration(Cfg.ProcessInterval) * time.Minute)
	}
}
