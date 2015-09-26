package rss

import (
	"fmt"
	"net/http"
)

type Checker struct {
	feedURL    string
	seen       map[string]bool
	rememberer Rememberer
}

func NewChecker(feedURL string, rememberer Rememberer) *Checker {
	return &Checker{
		feedURL:    feedURL,
		seen:       make(map[string]bool),
		rememberer: rememberer,
	}
}

func (c *Checker) Check() ([]Item, error) {
	resp, err := http.Get(c.feedURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching feed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad response: %v", resp.StatusCode)
	}
	r, err := Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing feed: %v", err)
	}
	var ret []Item
	for _, item := range r.Channel.Items {
		if !c.rememberer.Ask(item.GUID) {
			ret = append(ret, item)
			c.rememberer.Tell(item.GUID)
		}
	}
	return ret, nil
}

type Rememberer interface {
	Ask(string) bool
	Tell(string)
}
