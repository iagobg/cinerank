package main

import (
	"log"
	"net/http"
	"cinerank/internal/ui"
)

func main() {
	mux := http.NewServeMux()

	// Home route
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := ui.HomePage().Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("render error: " + err.Error()))
			return
		}
	})

	// Demo HTMX endpoint
	mux.HandleFunc("/clicked", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := ui.ClickResult().Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("render error: " + err.Error()))
		}
	})

	// Static assets
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
