package vinscraper

import (
	"net/http"

	"github.com/dyatlov/go-htmlinfo/htmlinfo"
)

type ScraperGeneric struct {
}

const (
	SourceURL = "url"
)

// The generic one wants them all
func (s *ScraperGeneric) WantsURL(url string) bool {
	return true
}

func (s *ScraperGeneric) Scrape(link string) (*ScrapeInfo, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	info := htmlinfo.NewHTMLInfo()

	// if url can be nil too, just then we won't be able to fetch (and generate) oembed information
	err = info.Parse(resp.Body, &link, nil)

	if err != nil {
		panic(err)
	}

	item := &ScrapeInfo{
		SourceType: SourceURL,
		SourceKey: link,
	}

	if info.OGInfo != nil {
		item.Title = info.OGInfo.Title
		item.Description = info.OGInfo.Description
		if len(info.OGInfo.Images) > 0 {
			item.ThumbnailSource = info.OGInfo.Images[0].URL
		}
	}

	if item.Title == "" {
		item.Title = info.Title
	}

	if item.Description == "" {
		item.Description = info.Description
	}

	if info.AuthorName != "" {
		item.CreditTitle = info.AuthorName
	}

	if info.ImageSrcURL != "" {
		item.ThumbnailSource = info.ImageSrcURL
	}

	return item, nil
}
