# GoMark

**GoMark** is a small self-hosted Markdown blog written in Go. Drop `.md` files
into `posts/` and they are rendered and served automatically. Edits are picked
up on the fly, no restart needed.

## Features

- Markdown posts with GitHub Flavored Markdown (tables, task lists and more).
- Optional front matter for title, tags, date, previous/next links and drafts.
- Full-text search with tag filtering that also works without JavaScript.
- RSS feed at `/feed.xml`.
- Automatic reload when posts are added, edited or removed.
- Previous/next navigation between posts.
- Copy-to-clipboard button on code blocks.
- Security headers and graceful shutdown out of the box.

## Installation

1. Clone the repository:

```bash
git clone https://github.com/nhrx1337/GoMark.git

cd GoMark
```

2. Download the dependencies:

```bash
go mod download
```

## Running

Start the server:

```bash
go run .
```

The blog is now available at http://localhost:8080.

Build a binary for production:

```bash
go build -o gomark .
./gomark
```

## Configuration

The server is configured with flags or environment variables:

| Flag      | Environment | Default | Description                        |
| --------- | ----------- | ------- | ---------------------------------- |
| `-addr`   | `BLOG_ADDR` | `:8080` | HTTP listen address                |
| `-posts`  | `BLOG_POSTS`| `posts` | Directory containing Markdown posts|

Example:

```bash
BLOG_ADDR=":3000" BLOG_POSTS="./content" ./gomark
```

## Writing posts

Create a Markdown file inside the `posts/` directory. The filename becomes the
URL slug, so `hello-world.md` is served at `/posts/hello-world`.

A post can start with optional front matter:

```markdown
---
title: Hello World
tags: go, web
date: 2026-01-31
previous: [[first-post]]
next: [[next-post]]
draft: false
---

Your content goes here.
```

Supported fields:

- `title` – post title (falls back to the first `# heading` or the filename).
- `tags` – comma separated tags.
- `date` – `YYYY-MM-DD`, `YYYY-MM-DD HH:MM` or RFC3339 (falls back to the file
  modification time).
- `previous` / `prev` – slug of the previous post, written as `[[slug]]`.
- `next` – slug of the next post, written as `[[slug]]`.
- `draft` – set to `true` to hide the post.

Everything after the closing `---` is rendered as Markdown.

## Routes

| Route            | Description                          |
| ---------------- | ------------------------------------ |
| `/`              | Post list with search and tags       |
| `/posts/{slug}`  | A single post                        |
| `/feed.xml`      | RSS 2.0 feed                         |
| `/healthz`       | Health check (`ok`)                  |

## Project Structure

```text
GoMark/
├── main.go                  # entry point, flags, server and shutdown
├── internal/
│   ├── blog/                # loading, rendering, handlers, feed, watcher
│   └── web/                 # logging, security headers, base URL helpers
├── templates/
│   └── page.html            # single embedded template
├── posts/                   # your Markdown posts
├── go.mod
└── go.sum
```
