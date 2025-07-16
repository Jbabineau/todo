package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	store := NewTodoStore()
	handlers := NewHandlers(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("POST /todos", handlers.AddTodoHandler)
	mux.HandleFunc("POST /todos/reorder", handlers.ReorderTodosHandler)
	mux.HandleFunc("PATCH /todos/{id}/toggle", handlers.ToggleTodoHandler)
	mux.HandleFunc("DELETE /todos/{id}", handlers.DeleteTodoHandler)
	mux.HandleFunc("POST /save", handlers.SaveHandler)
	mux.HandleFunc("GET /export", handlers.ExportHandler)
	
	fs := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}