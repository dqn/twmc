package twmc

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type TWMC struct {
}

type TWMCConfig struct {
	Authentication     *Authentication
	Whitelist          []string
	Outdir             string
	StreamFilterParams *twitter.StreamFilterParams
}

type Authentication struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func containsString(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

func chooseHighestQualityVideoURL(variants []twitter.VideoVariant) string {
	var (
		maxBitrate int
		url        string
	)

	for _, v := range variants {
		println(v.ContentType)
		if (v.ContentType == "video/mp4") && (v.Bitrate > maxBitrate) {
			maxBitrate = v.Bitrate
			url = v.URL
		}
	}

	return strings.Split(url, "?")[0]
}

func getMediaEntity(t *twitter.Tweet) []twitter.MediaEntity {
	if t.ExtendedEntities != nil {
		return t.ExtendedEntities.Media
	}
	if t.Entities != nil {
		return t.Entities.Media
	}

	return nil
}

func chooseMediaURL(m *twitter.MediaEntity) string {
	if variants := m.VideoInfo.Variants; len(variants) > 0 {
		return chooseHighestQualityVideoURL(variants)
	}

	return m.MediaURLHttps
}

func makeTwitterClient(auth *Authentication) *twitter.Client {
	config := oauth1.NewConfig(auth.ConsumerKey, auth.ConsumerSecret)
	token := oauth1.NewToken(auth.AccessToken, auth.AccessTokenSecret)
	client := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(client)
}

func Collect(config *TWMCConfig) error {
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(t *twitter.Tweet) {
		if config.Whitelist != nil && !containsString(config.Whitelist, t.Source) {
			return
		}

		for _, v := range getMediaEntity(t) {
			url := chooseMediaURL(&v)

			_, file := path.Split(url)
			dest := path.Join(config.Outdir, file)
			download(url, dest)
		}
	}

	client := makeTwitterClient(config.Authentication)
	stream, err := client.Streams.Filter(config.StreamFilterParams)
	if err != nil {
		return err
	}

	go demux.HandleChan(stream.Messages)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	stream.Stop()

	return nil
}
