package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Config struct {
	Authentication     Authentication
	StreamFilterParams twitter.StreamFilterParams
	Source             Source
}

type Authentication struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

type Source struct {
	Whitelist []string
}

func download(url string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	_, filename := path.Split(url)
	out, err := os.Create(fmt.Sprintf("./media/%v", filename))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func getVideoUrl(variants []twitter.VideoVariant) string {
	var (
		maxBitrate int
		url        string
	)

	for _, variant := range variants {
		// 最も Bitrate が高いエンティティを探す
		if (variant.ContentType == "video/mp4") && (variant.Bitrate > maxBitrate) {
			maxBitrate = variant.Bitrate
			url = variant.URL
		}
	}

	return strings.Split(url, "?")[0]
}

func getMedia(tweet *twitter.Tweet) []twitter.MediaEntity {
	if tweet.ExtendedEntities != nil {
		return tweet.ExtendedEntities.Media
	}
	if tweet.Entities != nil {
		return tweet.Entities.Media
	}

	return nil
}

func run() error {
	var config Config

	_, err := toml.DecodeFile("./config.toml", &config)
	if err != nil {
		log.Fatal(err)
	}

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		isSourceAvailable := false
		for _, availableSource := range config.Source.Whitelist {
			isSourceAvailable = isSourceAvailable || strings.Contains(tweet.Source, availableSource)
		}
		if !isSourceAvailable {
			return
		}

		for _, medium := range getMedia(tweet) {
			var url string
			if variants := medium.VideoInfo.Variants; len(variants) > 0 {
				// 動画
				url = getVideoUrl(variants)
			} else {
				// 画像
				url = medium.MediaURLHttps
			}

			download(url)
		}
	}

	client := twitter.NewClient(
		oauth1.NewConfig(
			config.Authentication.ConsumerKey,
			config.Authentication.ConsumerSecret,
		).Client(
			oauth1.NoContext,
			oauth1.NewToken(
				config.Authentication.AccessToken,
				config.Authentication.AccessSecret,
			),
		),
	)

	stream, err := client.Streams.Filter(&config.StreamFilterParams)
	if err != nil {
		log.Fatal(err)
	}

	go demux.HandleChan(stream.Messages)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	stream.Stop()

	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:\n  twmc [options...] <search-word> <consumer-key> <consumer-secret> <access-token> <access-token-secret>\nOptions:")
		flag.PrintDefaults()
	}

	d := flag.String("d", "./", "output `directory`")
	w := flag.String("w", "", "comma separated `whitelist` for include Twitter clients")
	flag.Parse()

	if flag.NArg() != 5 {
		flag.Usage()
		os.Exit(1)
	}

	println(d, w)
}
