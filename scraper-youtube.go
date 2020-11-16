package vinscraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

var (
	ErrNoYouTubeId   = errors.New("could not find video ID in link")
	ErrVideoNotFound = errors.New("video not found through API")
)

const (
	SourceYouTubeVideo SourceType = "youtube_video"
)

type YouTubeVideoMeta struct {
	Duration string
	Tags     []string
}

type YouTubeScraper struct {
	OAuthConfig *oauth2.Config
	OAuthToken  *oauth2.Token
}

func LoadYouTubeScraper(configFile string, tokenFile string) (*YouTubeScraper, error) {
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	return &YouTubeScraper{
		OAuthConfig: config,
		OAuthToken:  token,
	}, nil
}

func (yt *YouTubeScraper) GetService() (*youtube.Service, error) {
	ctx := context.Background()

	client := yt.OAuthConfig.Client(ctx, yt.OAuthToken)
	service, err := youtube.New(client)
	return service, err
}

func (yt *YouTubeScraper) WantsURL(link string) bool {
	u, _ := url.Parse(link)
	host := strings.ToLower(u.Hostname())
	if !strings.HasSuffix(host, "youtube.com") && !strings.HasSuffix(host, "youtu.be") {
		return false
	}
	videoId := GetLinkYouTubeVideoId(link)
	if videoId != "" {
		return true
	}
	return false
}

func (yt *YouTubeScraper) Scrape(link string) (*ScrapeInfo, error) {
	id := GetLinkYouTubeVideoId(link)
	if id == "" {
		return nil, ErrNoYouTubeId
	}

	service, err := yt.GetService()
	if err != nil {
		return nil, err
	}

	call := service.Videos.List([]string{"contentDetails", "snippet"})
	call = call.Id(id)
	list, err := call.Do()
	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, errors.Wrap(ErrVideoNotFound, fmt.Sprintf("ID: '%s'", id))
	}

	snip := list.Items[0].Snippet
	durationS := list.Items[0].ContentDetails.Duration
	dur := parseDuration(durationS)
	since := time.Since(time.Now().Add(dur * -1))

	info := &ScrapeInfo{
		CreditTitle: snip.ChannelTitle,
		CreditURL:   fmt.Sprintf("https://www.youtube.com/channel/%s", snip.ChannelId),
		Description: snip.Description,
		Meta: &YouTubeVideoMeta{
			Duration: formatDuration(since),
			Tags:     snip.Tags,
		},
		SourceType: SourceYouTubeVideo,
		SourceKey:  id,
		Title:      snip.Title,
	}

	if snip.Thumbnails.Default != nil {
		info.ThumbnailSource = snip.Thumbnails.High.Url
	}

	return info, nil
}

// This function assumes that the link is a YouTube video link
// It does not do any validation of that assumption
func GetLinkYouTubeVideoId(link string) string {
	u, _ := url.Parse(link)
	query := u.Query()
	v := query.Get("v")
	if v != "" {
		return v
	}

	// Very presumptious that the last part is the ID
	path := u.Path
	pieces := strings.Split(path, "/")
	return pieces[len(pieces)-1]
}

func formatDuration(dur time.Duration) string {
	str := dur.String()
	str = strings.Replace(str, "h0m", "h00m", 1)
	str = strings.Replace(str, "m0s", "m00s", 1)
	reg, _ := regexp.Compile("\\.[0-9]+s")
	if reg.MatchString(str) {
		str = reg.ReplaceAllString(str, "s")
	}
	return str
}

func parseDuration(str string) time.Duration {
	durationRegex := regexp.MustCompile(`P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`)
	matches := durationRegex.FindStringSubmatch(str)

	years := parseInt64(matches[1])
	months := parseInt64(matches[2])
	days := parseInt64(matches[3])
	hours := parseInt64(matches[4])
	minutes := parseInt64(matches[5])
	seconds := parseInt64(matches[6])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)
	return time.Duration(years*24*365*hour + months*30*24*hour + days*24*hour + hours*hour + minutes*minute + seconds*second)
}

func parseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}
	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}
	return int64(parsed)
}
