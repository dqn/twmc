# twmc

Collect images and videos from Twitter using Streaming API.

## Installation

```bash
$ go get github.com/dqn/twmc
```

## Usage

```sh
$ twmc [options...] <consumer-key> <consumer-secret> <access-token> <access-token-secret> <search-word>
```

### Options

- `-w`: comma separated whitelist for include Twitter clients.
- `-d`: output directory  (default `./`).

e.g.

```bash
$ twmc -w "Twitter for iPhone,Twitter for Android" -d "./media/" XXXXX XXXXX XXXXX XXXXX golang
```
