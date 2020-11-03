package vinscraper

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/Monstercat/golib/request"
)

var (
	ErrRedditNoChildren = errors.New("no children in reddit")
)

const (
	SourceRedditPost = "reddit_post"
	SourceRedditComment = "reddit_comment"
)

var redditUrlRegexp = "\\/comments\\/([a-zA-Z0-9]+)\\/?[[a-zA-Z0-9\\_]+?\\/([a-zA-Z0-9]+)?"

type RedditPostInfo struct {
	RedditThing
	Title string `json:"title"`
}

type RedditCommentInfo struct {
	RedditThing
}

// Comment or Post
type RedditThing struct {
	Author            string        `json:"author"`
	Body              string        `json:"body"`
	Created           float64       `json:"created"`
	Crossposts        []RedditThing `json:"crosspost_parent_list"`
	Permalink         string        `json:"permalink"`
	Subreddit         string        `json:"subreddit"`
	SubredditPrefixed string        `json:"subreddit_name_prefixed"`
	LinkId            string        `json:"link_id"`
	Spoiler           bool          `json:"spoiler"`
	URL               string        `json:"url"`
}

// Implements ScrapeMeta
type RedditMeta struct {
	Subreddit string // eg: "AskReddit"
}

type RedditScraper struct {
	UserAgent string
}

func (rs *RedditScraper) WantsURL(link string) bool {
	u, _ := url.Parse(link)
	host := strings.ToLower(u.Hostname())
	if !strings.HasSuffix(host, "reddit.com") {
		return false
	}
	r, err := regexp.Compile(redditUrlRegexp)
	if err != nil {
		return false
	}
	result := r.FindStringSubmatch(link)

	if len(result) < 3 {
		return false
	}
	return true
}

func (rs *RedditScraper) Scrape(urlS string) (*ScrapeInfo, error) {
	r, err := regexp.Compile(redditUrlRegexp)
	if err != nil {
		return nil, err
	}
	result := r.FindStringSubmatch(urlS)

	if result[2] == "" {
		return rs.ScrapePost(result[1])
	}
	return nil, errors.New("do comment next")
	return rs.ScrapeComment(result[2])
}

func (rs *RedditScraper) ScrapePost(postId string) (*ScrapeInfo, error) {
	src := &RedditPostSource{
		ID: postId,
	}
	info, err := src.GetPostInfo(rs)
	if err != nil {
		return nil, err
	}

	item := &ScrapeInfo{
		CreditTitle: info.Author,
		CreditURL:   getRedditAuthorURL(info.Author),
		Meta: &info.RedditThing,
		SourceType: SourceRedditPost,
		Title:       info.Title,
	}

	if IsImageLink(info.URL) {
		item.ThumbnailSource = info.URL
	}

	return item, nil
}

func (rs *RedditScraper) ScrapeComment(postId string) (*ScrapeInfo, error) {
	src := &RedditCommentSource{
		ID: postId,
	}
	info, err := src.GetCommentInfo(rs)
	if err != nil {
		return nil, err
	}
	return &ScrapeInfo{
		Title:       "Comment by " + info.Author,
		CreditTitle: info.Author,
		CreditURL:   getRedditAuthorURL(info.Author),
	}, nil
}

type RedditCommentInfoResponse struct {
	Data struct {
		Children []struct {
			Data RedditCommentInfo `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type RedditPostInfoResponse struct {
	Data struct {
		Children []struct {
			Data RedditPostInfo `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type RedditPostSource struct {
	ID string
}

type RedditCommentSource struct {
	ID string
}

func (p *RedditPostSource) GetKey() string {
	return p.ID
}

func (p *RedditPostSource) GetPostInfo(rs *RedditScraper) (*RedditPostInfo, error) {
	var body RedditPostInfoResponse
	params := request.Params{
		Url: "https://api.reddit.com/api/info?id=t3_" + p.ID,
	}
	if err := rs.RedditRequest(&params, nil, &body); err != nil {
		return nil, err
	}
	if len(body.Data.Children) == 0 {
		return nil, ErrRedditNoChildren
	}
	dat := body.Data.Children[0].Data
	return &dat, nil
}

func (p *RedditCommentSource) GetKey() string {
	return p.ID
}

func (rs *RedditScraper) RedditRequest(params *request.Params, payload interface{}, body interface{}) error {
	if params.Headers == nil {
		params.Headers = make(map[string]string)
	}
	params.Headers["User-agent"] = rs.UserAgent
	return request.Request(params, payload, body)
}

func (p *RedditCommentSource) GetCommentInfo(rs *RedditScraper) (*RedditCommentInfo, error) {
	var body RedditCommentInfoResponse
	params := request.Params{
		Url: "https://api.reddit.com/api/info?id=t1_" + p.ID,
	}
	if err := rs.RedditRequest(&params, nil, &body); err != nil {
		return nil, err
	}
	if len(body.Data.Children) == 0 {
		return nil, ErrRedditNoChildren
	}
	dat := body.Data.Children[0].Data
	return &dat, nil
}

func getRedditAuthorURL(username string) string {
	return "https://www.reddit.com/u/" + username
}
