package main

import (
	"github.com/Monstercat/golib/expectm"
	"github.com/pkg/errors"
)

type ScrapeTest struct {
	ExpectedError error
	ExpectedM     *expectm.ExpectedM
	Scraper       Scraper
	URL           string
	Wants         bool
}

func (st *ScrapeTest) Run() error {
	if st.Scraper.WantsURL(st.URL) != st.Wants {
		if st.Wants {
			return errors.New("Test wanted scraper to want " + st.URL + " but it did NOT")
		}

		return errors.New("Test wanted scraper to NOT want " + st.URL + " but it DID")
	}

	res, err := st.Scraper.Scrape(st.URL)

	if err != st.ExpectedError {
		if st.ExpectedError == nil {
			return errors.Errorf("Expected no error but got '%s'", err)
		}
		return errors.Errorf("Expected error '%s' but bot '%s'", st.ExpectedError, err)
	}

	if st.ExpectedM != nil {
		if err := expectm.CheckJSON(res, st.ExpectedM); err != nil {
			return err
		}
	}

	return nil
}

func RunTests(sts []*ScrapeTest) error {
	for i, test := range sts {
		if err := test.Run(); err != nil {
			return errors.Errorf("[%d] %s", i, err)
		}
	}

	return nil
}

func CreateWantTests(scraper Scraper, urlsWanted []string, urlsNotWanted... string) []*ScrapeTest {
	numWanted := len(urlsWanted)
	tests := make([]*ScrapeTest, len(urlsNotWanted) + numWanted)
	for i, url := range urlsWanted {
		tests[i] = &ScrapeTest{
			Scraper: scraper,
			Wants:   true,
			URL:     url,
		}
	}

	for i, url := range urlsNotWanted {
		tests[i+numWanted] = &ScrapeTest{
			Scraper: scraper,
			Wants:   false,
			URL:     url,
		}
	}

	return tests
}

func ApplyScraperTests(scraper Scraper, tests []*ScrapeTest) []*ScrapeTest {
	for i, _ := range tests {
		tests[i].Scraper = scraper
		tests[i].Wants = true
	}
	return tests
}
