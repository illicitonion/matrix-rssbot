package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.org/illicitonion/matrix-rssbot/bridge"
	"github.org/illicitonion/matrix-rssbot/event"
	"github.org/illicitonion/matrix-rssbot/rss"

	yaml "gopkg.in/yaml.v2"
)

var (
	configFile = flag.String("config_file", "config.yaml", "Path to config file")
	port       = flag.Int("port", 9000, "Port on which to listen")
)

func handle(w http.ResponseWriter, req *http.Request) {
	b, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(b))
	w.Write([]byte("{}"))
}

type config struct {
	HomeserverURL  string        `yaml:"homeserver_url"`
	ASToken        string        `yaml:"as_token"`
	BotUsername    string        `yaml:"bot_username"`
	RoomMappings   []roomMapping `yaml:"room_mapping"`
	EntryCachePath string        `yaml:"entry_cache_path"`
}

type roomMapping struct {
	RoomID  string `yaml:"room_id"`
	FeedURL string `yaml:"feed_url"`
}

func main() {
	c := readConfig()
	b := bridge.New(c.HomeserverURL, c.ASToken, c.BotUsername)

	rememberer, err := NewfileRememberer(c.EntryCachePath)
	if err != nil {
		log.Fatal(err)
	}

	for _, rm := range c.RoomMappings {
		watcher := rss.NewChecker(rm.FeedURL, rememberer)
		go func(rm roomMapping, w *rss.Checker) {
			t := time.Tick(1 * time.Minute)
			for _ = range t {
				log.Printf("Checking for new posts...")
				entries, err := watcher.Check()
				if err != nil {
					log.Printf("Error checking: %v", err)
					continue
				}
				for i := len(entries) - 1; i >= 0; i-- {
					entry := entries[i]
					messages, err := formatEntry(entry, rm)
					if err != nil {
						log.Printf("Error formatting entry: %v", err)
						continue
					}
					for _, m := range messages {
						b.SendMessage(rm.RoomID, &m)
					}
				}
			}
		}(rm, watcher)
	}

	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func formatEntry(item rss.Item, rm roomMapping) ([]event.Message, error) {
	if rm.FeedURL == "https://xkcd.com/rss.xml" {
		return parseXKCD(item)
	}

	return []event.Message{{
		Body: fmt.Sprintf("%s - %s", item.Title, item.Link),
		//Format:        "org.matrix.custom.html",
		//FormattedBody: fmt.Sprintf(`<a href="%s">%s</a>`, entry.Link, entry.Title),
		Msgtype: "m.text",
	}}, nil
}

func readConfig() config {
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	var c config
	if err := yaml.Unmarshal(b, &c); err != nil {
		log.Fatal(err)
	}
	return c
}

type fileRememberer struct {
	cache map[string]bool
	path  string
	mu    sync.Mutex
}

func NewfileRememberer(path string) (*fileRememberer, error) {
	r := &fileRememberer{
		cache: make(map[string]bool),
		path:  path,
	}
	f, err := os.Open(path)
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok && pathErr.Err == syscall.ENOENT {
			return r, nil
		}
		return r, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		r.cache[scanner.Text()] = true
	}
	return r, scanner.Err()
}

func (r *fileRememberer) Ask(s string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.cache[stripNewlines(s)]
}

func (r *fileRememberer) Tell(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[stripNewlines(s)] = true

	f, err := os.OpenFile(r.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Error opening rememberer file: %v", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(stripNewlines(s) + "\n"); err != nil {
		log.Printf("Error writing to rememberer file: %v", err)
	}
}

func stripNewlines(s string) string {
	return strings.Replace(s, "\n", "", -1)
}

type ImgTag struct {
	Src   string `xml:"src,attr"`
	Title string `xml:"title,attr"`
}

func parseXKCD(i rss.Item) ([]event.Message, error) {
	var img ImgTag
	if err := xml.Unmarshal([]byte(i.Description), &img); err != nil {
		return nil, err
	}

	return []event.Message{
		{
			Body:    i.Title,
			Msgtype: "m.text",
		},
		{
			Body:    i.Link,
			URL:     strings.Replace(img.Src, "http://", "https://", 1),
			Msgtype: "m.image",
		},
		{
			Body:    img.Title,
			Msgtype: "m.text",
		},
	}, nil
}
