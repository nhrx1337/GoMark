package blog

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// Blog owns the loaded posts and their template.
type Blog struct {
	postsDir string
	template *template.Template
	markdown goldmark.Markdown
	mu       sync.RWMutex
	posts    []Post
}

func New(postsDir string, templateFS fs.FS) (*Blog, error) {
	tpl, err := template.ParseFS(templateFS, "templates/page.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	if err := os.MkdirAll(postsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create posts directory: %w", err)
	}

	b := &Blog{
		postsDir: postsDir,
		template: tpl,
		markdown: goldmark.New(goldmark.WithExtensions(extension.GFM)),
	}

	posts, err := b.loadPosts()
	if err != nil {
		return nil, err
	}
	b.posts = posts

	return b, nil
}

func (b *Blog) snapshot() []Post {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return append([]Post(nil), b.posts...)
}

func (b *Blog) setPosts(posts []Post) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.posts = posts
}
