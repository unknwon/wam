Weibo Automated Management
===

wam(Weibo Automated Management) - [Golang中国微博](http://weibo.com/u/3211200050)自动化管理程序。

## 开发目的

参考 Twitter 上一个 Golang 新闻帐号的做法，采用自动抓取全网 Golang 更新的程序进行帐号管理。不过本程序在其基础上增加了对用户转发诉求的特性，支持每天定量转发，以助社区产品推广。

## 功能特性

- 抓取 [HackerNews](http://news.ycombinator.com/)、[Reddit](http://www.reddit.com/r/golang)、[Gopher Academy Blog](http://blog.gopheracademy.com/) 和 [Go 官方博客](http://blog.golang.org/) 上有关 Go 语言的新闻更新。
- 接受来自用户的 @ 转发诉求，转发内容中必须包含 **#golang#** 标签，否则无视。
- 定时发布抓取到的新闻或转发诉求（未授权应用每天限量 50 条）。

## 特别鸣谢

- 推特上的 Golang News [抓取程序](https://github.com/haarts/golang_news)。

