package main

import (
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

func main() {
	var config Config

	_, err := toml.DecodeFile("./config.toml", &config)
	if err != nil {
		log.Fatal(err)
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

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		isPermittedSource := false

		for _, source := range config.Source.Whitelist {
			isPermittedSource = isPermittedSource || strings.Contains(tweet.Source, source)
		}

		if !isPermittedSource {
			return
		}

		media := func() []twitter.MediaEntity {
			if tweet.ExtendedEntities != nil {
				return tweet.ExtendedEntities.Media
			}
			if tweet.Entities != nil {
				return tweet.Entities.Media
			}
			return nil
		}()

		for _, medium := range media {
			if variants := medium.VideoInfo.Variants; len(variants) > 0 {
				// 動画
				var (
					maxBitrate int
					url        string
				)

				for _, variant := range variants {
					// 最も Bitrate が高いエンティティを探す
					if variant.ContentType != "video/mp4" {
						continue
					}

					if variant.Bitrate < maxBitrate {
						continue
					}

					maxBitrate = variant.Bitrate
					url = variant.URL
				}

				download(strings.Split(url, "?")[0])
				fmt.Println(url)
			} else {
				// 画像
				download(medium.MediaURLHttps)
				fmt.Println(medium.MediaURLHttps)
			}
		}
	}

	stream, err := client.Streams.Filter(&config.StreamFilterParams)
	if err != nil {
		log.Fatal(err)
	}

	go demux.HandleChan(stream.Messages)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	stream.Stop()
}
