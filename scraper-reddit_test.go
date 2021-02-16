package vinscraper

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/monstercat/golib/expectm"
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

	tests := CreateWantTests(scraper, []string{
		"https://www.reddit.com/r/boardgames/comments/jn78c5/the_3_minute_board_games_top_100_games_2020/",
		"https://old.reddit.com/r/Warhammer40k/comments/jnaol2/my_halloween_costume_made_in_3_days_salamander/gb077ru?utm_source=share&utm_medium=web2x&context=3",
	}, []string{
		"https://wordpress.org/showcase/ladybird-education/",
		"https://google.com",
		"not a real url",
	}...)

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}

func ExpectContains(finds ...string) func(val interface{}) error {
	return func(val interface{}) error {
		for _, find := range finds {
			s := fmt.Sprintf("%s", val)
			if !strings.Contains(s, find) {
				return errors.New(fmt.Sprintf(`string does not contain "%s" in '%s'`, find, s))

			}
		}
		return nil
	}
}

func TestScrapeRedditPost(t *testing.T) {
	scraper := getTestRedditScraper(t)
	tests := ApplyScraperTests(scraper, []*ScrapeTest{
		{
			URL: "https://www.reddit.com/r/boardgames/comments/jn78c5/the_3_minute_board_games_top_100_games_2020/",
			ExpectedM: &expectm.ExpectedM{
				"Title":                  "The 3 minute board games top 100 games (2020)",
				"CreditURL":              "https://www.reddit.com/u/3minuteboardgames",
				"CreditTitle":            "3minuteboardgames",
				"SourceKey":              "jn78c5",
				"SourceType":             "reddit_post",
				"Meta.SubredditPrefixed": "r/boardgames",
			},
		},
		{
			URL: "https://www.reddit.com/r/wholesomememes/comments/jo26nf/he_lives_in_my_house/",
			ExpectedM: &expectm.ExpectedM{
				"SourceKey":              "jo26nf",
				"SourceType":             "reddit_post",
				"ThumbnailSources.0":     "https://i.redd.it/tjvedcdip9x51.jpg",
				"Meta.SubredditPrefixed": "r/wholesomememes",
			},
		},
		{
			// Album with 7 images
			URL: "https://www.reddit.com/r/Warhammer40k/comments/lki67h/some_tau_this_time_ghostkeel_leaving_its_stealth/",
			ExpectedM: &expectm.ExpectedM{
				"SourceType":             "reddit_post",
				"ThumbnailSources.#":     8,
				"Meta.SubredditPrefixed": "r/Warhammer40k",
				"ThumbnailSources.0":     ExpectContains("4dvn70o8doh61.jpg"),
				"ThumbnailSources.7":     ExpectContains("j61kr1n9doh61.jpg"),
			},
		},
	})

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}

func TestScrapeRedditComment(t *testing.T) {
	scraper := getTestRedditScraper(t)
	tests := ApplyScraperTests(scraper, []*ScrapeTest{
		{
			URL: "https://www.reddit.com/r/wholesomememes/comments/jni5k5/i_love_my_dad/gb1mxkl?utm_source=share&utm_medium=web2x&context=3",
			ExpectedM: &expectm.ExpectedM{
				"Title":                  "Comment by nodgers132",
				"CreditURL":              "https://www.reddit.com/u/nodgers132",
				"CreditTitle":            "nodgers132",
				"SourceKey":              "gb1mxkl",
				"SourceType":             "reddit_comment",
				"Meta.SubredditPrefixed": "r/wholesomememes",
			},
		},
	})

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}
