package blog

import (
	"net/http"
	"strings"

	"GoMark/internal/web"
)

func (b *Blog) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", b.handleIndex)
	mux.HandleFunc("GET /posts/{slug}", b.handlePost)
	mux.HandleFunc("GET /feed.xml", b.handleFeed)
	mux.HandleFunc("GET /healthz", handleHealth)

	return web.LogRequests(web.CommonHeaders(mux))
}

func (b *Blog) handleIndex(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))

	posts := b.snapshot()
	filtered := make([]Post, 0, len(posts))
	results := 0

	for _, post := range posts {
		if tag != "" && !hasTag(post.Tags, tag) {
			continue
		}
		// Keep every post in the page so client-side search still works,
		// but hide non-matches so search also works without JavaScript.
		post.Hidden = !matchesQuery(post, query)
		if !post.Hidden {
			results++
		}
		filtered = append(filtered, post)
	}

	b.render(w, r, http.StatusOK, Page{
		Posts:   filtered,
		AllTags: collectTags(posts),
		Query:   query,
		Tag:     tag,
		Results: results,
	})
}

func (b *Blog) handlePost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	posts := b.snapshot()

	var selected *Post
	for i := range posts {
		if posts[i].Slug == slug {
			selected = &posts[i]
			break
		}
	}

	if selected == nil {
		b.notFound(w, r)
		return
	}

	b.render(w, r, http.StatusOK, Page{Post: selected})
}

func (b *Blog) notFound(w http.ResponseWriter, r *http.Request) {
	b.render(w, r, http.StatusNotFound, Page{NotFound: true})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte("ok\n"))
}
