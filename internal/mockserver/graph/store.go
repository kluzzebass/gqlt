package graph

import (
	"fmt"
	"sync"
	"time"

	"github.com/kluzzebass/gqlt/internal/mockserver/graph/model"
)

// Store provides thread-safe in-memory storage for the mock server
type Store struct {
	mu sync.RWMutex

	// Entity stores with global ID format "TypeName:localId"
	users           map[string]*model.User
	todos           map[string]*model.Todo
	fileAttachments map[string]*model.FileAttachment
	linkAttachments map[string]*model.LinkAttachment

	// Todo-attachment relationships (todoID -> []attachmentID)
	todoAttachments map[string][]string

	// ID counters
	nextUserID           int
	nextTodoID           int
	nextFileAttachmentID int
	nextLinkAttachmentID int
}

// NewStore creates a new Store with pre-seeded data
func NewStore() *Store {
	s := &Store{
		users:                make(map[string]*model.User),
		todos:                make(map[string]*model.Todo),
		fileAttachments:      make(map[string]*model.FileAttachment),
		linkAttachments:      make(map[string]*model.LinkAttachment),
		todoAttachments:      make(map[string][]string),
		nextUserID:           1,
		nextTodoID:           1,
		nextFileAttachmentID: 1,
		nextLinkAttachmentID: 1,
	}

	// Pre-seed with 3 sample users
	s.seedUsers()

	return s
}

// seedUsers creates 3 initial users with different roles
func (s *Store) seedUsers() {
	now := time.Now()

	users := []*model.User{
		{
			ID:        "User:1",
			Name:      "Alice Admin",
			Email:     "alice@example.com",
			Role:      model.UserRoleAdmin,
			CreatedAt: now.Add(-30 * 24 * time.Hour), // 30 days ago
			Website:   strPtr("https://alice.example.com"),
		},
		{
			ID:        "User:2",
			Name:      "Bob User",
			Email:     "bob@example.com",
			Role:      model.UserRoleUser,
			CreatedAt: now.Add(-15 * 24 * time.Hour), // 15 days ago
			Website:   nil,
		},
		{
			ID:        "User:3",
			Name:      "Charlie Guest",
			Email:     "charlie@example.com",
			Role:      model.UserRoleGuest,
			CreatedAt: now.Add(-7 * 24 * time.Hour), // 7 days ago
			Website:   nil,
		},
	}

	for _, user := range users {
		s.users[user.ID] = user
	}
	s.nextUserID = 4
}

// === User Methods ===

// GetUser retrieves a user by global ID
func (s *Store) GetUser(id string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, nil // Not found
	}
	return user, nil
}

// GetUsers returns all users
func (s *Store) GetUsers() []*model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*model.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

// CreateUser adds a new user to the store
func (s *Store) CreateUser(name, email string, role model.UserRole, website *string) *model.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("User:%d", s.nextUserID)
	s.nextUserID++

	user := &model.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Role:      role,
		CreatedAt: time.Now(),
		Website:   website,
	}

	s.users[id] = user
	return user
}

// === Todo Methods ===

// GetTodo retrieves a todo by global ID
func (s *Store) GetTodo(id string) (*model.Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, nil // Not found
	}
	return todo, nil
}

// GetTodos returns all todos
func (s *Store) GetTodos() []*model.Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]*model.Todo, 0, len(s.todos))
	for _, todo := range s.todos {
		todos = append(todos, todo)
	}
	return todos
}

// CreateTodo adds a new todo to the store
func (s *Store) CreateTodo(title string, createdByID string, input *model.CreateTodoInput) *model.Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("Todo:%d", s.nextTodoID)
	s.nextTodoID++

	now := time.Now()
	todo := &model.Todo{
		ID:        id,
		Title:     title,
		Notes:     input.Notes,
		Status:    model.TodoStatusPending,
		Priority:  model.TodoPriorityNormal,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      []string{},
	}

	if input.Priority != nil {
		todo.Priority = *input.Priority
	}
	if input.Tags != nil {
		todo.Tags = input.Tags
	}
	if input.DueDate != nil {
		todo.DueDate = input.DueDate
	}

	s.todos[id] = todo
	return todo
}

// UpdateTodo updates an existing todo
func (s *Store) UpdateTodo(id string, input *model.UpdateTodoInput) (*model.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	// Apply updates
	if input.Title != nil {
		todo.Title = *input.Title
	}
	if input.Notes != nil {
		todo.Notes = input.Notes
	}
	if input.Status != nil {
		todo.Status = *input.Status
	}
	if input.Priority != nil {
		todo.Priority = *input.Priority
	}
	if input.DueDate != nil {
		todo.DueDate = input.DueDate
	}
	if input.Tags != nil {
		todo.Tags = input.Tags
	}

	todo.UpdatedAt = time.Now()
	return todo, nil
}

// DeleteTodo removes a todo from the store
func (s *Store) DeleteTodo(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.todos[id]; !exists {
		return false
	}

	delete(s.todos, id)
	return true
}

// === Attachment Methods ===

// GetFileAttachment retrieves a file attachment by global ID
func (s *Store) GetFileAttachment(id string) (*model.FileAttachment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attachment, exists := s.fileAttachments[id]
	if !exists {
		return nil, nil
	}
	return attachment, nil
}

// GetLinkAttachment retrieves a link attachment by global ID
func (s *Store) GetLinkAttachment(id string) (*model.LinkAttachment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attachment, exists := s.linkAttachments[id]
	if !exists {
		return nil, nil
	}
	return attachment, nil
}

// CreateFileAttachment adds a new file attachment
func (s *Store) CreateFileAttachment(title, filename, mimeType string, size int) *model.FileAttachment {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("FileAttachment:%d", s.nextFileAttachmentID)
	s.nextFileAttachmentID++

	attachment := &model.FileAttachment{
		ID:        id,
		Title:     title,
		CreatedAt: time.Now(),
		Filename:  filename,
		MimeType:  mimeType,
		Size:      int32(size),
	}

	s.fileAttachments[id] = attachment
	return attachment
}

// CreateLinkAttachment adds a new link attachment
func (s *Store) CreateLinkAttachment(title, url string, description *string) *model.LinkAttachment {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("LinkAttachment:%d", s.nextLinkAttachmentID)
	s.nextLinkAttachmentID++

	attachment := &model.LinkAttachment{
		ID:          id,
		Title:       title,
		CreatedAt:   time.Now(),
		URL:         url,
		Description: description,
	}

	s.linkAttachments[id] = attachment
	return attachment
}

// AddAttachmentToTodo links an attachment to a todo
func (s *Store) AddAttachmentToTodo(todoID, attachmentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.todoAttachments[todoID] = append(s.todoAttachments[todoID], attachmentID)
}

// RemoveAttachmentFromTodo unlinks an attachment from a todo
func (s *Store) RemoveAttachmentFromTodo(todoID, attachmentID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	attachments := s.todoAttachments[todoID]
	for i, id := range attachments {
		if id == attachmentID {
			s.todoAttachments[todoID] = append(attachments[:i], attachments[i+1:]...)
			return true
		}
	}
	return false
}

// === Helper Functions ===

func strPtr(s string) *string {
	return &s
}
