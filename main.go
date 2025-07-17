package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var appVersion = strconv.FormatInt(time.Now().Unix(), 10)

func main() {
	store := NewTodoStore()
	handlers := NewHandlers(store, appVersion)

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("POST /todos", handlers.AddTodoHandler)
	mux.HandleFunc("PUT /todos/{id}", handlers.UpdateTodoHandler)
	mux.HandleFunc("POST /todos/reorder", handlers.ReorderTodosHandler)
	mux.HandleFunc("PATCH /todos/{id}/toggle", handlers.ToggleTodoHandler)
	mux.HandleFunc("DELETE /todos/{id}", handlers.DeleteTodoHandler)
	mux.HandleFunc("POST /save", handlers.SaveHandler)
	mux.HandleFunc("GET /export", handlers.ExportHandler)
	
	fs := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", addCacheHeaders(fs)))

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func addCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if version query parameter is present
		if r.URL.Query().Get("v") != "" {
			// If versioned, cache for a long time
			w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
		} else {
			// If not versioned, don't cache
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		next.ServeHTTP(w, r)
	})
}