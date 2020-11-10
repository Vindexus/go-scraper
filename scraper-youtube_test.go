package vinscraper

import (
	"github.com/monstercat/golib/expectm"
	"os"
	"testing"
)

func getTestYouTubeScraper(t *testing.T) *YouTubeScraper {
	configFile := os.Getenv("YOUTUBE_CONFIG_FILE")
	if configFile == "" {
		t.Fatal("Set the YOUTUBE_CONFIG_FILE env variable to be able to run this test.")
	}
	tokenFile := os.Getenv("YOUTUBE_TOKEN_FILE")
	if tokenFile == "" {
		t.Fatal("Set the YOUTUBE_TOKEN_FILE env variable to be able to run this test.")
	}
	scraper, err := LoadYouTubeScraper(configFile, tokenFile)
	if err != nil {
		t.Fatal(err)
	}
	return scraper
}

func TestScrapeYouTubeWants(t *testing.T) {
	scraper := &YouTubeScraper{}

	// The generic scraper wants EVERY url, even broken ones
	tests := CreateWantTests(scraper, []string{
		"https://www.youtube.com/watch?v=DP0t2MmOMEA",
		"https://youtu.be/DP0t2MmOMEA",
	}, []string{
		"https://www.reddit.com/r/boardgames/comments/jn78c5/the_3_minute_board_games_top_100_games_2020/",
		"https://wordpress.org/showcase/ladybird-education/",
		"https://google.com",
		"not a real url",
	}...)

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}

func TestScrapeYouTubeVideos(t *testing.T) {
	scraper := getTestYouTubeScraper(t)
	tests := ApplyScraperTests(scraper, []*ScrapeTest{
		{
			URL: "https://www.youtube.com/watch?v=DP0t2MmOMEA&t=432s",
			ExpectedM: &expectm.ExpectedM{
				"Title": "Primitive Technology: Wood Ash Cement",
				"CreditURL": "https://www.youtube.com/channel/UCAL3JXZSzSm8AlZyD3nQdBA",
				"CreditTitle": "Primitive Technology",
				"SourceKey": "DP0t2MmOMEA",
				"SourceType": "youtube_video",
				"Meta.Duration": "3m54s",
			},
		},
		{
			URL: "https://youtu.be/DP0t2MmOMEA",
			ExpectedM: &expectm.ExpectedM{
				"Title": "Primitive Technology: Wood Ash Cement",
				"CreditURL": "https://www.youtube.com/channel/UCAL3JXZSzSm8AlZyD3nQdBA",
				"CreditTitle": "Primitive Technology",
				"SourceKey": "DP0t2MmOMEA",
				"SourceType": "youtube_video",
				"Meta.Duration": "3m54s",
				"ThumbnailSource": "https://i.ytimg.com/vi/DP0t2MmOMEA/default.jpg",
			},
		},
	})

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}
