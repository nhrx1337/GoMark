package blog

import (
	"bytes"
	"encoding/xml"
	"log"
	"net/http"
	"time"

	"GoMark/internal/web"
)

type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language,omitempty"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func (b *Blog) handleFeed(w http.ResponseWriter, r *http.Request) {
	base := web.BaseURL(r)
	posts := b.snapshot()

	feed := rss{
		Version: "2.0",
		Channel: rssChannel{
			Title:       "GoMark",
			Link:        base + "/",
			Description: "Recent posts",
			Language:    "en",
			Items:       make([]rssItem, 0, len(posts)),
		},
	}

	for _, post := range posts {
		link := base + post.URL
		feed.Channel.Items = append(feed.Channel.Items, rssItem{
			Title:       post.Title,
			Link:        link,
			GUID:        link,
			PubDate:     post.Date.Format(time.RFC1123Z),
			Description: string(post.HTML),
		})
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(feed); err != nil {
		log.Printf("Could not build feed: %v", err)
		http.Error(w, "Could not build feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("Could not write feed: %v", err)
	}
}
