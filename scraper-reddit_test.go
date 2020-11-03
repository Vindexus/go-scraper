package vinscraper

import (
	"github.com/Monstercat/golib/expectm"
	"os"
	"testing"
)

func getTestRedditScraper(t *testing.T) *RedditScraper {
	userAgent := os.Getenv("REDDIT_USER_AGENT")
	if userAgent == "" {
		t.Fatal("Set the REDDIT_USER_AGENT env variable to be able to run this test.")
	}
	return &RedditScraper{
		UserAgent: userAgent,
	}
}

func TestScrapeRedditWants(t *testing.T) {
	scraper := &RedditScraper{}

	// The generic scraper wants EVERY url, even broken ones
	tests := CreateWantTests(scraper, []string{
		"https://www.reddit.com/r/boardgames/comments/jn78c5/the_3_minute_board_games_top_100_games_2020/",
		"https://old.reddit.com/r/Warhammer40k/comments/jnaol2/my_halloween_costume_made_in_3_days_salamander/gb077ru?utm_source=share&utm_medium=web2x&context=3",
	}, []string{
		"https://blog.mywebsite.co.uk/posts/434324-how-do-i-shot-web",
		"https://google.com",
		"not a real url",
	}...)

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}

func TestScrapeRedditPost(t *testing.T) {
	scraper := getTestRedditScraper(t)
	tests := ApplyScraperTests(scraper, []*ScrapeTest{
		{
			URL: "https://www.reddit.com/r/boardgames/comments/jn78c5/the_3_minute_board_games_top_100_games_2020/",
			ExpectedM: &expectm.ExpectedM{
				"Title": "The 3 minute board games top 100 games (2020)",
				"CreditURL": "https://www.reddit.com/u/3minuteboardgames",
				"CreditTitle": "3minuteboardgames",
				"SourceType": "reddit_post",
				"Meta.subreddit_name_prefixed": "r/boardgames",
			},
		},
	})

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}
