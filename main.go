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
}

type Authentication struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
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
				for _, variant := range variants {
					if variant.ContentType == "video/mp4" {
						download(strings.Split(variant.URL, "?")[0])
						fmt.Println(variant.URL)
					}
				}
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
