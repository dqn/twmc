package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dqn/twmc/twmc"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:\n  twmc [options...] <consumer-key> <consumer-secret> <access-token> <access-token-secret> <search-word>\nOptions:")
		flag.PrintDefaults()
	}

	d := flag.String("d", "./", "output `directory`")
	w := flag.String("w", "", "comma separated `whitelist` for include Twitter clients")
	flag.Parse()

	if flag.NArg() != 5 {
		flag.Usage()
		os.Exit(1)
	}

	var wl []string
	if *w != "" {
		wl = strings.Split(*w, ",")
	}

	err := twmc.Collect(&twmc.TWMCConfig{
		Authentication: &twmc.Authentication{
			ConsumerKey:       flag.Arg(0),
			ConsumerSecret:    flag.Arg(1),
			AccessToken:       flag.Arg(2),
			AccessTokenSecret: flag.Arg(3),
		},
		Whitelist: wl,
		Outdir:    *d,
		StreamFilterParams: &twitter.StreamFilterParams{
			Track: strings.Split(flag.Arg(4), " "),
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}
