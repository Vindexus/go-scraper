package vinscraper

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	SourceTwitterTweet = "twitter_tweet"
)

var (
	ErrTwitterNoConsumerKey    = errors.New("twitter ConsumerKey is blank")
	ErrTwitterNoConsumerSecret = errors.New("twitter ConsumerSecret is blank")
	ErrTwitterCantFindLinkId   = errors.New("could not find tweet id in link")
)

var tweetUrlRegexp = "twitter\\.com\\/.*\\/status(?:es)?\\/([^\\/\\?]+)"

type TwitterTweetMeta struct {
	AuthorAvatar     string
	AuthorName       string
	AuthorScreenName string
	Content          string
	LikesCount       int
	QuoteCount       int
	ReplyCount       int
	RetweetCount     int
}

type TwitterScraper struct {
	ConsumerKey    string
	ConsumerSecret string
}

func (ts *TwitterScraper) NewClient() (*twitter.Client, error) {
	if ts.ConsumerKey == "" {
		return nil, ErrTwitterNoConsumerKey
	}

	if ts.ConsumerSecret == "" {
		return nil, ErrTwitterNoConsumerSecret
	}

	// oauth2 configures a client that uses app credentials to keep a fresh token
	config := &clientcredentials.Config{
		ClientID:     ts.ConsumerKey,
		ClientSecret: ts.ConsumerSecret,
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth2.NoContext)

	// Twitter client
	client := twitter.NewClient(httpClient)
	return client, nil
}

func (ts *TwitterScraper) GetLinkTweetId(link string) (id int64, ok bool) {
	ok = false
	u, _ := url.Parse(link)
	host := strings.ToLower(u.Hostname())
	if !strings.HasSuffix(host, "twitter.com") {
		return
	}
	r, err := regexp.Compile(tweetUrlRegexp)
	if err != nil {
		return
	}
	result := r.FindStringSubmatch(link)

	if len(result) < 2 {
		return
	}

	id, err = strconv.ParseInt(result[1], 10, 64)
	if err != nil {
		return
	}

	ok = true
	return
}

func (ts *TwitterScraper) WantsURL(link string) bool {
	_, ok := ts.GetLinkTweetId(link)
	return ok
}

func (ts *TwitterScraper) Scrape(link string) (*ScrapeInfo, error) {
	client, err := ts.NewClient()
	if err != nil {
		return nil, err
	}

	id, ok := ts.GetLinkTweetId(link)
	if !ok {
		return nil, ErrTwitterCantFindLinkId
	}

	tweet, _, err := client.Statuses.Show(id, &twitter.StatusShowParams{
		TweetMode: "extended",
	})

	thumbnail := ""

	if len(tweet.Entities.Media) > 0 {
		for _, media := range tweet.Entities.Media {
			if media.Type == "photo" {
				thumbnail = media.MediaURLHttps
				break
			}
		}
	}

	avatar := tweet.User.ProfileImageURLHttps
	avatar = strings.Replace(avatar, "_normal.", ".", 1)

	info := &ScrapeInfo{
		CreditTitle:     tweet.User.ScreenName,
		CreditURL:       fmt.Sprintf("https://twitter.com/%s", tweet.User.ScreenName),
		SourceKey:       fmt.Sprintf("%d", id),
		SourceType:      SourceTwitterTweet,
		Title:           `Tweet by ` + tweet.User.ScreenName,
		ThumbnailSource: thumbnail,
		Meta: &TwitterTweetMeta{
			AuthorScreenName: tweet.User.ScreenName,
			AuthorName:       tweet.User.Name,
			AuthorAvatar:     avatar,
			Content:          tweet.FullText,
			LikesCount:       tweet.FavoriteCount,
			QuoteCount:       tweet.QuoteCount,
			RetweetCount:     tweet.RetweetCount,
			ReplyCount:       tweet.ReplyCount,
		},
	}

	return info, nil
}
