package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo-app/templates"

	"github.com/a-h/templ"
)

type Handlers struct {
	store   *TodoStore
	version string
}

func NewHandlers(store *TodoStore, version string) *Handlers {
	return &Handlers{store: store, version: version}
}

func (h *Handlers) HomeHandler(w http.ResponseWriter, r *http.Request) {
	todos := h.store.GetTodos()
	templateTodos := make([]templates.Todo, len(todos))
	for i, todo := range todos {
		templateTodos[i] = templates.Todo{
			ID:        todo.ID,
			Text:      todo.Text,
			Priority:  todo.Priority,
			Category:  todo.Category,
			DueDate:   todo.DueDate,
			SortOrder: todo.SortOrder,
			Completed: todo.Completed,
		}
	}
	component := templates.TodoApp(templateTodos, h.version)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *Handlers) AddTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "Todo text is required", http.StatusBadRequest)
		return
	}

	priority := r.FormValue("priority")
	if priority == "" {
		priority = "medium"
	}

	category := r.FormValue("category")

	var dueDate *time.Time
	dueDateStr := r.FormValue("due_date")
	if dueDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dueDateStr); err == nil {
			dueDate = &parsed
		}
	}

	todo := h.store.AddTodo(text, priority, category, dueDate)
	templateTodo := templates.Todo{
		ID:        todo.ID,
		Text:      todo.Text,
		Priority:  todo.Priority,
		Category:  todo.Category,
		DueDate:   todo.DueDate,
		SortOrder: todo.SortOrder,
		Completed: todo.Completed,
	}
	component := templates.TodoItem(templateTodo)
	templ.Handler(component).ServeHTTP(w, r)
}

func (h *Handlers) ToggleTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if !h.store.ToggleTodo(id) {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	todos := h.store.GetTodos()
	for _, todo := range todos {
		if todo.ID == id {
			templateTodo := templates.Todo{
				ID:        todo.ID,
				Text:      todo.Text,
				Priority:  todo.Priority,
				Category:  todo.Category,
				DueDate:   todo.DueDate,
				SortOrder: todo.SortOrder,
				Completed: todo.Completed,
			}
			component := templates.TodoItem(templateTodo)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}
	}
}

func (h *Handlers) DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if !h.store.DeleteTodo(id) {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) UpdateTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "Todo text is required", http.StatusBadRequest)
		return
	}

	priority := r.FormValue("priority")
	if priority == "" {
		priority = "medium"
	}

	category := r.FormValue("category")

	var dueDate *time.Time
	dueDateStr := r.FormValue("due_date")
	if dueDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dueDateStr); err == nil {
			dueDate = &parsed
		}
	}

	if !h.store.UpdateTodo(id, text, priority, category, dueDate) {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	todos := h.store.GetTodos()
	for _, todo := range todos {
		if todo.ID == id {
			templateTodo := templates.Todo{
				ID:        todo.ID,
				Text:      todo.Text,
				Priority:  todo.Priority,
				Category:  todo.Category,
				DueDate:   todo.DueDate,
				SortOrder: todo.SortOrder,
				Completed: todo.Completed,
			}
			component := templates.TodoItem(templateTodo)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}
	}
}

func (h *Handlers) ReorderTodosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	todoIDsStr := r.FormValue("todo-ids")
	if todoIDsStr == "" {
		http.Error(w, "Todo IDs are required", http.StatusBadRequest)
		return
	}

	todoIDStrs := strings.Split(todoIDsStr, ",")
	todoIDs := make([]int, len(todoIDStrs))
	
	for i, idStr := range todoIDStrs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			http.Error(w, "Invalid todo ID", http.StatusBadRequest)
			return
		}
		todoIDs[i] = id
	}

	if !h.store.ReorderTodos(todoIDs) {
		http.Error(w, "Failed to reorder todos", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) ExportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("todos_export_%s.json", timestamp)
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	
	todos := h.store.GetTodos()
	todoData := TodoData{
		Todos:  todos,
		NextID: h.store.nextID,
	}
	
	data, err := json.MarshalIndent(todoData, "", "  ")
	if err != nil {
		http.Error(w, "Error generating export", http.StatusInternalServerError)
		return
	}
	
	w.Write(data)
}

func (h *Handlers) SaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.store.SaveToFile(); err != nil {
		http.Error(w, "Error saving todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Todos saved successfully!"))
}