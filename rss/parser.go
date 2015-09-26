package rss

import (
	"encoding/xml"
	"io"
	"strings"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Description string `xml:"description"`
	Title       string `xml:"title"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

func Parse(r io.Reader) (*RSS, error) {
	var feed RSS
	err := xml.NewDecoder(r).Decode(&feed)
	if err != nil {
		return nil, err
	}
	feed.Channel.Description = strings.TrimSpace(feed.Channel.Description)
	feed.Channel.Title = strings.TrimSpace(feed.Channel.Title)
	for i, it := range feed.Channel.Items {
		feed.Channel.Items[i] = Item{
			Title:       strings.TrimSpace(it.Title),
			Description: strings.TrimSpace(it.Description),
			Link:        strings.TrimSpace(it.Link),
			GUID:        strings.TrimSpace(it.GUID),
			PubDate:     strings.TrimSpace(it.PubDate),
		}
	}
	return &feed, nil
}
