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
	"log"
	"regexp"
	"time"

	rss "github.com/haarts/go-pkg-rss"
	"github.com/huichen/gobo"
)

const TIME_OUT = 30

var (
	Archived = map[string]bool{}
	WaitList = []string{"#golang #自动化管理微博测试，目前还处于阳春阶段 http://golang.org"}

	weibo       = gobo.Weibo{}
	AccessToken = "2.00Mpr_VDkZyzfB64ba8a1c1bPLPVmC"
	AppKey      = "1536733472"
)

func main() {
	go PollFeed("http://blog.golang.org/feed.atom", itemHandlerGoBlog)
	go PollFeed("http://blog.gopheracademy.com/feed.atom", itemHandlerGaBlog)
	go PollFeed("https://news.ycombinator.com/rss", itemHandlerHackerNews)
	PollFeed("http://www.reddit.com/r/golang.rss", itemHandlerReddit)
}

func PollFeed(uri string, itemHandler rss.ItemHandler) {
	feed := rss.New(TIME_OUT, true, chanHandler, itemHandler)
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
		PostWeibo(short_title + " " + item.Links[0].Href)
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
		PostWeibo(short_title + " " + item.Links[0].Href)
	}

	if _, ok := Archived["ga"]; !ok {
		Archived["ga"] = false
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
			PostWeibo(short_title + " " + item.Links[0].Href)
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
			PostWeibo(short_title + " " + item.Links[0].Href)
		}
	}

	if _, ok := Archived["reddit"]; !ok {
		Archived["reddit"] = false
	} else {
		genericItemHandler(feed, ch, newItems, f)
	}
}

func PostWeibo(content string) {
	var statuses gobo.Statuses
	params := gobo.Params{"source": AppKey, "status": "#golang# " + content}
	if err := weibo.Call("statuses/update", "post", AccessToken, params, &statuses); err != nil {
		log.Printf("[ERROR] PostWeibo: %s\n", err)
	}
	for _, status := range statuses.Statuses {
		log.Println(status.Text)
	}
}
