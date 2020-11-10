package vinscraper

import "testing"

func TestScrapeGeneric(t *testing.T) {
	scraper := &ScraperGeneric{}

	// The generic scraper wants EVERY url, even broken ones
	tests := CreateWantTests(scraper, []string{
		"https://www.reddit.com/r/boardgames/comments/jn78c5/the_3_minute_board_games_top_100_games_2020/",
		"https://google.com",
		"https://wordpress.org/showcase/ladybird-education/",
		"this isn't a url",
	})

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}
