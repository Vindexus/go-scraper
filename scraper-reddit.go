package vinscraper

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/monstercat/golib/request"
)

var (
	ErrRedditNoChildren = errors.New("no children in reddit")
)

const (
	SourceRedditPost    = "reddit_post"
	SourceRedditComment = "reddit_comment"
)

var redditUrlRegexp = "\\/comments\\/([%a-zA-Z0-9]+)\\/?[[%a-zA-Z0-9\\_]+?\\/([%a-zA-Z0-9]+)?"

// Used as return data, can be our own structure
type RedditThingMeta struct {
	Author            string
	Body              string
	Created           float64
	Id                string
	Permalink         string
	Subreddit         string
	SubredditPrefixed string
	URL               string
}

type RedditPostMeta struct {
	RedditThingMeta
	Crossposts []RedditThing
	Spoiler    bool
	URL        string
}
type RedditCommentMeta struct {
	RedditThingMeta
	IsSubmitter bool
}
type RedditMediaMetadata struct {
	Status string
	Source struct {
		Width  int    `json:"x"`
		Height int    `json:"y"`
		URL    string `json:"u"`
	} `json:"s"`
}

// Comment or Posts are "things" to reddit
// this struct has the properties that will appear in both
// requests for Posts and requests for comments
type RedditThing struct {
	Author            string  `json:"author"`
	Body              string  `json:"body"`
	Created           float64 `json:"created"`
	Id                string  `json:"id"` // id of comment or post
	Permalink         string  `json:"permalink"`
	Subreddit         string  `json:"subreddit"`
	SubredditPrefixed string  `json:"subreddit_name_prefixed"`
}

type RedditGalleryData struct {
	Items []struct {
		Id      int    `json:"id"`
		MediaId string `json:"media_id"`
	} `json:"items"`
}

// Used to fetch data from the API, must match reddit's structure
type RedditPostInfo struct {
	RedditThing

	// These fields are in posts, but not comments
	Crossposts    []RedditThing                  `json:"crosspost_parent_list"`
	GalleryData   RedditGalleryData              `json:"gallery_data"`
	MediaMetadata map[string]RedditMediaMetadata `json:"media_metadata"`
	Spoiler       bool                           `json:"spoiler"`
	Title         string                         `json:"title"`
	URL           string                         `json:"url"`
}

type RedditCommentInfo struct {
	RedditThing

	// These fields are in comments, but not posts
	IsSubmitter bool   `json:"is_submitter"`
	LinkId      string `json:"link_id"`
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

	var info *ScrapeInfo

	// Reddit urls look kind of like /r/subreddit/1235234/comments when it's a link to a post
	// and a link to a comment will append the comment id
	// So if no second id is found, then this is a post link and we scrape a post
	if result[2] == "" {
		info, err = rs.ScrapePost(result[1])
		if err != nil {
			return nil, err
		}

		if IsImageLink(info.URL) {
			info.ThumbnailSources = []string{info.URL}
		}
	} else {
		info, err = rs.ScrapeComment(result[2])
		if err != nil {
			return nil, err
		}
	}

	info.URL = urlS
	return info, nil
}

func (rs *RedditScraper) ScrapePost(postId string) (*ScrapeInfo, error) {
	src := &RedditPostSource{
		ID: postId,
	}
	info, err := src.GetPostInfo(rs)
	if err != nil {
		return nil, err
	}

	result := info.BasicScrapeInfo()
	result.SourceType = SourceRedditPost
	result.Title = info.Title
	result.URL = info.URL

	result.Meta = &RedditPostMeta{
		Crossposts:      info.Crossposts,
		RedditThingMeta: info.RedditThing.ToMeta(),
		Spoiler:         info.Spoiler,
		URL:             info.URL,
	}

	result.ThumbnailSources = make([]string, 0)
	// If metadata has items in it then this reddit post is a gallery
	if len(info.MediaMetadata) > 0 {
		mediaThumbs := map[string]string{}

		// This data is always sorted randomly
		// Golang automatically randomizes JSON key order because the order
		// can't be guaranteed
		for k, v := range info.MediaMetadata {
			thumb := strings.ReplaceAll(v.Source.URL, "&amp;", "&")
			mediaThumbs[k] = thumb
			// For some reason reddit does this encoding to their URL params
		}

		// The gallery data IS in order, however
		for _, v := range info.GalleryData.Items {
			thumb, ok := mediaThumbs[v.MediaId]
			if ok {
				result.ThumbnailSources = append(result.ThumbnailSources, thumb)
			} else {
				return nil, errors.New("could not find media from gallery with id: " + v.MediaId)
			}
		}
	}

	return result, nil
}

func (rs *RedditScraper) ScrapeComment(postId string) (*ScrapeInfo, error) {
	src := &RedditCommentSource{
		ID: postId,
	}
	info, err := src.GetCommentInfo(rs)
	if err != nil {
		return nil, err
	}

	result := info.BasicScrapeInfo()
	result.SourceType = SourceRedditComment
	result.Title = "Comment by " + info.Author

	result.Meta = &RedditCommentMeta{
		RedditThingMeta: info.RedditThing.ToMeta(),
		IsSubmitter:     info.IsSubmitter,
	}

	return result, nil
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

func (rt *RedditThing) BasicScrapeInfo() *ScrapeInfo {
	item := &ScrapeInfo{
		CreditTitle: rt.Author,
		CreditURL:   getRedditAuthorURL(rt.Author),
		SourceKey:   rt.Id,
	}
	return item
}

func (rt *RedditThing) ToMeta() RedditThingMeta {
	return RedditThingMeta{
		Author:            rt.Author,
		Body:              rt.Body,
		Created:           rt.Created,
		Id:                rt.Id,
		Permalink:         rt.Permalink,
		Subreddit:         rt.Subreddit,
		SubredditPrefixed: rt.SubredditPrefixed,
	}
}

func getRedditAuthorURL(username string) string {
	return "https://www.reddit.com/u/" + username
}
