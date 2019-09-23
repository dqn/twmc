package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Config struct {
	Authentication Authentication
	StreamFilterParams twitter.StreamFilterParams
}

type Authentication struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
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
		fmt.Println(tweet.Text)
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
