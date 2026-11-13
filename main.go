package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"GoMark/internal/blog"
)

//go:embed templates/page.html
var templateFS embed.FS

func main() {
	addr := flag.String("addr", envOr("BLOG_ADDR", ":8080"), "HTTP listen address")
	postsDir := flag.String("posts", envOr("BLOG_POSTS", "posts"), "directory containing Markdown posts")
	flag.Parse()

	site, err := blog.New(*postsDir, templateFS)
	if err != nil {
		log.Fatalf("Could not start blog: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go site.Watch(ctx)

	server := &http.Server{
		Addr:              *addr,
		Handler:           site.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	log.Printf("GoMark running at http://%s", displayAddr(*addr))

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	case <-ctx.Done():
		log.Println("Shutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}
}

func displayAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "localhost" + addr
	}
	return addr
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
