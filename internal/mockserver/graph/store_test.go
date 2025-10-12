package graph

import (
	"testing"

	"github.com/kluzzebass/gqlt/internal/mockserver/graph/model"
)

func TestNewStore(t *testing.T) {
	store := NewStore()

	if store == nil {
		t.Fatal("NewStore returned nil")
	}

	// Verify pre-seeded users exist
	users := store.GetUsers()
	if len(users) != 3 {
		t.Errorf("Expected 3 pre-seeded users, got %d", len(users))
	}

	// Verify user IDs follow global ID pattern
	for _, user := range users {
		if len(user.ID) < 6 || user.ID[:5] != "User:" {
			t.Errorf("User ID doesn't follow 'User:' pattern: %s", user.ID)
		}
	}
}

func TestCreateUser(t *testing.T) {
	store := NewStore()

	website := "https://test.com"
	user := store.CreateUser("Test User", "test@example.com", model.UserRoleUser, &website)

	if user == nil {
		t.Fatal("CreateUser returned nil")
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got %s", user.Name)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}

	if user.Role != model.UserRoleUser {
		t.Errorf("Expected role USER, got %v", user.Role)
	}

	// Verify user is in store
	retrieved, err := store.GetUser(user.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Retrieved user ID doesn't match: expected %s, got %s", user.ID, retrieved.ID)
	}
}

func TestCreateTodo(t *testing.T) {
	store := NewStore()

	input := &model.CreateTodoInput{
		Title: "Test Todo",
		Notes: strPtr("Test notes"),
		Tags:  []string{"test", "demo"},
	}

	todo := store.CreateTodo("Test Todo", "User:1", input)

	if todo == nil {
		t.Fatal("CreateTodo returned nil")
	}

	if todo.Title != "Test Todo" {
		t.Errorf("Expected title 'Test Todo', got %s", todo.Title)
	}

	if todo.Status != model.TodoStatusPending {
		t.Errorf("Expected status PENDING, got %v", todo.Status)
	}

	if len(todo.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(todo.Tags))
	}

	// Verify todo is in store
	retrieved, err := store.GetTodo(todo.ID)
	if err != nil {
		t.Fatalf("GetTodo failed: %v", err)
	}

	if retrieved.ID != todo.ID {
		t.Errorf("Retrieved todo ID doesn't match")
	}
}

func TestUpdateTodo(t *testing.T) {
	store := NewStore()

	// Create a todo first
	input := &model.CreateTodoInput{
		Title: "Original Title",
	}
	todo := store.CreateTodo("Original Title", "User:1", input)

	// Update it
	newTitle := "Updated Title"
	newStatus := model.TodoStatusInProgress
	updateInput := &model.UpdateTodoInput{
		ID:     todo.ID,
		Title:  &newTitle,
		Status: &newStatus,
	}

	updated, err := store.UpdateTodo(todo.ID, updateInput)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Title not updated: got %s", updated.Title)
	}

	if updated.Status != model.TodoStatusInProgress {
		t.Errorf("Status not updated: got %v", updated.Status)
	}
}

func TestDeleteTodo(t *testing.T) {
	store := NewStore()

	// Create a todo
	input := &model.CreateTodoInput{
		Title: "To Delete",
	}
	todo := store.CreateTodo("To Delete", "User:1", input)

	// Delete it
	deleted := store.DeleteTodo(todo.ID)
	if !deleted {
		t.Error("DeleteTodo returned false for existing todo")
	}

	// Verify it's gone
	retrieved, _ := store.GetTodo(todo.ID)
	if retrieved != nil {
		t.Error("Todo still exists after deletion")
	}

	// Try deleting again
	deleted = store.DeleteTodo(todo.ID)
	if deleted {
		t.Error("DeleteTodo returned true for non-existent todo")
	}
}

func TestConcurrentAccess(t *testing.T) {
	store := NewStore()

	// Create 10 users concurrently
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			store.CreateUser(
				"User"+string(rune(n)),
				"user"+string(rune(n))+"@example.com",
				model.UserRoleUser,
				nil,
			)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify we have 13 users total (3 seeded + 10 created)
	users := store.GetUsers()
	if len(users) != 13 {
		t.Errorf("Expected 13 users after concurrent creation, got %d", len(users))
	}
}

