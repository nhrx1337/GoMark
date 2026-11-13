package blog

import (
	"context"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

func (b *Blog) Watch(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Could not start file watcher: %v", err)
		return
	}
	defer watcher.Close()

	if err := watcher.Add(b.postsDir); err != nil {
		log.Printf("Could not watch directory %q: %v", b.postsDir, err)
		return
	}

	log.Printf("Watching %s for changes", b.postsDir)

	var timer *time.Timer
	var timerC <-chan time.Time

	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			// Coalesce multiple file events into a single reload.
			if timer == nil {
				timer = time.NewTimer(150 * time.Millisecond)
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(150 * time.Millisecond)
			}
			timerC = timer.C

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)

		case <-timerC:
			timerC = nil

			posts, err := b.loadPosts()
			if err != nil {
				log.Printf("Could not reload posts: %v", err)
				continue
			}

			b.setPosts(posts)
			log.Printf("Posts reloaded: %d posts", len(posts))
		}
	}
}
