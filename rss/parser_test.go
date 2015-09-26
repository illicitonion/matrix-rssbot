package rss

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	xml := `<rss xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
	<channel>
		<description>A blog of stuff and things</description>
		<title>Quinoa and Beetroot</title>
		<generator>Tumblr (3.0; @quinoaandbeetroot)</generator>
		<item>
			<title>Something good</title>
      <description>
      Something really, really good.
      </description>
			<link>
			http://quinoaandbeetroot.tumblr.com/post/129857004035
			</link>
			<guid>
			http://quinoaandbeetroot.tumblr.com/post/129857004035
			</guid>
			<pubDate>Fri, 25 Sep 2015 13:43:44 -0400</pubDate>
		</item>
		<item>
			<title>
			Something bad
			</title>
			<link>
			http://quinoaandbeetroot.tumblr.com/post/129664837745
			</link>
			<guid>
			http://quinoaandbeetroot.tumblr.com/post/129664837745
			</guid>
			<pubDate>Tue, 22 Sep 2015 17:35:27 -0400</pubDate>
		</item>
	</channel>
</rss>`

	want := &RSS{
		Channel: Channel{
			Description: "A blog of stuff and things",
			Title:       "Quinoa and Beetroot",
			Items: []Item{
				{
					Title:       "Something good",
					Description: "Something really, really good.",
					Link:        "http://quinoaandbeetroot.tumblr.com/post/129857004035",
					GUID:        "http://quinoaandbeetroot.tumblr.com/post/129857004035",
					PubDate:     "Fri, 25 Sep 2015 13:43:44 -0400",
				},
				{
					Title:       "Something bad",
					Description: "",
					Link:        "http://quinoaandbeetroot.tumblr.com/post/129664837745",
					GUID:        "http://quinoaandbeetroot.tumblr.com/post/129664837745",
					PubDate:     "Tue, 22 Sep 2015 17:35:27 -0400",
				},
			},
		},
	}
	got, err := Parse(bytes.NewBufferString(xml))
	if err != nil {
		t.Fatalf("Got err: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Bad parse: want:\n%+v\ngot:\n%+v\n", want, got)
	}
}
