package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

type Todo struct {
	ID        int        `json:"id"`
	Text      string     `json:"text"`
	Priority  string     `json:"priority"`
	Category  string     `json:"category"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	SortOrder int        `json:"sort_order"`
	Completed bool       `json:"completed"`
	CreatedAt time.Time  `json:"created_at"`
}

type TodoStore struct {
	mu       sync.RWMutex
	todos    []Todo
	nextID   int
	filename string
}

func NewTodoStore() *TodoStore {
	store := &TodoStore{
		todos:    make([]Todo, 0),
		nextID:   1,
		filename: "todos.json",
	}
	store.LoadFromFile()
	return store
}

func (ts *TodoStore) AddTodo(text, priority, category string, dueDate *time.Time) Todo {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	sortOrder := len(ts.todos)
	todo := Todo{
		ID:        ts.nextID,
		Text:      text,
		Priority:  priority,
		Category:  category,
		DueDate:   dueDate,
		SortOrder: sortOrder,
		Completed: false,
		CreatedAt: time.Now(),
	}
	
	ts.todos = append(ts.todos, todo)
	ts.nextID++
	ts.saveToFileAsync()
	
	return todo
}

func (ts *TodoStore) GetTodos() []Todo {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	todosCopy := make([]Todo, len(ts.todos))
	copy(todosCopy, ts.todos)
	
	sort.Slice(todosCopy, func(i, j int) bool {
		return todosCopy[i].SortOrder < todosCopy[j].SortOrder
	})
	
	return todosCopy
}

func (ts *TodoStore) ToggleTodo(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	for i, todo := range ts.todos {
		if todo.ID == id {
			ts.todos[i].Completed = !ts.todos[i].Completed
			ts.saveToFileAsync()
			return true
		}
	}
	return false
}

func (ts *TodoStore) DeleteTodo(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	for i, todo := range ts.todos {
		if todo.ID == id {
			ts.todos = append(ts.todos[:i], ts.todos[i+1:]...)
			ts.saveToFileAsync()
			return true
		}
	}
	return false
}

func (ts *TodoStore) UpdateTodo(id int, text, priority, category string, dueDate *time.Time) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	for i, todo := range ts.todos {
		if todo.ID == id {
			ts.todos[i].Text = text
			ts.todos[i].Priority = priority
			ts.todos[i].Category = category
			ts.todos[i].DueDate = dueDate
			ts.saveToFileAsync()
			return true
		}
	}
	return false
}

func (ts *TodoStore) ReorderTodos(todoIDs []int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	if len(todoIDs) != len(ts.todos) {
		return false
	}
	
	todoMap := make(map[int]*Todo)
	for i := range ts.todos {
		todoMap[ts.todos[i].ID] = &ts.todos[i]
	}
	
	for i, id := range todoIDs {
		if todo, exists := todoMap[id]; exists {
			todo.SortOrder = i
		} else {
			return false
		}
	}
	
	ts.saveToFileAsync()
	return true
}

type TodoData struct {
	Todos  []Todo `json:"todos"`
	NextID int    `json:"next_id"`
}

func (ts *TodoStore) LoadFromFile() error {
	if _, err := os.Stat(ts.filename); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(ts.filename)
	if err != nil {
		log.Printf("Error reading todos file: %v", err)
		return err
	}

	var todoData TodoData
	if err := json.Unmarshal(data, &todoData); err != nil {
		log.Printf("Error parsing todos file: %v", err)
		return err
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	ts.todos = todoData.Todos
	ts.nextID = todoData.NextID
	
	if ts.nextID <= 0 {
		ts.nextID = 1
	}
	
	if len(ts.todos) > 0 {
		maxID := 0
		for _, todo := range ts.todos {
			if todo.ID > maxID {
				maxID = todo.ID
			}
		}
		if maxID >= ts.nextID {
			ts.nextID = maxID + 1
		}
	}

	log.Printf("Loaded %d todos from file", len(ts.todos))
	return nil
}

func (ts *TodoStore) SaveToFile() error {
	ts.mu.RLock()
	todoData := TodoData{
		Todos:  make([]Todo, len(ts.todos)),
		NextID: ts.nextID,
	}
	copy(todoData.Todos, ts.todos)
	ts.mu.RUnlock()

	data, err := json.MarshalIndent(todoData, "", "  ")
	if err != nil {
		log.Printf("Error marshaling todos: %v", err)
		return err
	}

	if err := os.WriteFile(ts.filename, data, 0644); err != nil {
		log.Printf("Error writing todos file: %v", err)
		return err
	}

	return nil
}

func (ts *TodoStore) saveToFileAsync() {
	go func() {
		if err := ts.SaveToFile(); err != nil {
			log.Printf("Failed to save todos to file: %v", err)
		}
	}()
}

func (ts *TodoStore) ExportToFile(filename string) error {
	ts.mu.RLock()
	todoData := TodoData{
		Todos:  make([]Todo, len(ts.todos)),
		NextID: ts.nextID,
	}
	copy(todoData.Todos, ts.todos)
	ts.mu.RUnlock()

	data, err := json.MarshalIndent(todoData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling todos: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing todos file: %w", err)
	}

	return nil
}