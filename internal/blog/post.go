package blog

import (
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Post is a single published entry.
type Post struct {
	Slug          string
	Title         string
	URL           string
	Date          time.Time
	Tags          []string
	HTML          template.HTML
	PreviousURL   string
	PreviousTitle string
	NextURL       string
	NextTitle     string
	Hidden        bool

	previousSlug string
	nextSlug     string
}

// Page is the data passed to the HTML template.
type Page struct {
	Posts    []Post
	Post     *Post
	AllTags  []string
	Query    string
	Tag      string
	Results  int
	NotFound bool
}

type frontMatter struct {
	Title        string
	Tags         []string
	Date         time.Time
	PreviousSlug string
	NextSlug     string
	Draft        bool
}

func (b *Blog) loadPosts() ([]Post, error) {
	entries, err := os.ReadDir(b.postsDir)
	if err != nil {
		return nil, fmt.Errorf("read posts directory: %w", err)
	}

	posts := make([]Post, 0, len(entries))

	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || !strings.EqualFold(ext, ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", entry.Name(), err)
		}

		data, err := os.ReadFile(filepath.Join(b.postsDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		slug := strings.TrimSuffix(entry.Name(), ext)
		if slug == "" {
			continue
		}

		meta, body := parsePostFile(data)
		if meta.Draft {
			continue
		}

		var rendered strings.Builder
		if err := b.markdown.Convert(body, &rendered); err != nil {
			return nil, fmt.Errorf("render %s: %w", entry.Name(), err)
		}

		date := meta.Date
		if date.IsZero() {
			date = info.ModTime()
		}

		title := meta.Title
		if title == "" {
			title = slug
		}

		posts = append(posts, Post{
			Slug:         slug,
			Title:        title,
			URL:          "/posts/" + url.PathEscape(slug),
			Date:         date,
			Tags:         meta.Tags,
			HTML:         template.HTML(rendered.String()),
			previousSlug: meta.PreviousSlug,
			nextSlug:     meta.NextSlug,
		})
	}

	sort.SliceStable(posts, func(i, j int) bool {
		if !posts[i].Date.Equal(posts[j].Date) {
			return posts[i].Date.After(posts[j].Date)
		}
		return strings.ToLower(posts[i].Title) < strings.ToLower(posts[j].Title)
	})

	bySlug := make(map[string]*Post, len(posts))
	for i := range posts {
		bySlug[posts[i].Slug] = &posts[i]
	}

	for i := range posts {
		if target, ok := bySlug[posts[i].previousSlug]; ok {
			posts[i].PreviousURL = target.URL
			posts[i].PreviousTitle = target.Title
		}
		if target, ok := bySlug[posts[i].nextSlug]; ok {
			posts[i].NextURL = target.URL
			posts[i].NextTitle = target.Title
		}
	}

	return posts, nil
}

func parsePostFile(data []byte) (frontMatter, []byte) {
	var meta frontMatter
	lines := strings.Split(string(data), "\n")
	body := data

	// Optional front matter:
	// ---
	// title: Post title
	// tags: go, web
	// date: 2025-01-31
	// previous: [[previous-post]]
	// next: [[next-post]]
	// draft: false
	// ---
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		end := -1

		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				end = i
				break
			}

			key, value, ok := strings.Cut(lines[i], ":")
			if !ok {
				continue
			}

			switch strings.ToLower(strings.TrimSpace(key)) {
			case "title":
				meta.Title = unquote(value)
			case "tags":
				meta.Tags = parseTags(value)
			case "date":
				meta.Date = parseDate(value)
			case "previous", "prev":
				meta.PreviousSlug = parseLinkValue(value)
			case "next":
				meta.NextSlug = parseLinkValue(value)
			case "draft":
				meta.Draft = strings.EqualFold(unquote(value), "true")
			}
		}

		if end >= 0 {
			body = []byte(strings.Join(lines[end+1:], "\n"))
		} else {
			// No closing delimiter: treat the whole file as content.
			meta = frontMatter{}
		}
	}

	if meta.Title == "" {
		for _, line := range strings.Split(string(body), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") {
				meta.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
				break
			}
		}
	}

	return meta, body
}

func parseTags(value string) []string {
	value = strings.Trim(strings.TrimSpace(value), "[]")
	var tags []string

	for _, tag := range strings.Split(value, ",") {
		tag = unquote(tag)
		if tag != "" && !hasTag(tags, tag) {
			tags = append(tags, tag)
		}
	}

	return tags
}

func parseDate(value string) time.Time {
	value = unquote(value)
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04", time.RFC3339} {
		if date, err := time.Parse(layout, value); err == nil {
			return date
		}
	}
	return time.Time{}
}

func parseLinkValue(value string) string {
	value = unquote(value)
	if strings.HasPrefix(value, "[[") && strings.HasSuffix(value, "]]") {
		return strings.TrimSpace(value[2 : len(value)-2])
	}
	return value
}

func unquote(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

func matchesQuery(post Post, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}

	if strings.Contains(strings.ToLower(post.Title), query) {
		return true
	}

	for _, tag := range post.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	return false
}

func hasTag(tags []string, target string) bool {
	for _, tag := range tags {
		if strings.EqualFold(tag, target) {
			return true
		}
	}
	return false
}

func collectTags(posts []Post) []string {
	seen := make(map[string]bool)
	var tags []string

	for _, post := range posts {
		for _, tag := range post.Tags {
			key := strings.ToLower(tag)
			if !seen[key] {
				seen[key] = true
				tags = append(tags, tag)
			}
		}
	}

	sort.Slice(tags, func(i, j int) bool {
		return strings.ToLower(tags[i]) < strings.ToLower(tags[j])
	})

	return tags
}
