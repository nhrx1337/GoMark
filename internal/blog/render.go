package blog

import (
	"bytes"
	"log"
	"net/http"
)

func (b *Blog) render(w http.ResponseWriter, r *http.Request, status int, page Page) {
	var buf bytes.Buffer
	if err := b.template.ExecuteTemplate(&buf, "page.html", page); err != nil {
		log.Printf("Could not render %s: %v", r.URL.Path, err)
		http.Error(w, "Could not render page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if r.Method != http.MethodHead {
		if _, err := buf.WriteTo(w); err != nil {
			log.Printf("Could not write response for %s: %v", r.URL.Path, err)
		}
	}
}
