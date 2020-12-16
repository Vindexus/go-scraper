package vinscraper

import (
	"os"
	"testing"

	"github.com/monstercat/golib/expectm"
)

func getTestTwitterScraper(t *testing.T) *TwitterScraper {
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	if consumerKey == "" {
		t.Fatal("Set the TWITTER_CONSUMER_KEY env variable to be able to run this test.")
	}
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	if consumerKey == "" {
		t.Fatal("Set the TWITTER_CONSUMER_SECRET env variable to be able to run this test.")
	}
	return &TwitterScraper{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
	}
}

func TestScrapeTwitterWants(t *testing.T) {
	scraper := &TwitterScraper{}

	tests := CreateWantTests(scraper, []string{
		"https://twitter.com/Cephalofair/status/1328452020060254210",
		"https://twitter.com/Cephalofair/status/1328452020060254210?stuff=tre#whatever",
	}, []string{
		"https://wordpress.org/showcase/ladybird-education/",
		"https://google.com",
		"not a real url",
	}...)

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}

func TestScrapeTwitterTweet(t *testing.T) {
	scraper := getTestTwitterScraper(t)
	tests := ApplyScraperTests(scraper, []*ScrapeTest{
		{
			URL: "https://twitter.com/GauntletRPG/status/1329107098106466310",
			ExpectedM: &expectm.ExpectedM{
				"Meta.Content": `Codex Flashback: Iron (April '17)
Includes: The Gates of Cold Iron Pass, an OSR adventure; Wind on the Path, a game of samurai duels; Four Dwarven Shrines, a collection of elements for Dungeon World, and more.
Find back issues of Codex on DTRPG:
https://t.co/wszCC71oCU https://t.co/LVfVx8bzOm`,
			},
		},
		{
			// This has an image embed
			URL: "https://twitter.com/Cephalofair/status/1321504060680343552",
			ExpectedM: &expectm.ExpectedM{
				"Title":           "Tweet by Cephalofair",
				"CreditURL":       "https://twitter.com/Cephalofair",
				"CreditTitle":     "Cephalofair",
				"SourceKey":       "1321504060680343552",
				"SourceType":      "twitter_tweet",
				"ThumbnailSource": "https://pbs.twimg.com/media/ElbspJoXUAE4yzs.jpg",
				"Meta.LikesCount": 88, // This is probably not a good test since likes can change
			},
		},
		{
			// This has a video embed
			URL: "https://twitter.com/Cephalofair/status/1317517262568497152",
			ExpectedM: &expectm.ExpectedM{
				"Title":             "Tweet by Cephalofair",
				"ThumbnailSource":   "https://pbs.twimg.com/tweet_video_thumb/EkjCunXXIAYcRz3.jpg",
				"Meta.AuthorAvatar": "https://pbs.twimg.com/profile_images/1256353589552955398/Azn12qgL.jpg",
			},
		},
		{
			// No media embed
			URL: "https://twitter.com/Cephalofair/status/1306671389802364930",
			ExpectedM: &expectm.ExpectedM{
				"Meta.Content": "Online co-op is now live for Gloomhaven Digital! Congrats to @AsmodeeDigital and @FlamingFowl for making this happen! Check out all the new features in the trailer (SPOILERS for the Music Note: https://t.co/eUJ5SNuE5a), and grab it while it's on sale: https://t.co/lJ0BemUXNM!",
			},
		},
	})

	if err := RunTests(tests); err != nil {
		t.Error(err)
	}
}
