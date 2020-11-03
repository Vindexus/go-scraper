package vinscraper

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrNoConsumingScaper = errors.New("no scraper wanted to consume that url")
)

type SourceType string

type ScrapeReplacer struct {
	URLMatches  string // Will turn into regex. If the url matches, it will replace whatever matches FindMatches with ReplaceWith
	FindMatches string
	ReplaceWith string
}

type Scraping struct {
	Scrapers             []Scraper
	TitleReplacers       []ScrapeReplacer
	DescriptionReplacers []ScrapeReplacer
}

func NewScraping() *Scraping {
	return &Scraping{
		Scrapers: []Scraper{
			&RedditScraper{},
			&ScraperGeneric{},
		},
		TitleReplacers: []ScrapeReplacer{
		},
	}
}

type ScrapeInfo struct {
	CreditURL       string
	CreditTitle     string
	Description     string
	Meta interface{}
	SourceType      SourceType
	ThumbnailSource string
	Title           string
	URL string
	// TODO: Add logic to consolidate URL into a StandardizedURL so that youtu.be/123 and youtube.com/watch?v=123 and www.youtube.com/watch?v=123 all
	//  end up with the same StandardizedURL
}

type Scraper interface {
	// Returns true if this particular scraper wants to handle this url
	// This oftentimes looks at the domain name
	WantsURL(url string) bool
	Scrape(url string) (*ScrapeInfo, error)
}

func ScrapeReplace(link string, field *string, replaces []ScrapeReplacer) error {
	for _, v := range replaces {
		urlRE, err := regexp.Compile(link)
		if err != nil {
			return err
		}
		if urlRE.MatchString(link) {
			*field = strings.ReplaceAll(*field, v.FindMatches, v.ReplaceWith)
		}
	}
	return nil
}

func (s *Scraping) Scrape(link string) (*ScrapeInfo, error) {
	_, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	var item *ScrapeInfo
	for _, v := range s.Scrapers {
		if v.WantsURL(link) {
			item, err = v.Scrape(link)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if item == nil {
		return nil, ErrNoConsumingScaper
	}

	if err := ScrapeReplace(link, &item.Title, s.TitleReplacers); err != nil {
		return nil, err
	}

	if err := ScrapeReplace(link, &item.Description, s.DescriptionReplacers); err != nil {
		return nil, err
	}

	item.URL = link

	return item, nil
}

func IsImageLink(link string) bool {
	lower := strings.ToLower(link)
	parts := []string{".png", ".jpeg", ".jpg", ".gif"}
	for _, ext := range parts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}

	return false
}
