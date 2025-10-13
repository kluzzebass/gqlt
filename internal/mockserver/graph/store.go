package graph

import (
	"fmt"
	"sync"
	"time"

	"github.com/kluzzebass/gqlt/internal/mockserver/graph/model"
)

// EntityStore provides generic CRUD operations for entities using generics
type EntityStore[T any] struct {
	mu       sync.RWMutex
	entities map[string]T
	nextID   int
	typeName string
}

// NewEntityStore creates a new entity store
func NewEntityStore[T any](typeName string) *EntityStore[T] {
	return &EntityStore[T]{
		entities: make(map[string]T),
		nextID:   1,
		typeName: typeName,
	}
}

// Get retrieves an entity by global ID
func (es *EntityStore[T]) Get(id string) (T, bool) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	entity, exists := es.entities[id]
	return entity, exists
}

// GetAll returns all entities
func (es *EntityStore[T]) GetAll() []T {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make([]T, 0, len(es.entities))
	for _, entity := range es.entities {
		result = append(result, entity)
	}
	return result
}

// Create adds a new entity with an auto-generated ID
func (es *EntityStore[T]) Create(entity T) (string, T) {
	es.mu.Lock()
	defer es.mu.Unlock()

	id := fmt.Sprintf("%s:%d", es.typeName, es.nextID)
	es.nextID++
	es.entities[id] = entity
	return id, entity
}

// Update modifies an existing entity
func (es *EntityStore[T]) Update(id string, entity T) bool {
	es.mu.Lock()
	defer es.mu.Unlock()

	if _, exists := es.entities[id]; !exists {
		return false
	}
	es.entities[id] = entity
	return true
}

// Delete removes an entity
func (es *EntityStore[T]) Delete(id string) bool {
	es.mu.Lock()
	defer es.mu.Unlock()

	if _, exists := es.entities[id]; !exists {
		return false
	}
	delete(es.entities, id)
	return true
}

// SubscriberManager manages subscriptions for a specific event type using generics
type SubscriberManager[T any] struct {
	mu               sync.RWMutex
	subscribers      map[int]chan T
	nextSubscriberID int
}

// NewSubscriberManager creates a new subscriber manager
func NewSubscriberManager[T any]() *SubscriberManager[T] {
	return &SubscriberManager[T]{
		subscribers:      make(map[int]chan T),
		nextSubscriberID: 1,
	}
}

// Subscribe registers a channel to receive events
func (sm *SubscriberManager[T]) Subscribe(ch chan T) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := sm.nextSubscriberID
	sm.nextSubscriberID++
	sm.subscribers[id] = ch
	return id
}

// Unsubscribe removes a subscriber
func (sm *SubscriberManager[T]) Unsubscribe(id int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.subscribers, id)
}

// Broadcast sends an event to all active subscribers (non-blocking)
func (sm *SubscriberManager[T]) Broadcast(event T) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, ch := range sm.subscribers {
		select {
		case ch <- event:
			// Event sent successfully
		default:
			// Channel full or closed, skip this subscriber
		}
	}
}

// Store provides thread-safe in-memory storage for the mock server
type Store struct {
	mu sync.RWMutex

	// Entity stores using generics
	users           *EntityStore[*model.User]
	todos           *EntityStore[*model.Todo]
	fileAttachments *EntityStore[*model.FileAttachment]
	linkAttachments *EntityStore[*model.LinkAttachment]

	// Todo-attachment relationships (todoID -> []attachmentID)
	todoAttachments map[string][]string

	// Subscription management using generics
	todoSubscribers *SubscriberManager[*model.Todo]
	userSubscribers *SubscriberManager[*model.User]
}

// NewStore creates a new Store with pre-seeded data
func NewStore() *Store {
	s := &Store{
		users:           NewEntityStore[*model.User]("User"),
		todos:           NewEntityStore[*model.Todo]("Todo"),
		fileAttachments: NewEntityStore[*model.FileAttachment]("FileAttachment"),
		linkAttachments: NewEntityStore[*model.LinkAttachment]("LinkAttachment"),
		todoAttachments: make(map[string][]string),
		todoSubscribers: NewSubscriberManager[*model.Todo](),
		userSubscribers: NewSubscriberManager[*model.User](),
	}

	// Pre-seed with 3 sample users
	s.seedUsers()

	return s
}

// seedUsers creates 3 initial users with different roles
func (s *Store) seedUsers() {
	// Create users using the store's CreateUser function to maintain consistency
	s.CreateUser("Alice Admin", "alice@example.com", model.UserRoleAdmin, strPtr("https://alice.example.com"))
	s.CreateUser("Bob User", "bob@example.com", model.UserRoleUser, nil)
	s.CreateUser("Charlie Guest", "charlie@example.com", model.UserRoleGuest, nil)

	// Adjust CreatedAt timestamps for seeded users to show historical data
	now := time.Now()
	if user, _ := s.GetUser("User:1"); user != nil {
		user.CreatedAt = now.Add(-30 * 24 * time.Hour) // 30 days ago
	}
	if user, _ := s.GetUser("User:2"); user != nil {
		user.CreatedAt = now.Add(-15 * 24 * time.Hour) // 15 days ago
	}
	if user, _ := s.GetUser("User:3"); user != nil {
		user.CreatedAt = now.Add(-7 * 24 * time.Hour) // 7 days ago
	}
}

// === User Methods ===

// GetUser retrieves a user by global ID
func (s *Store) GetUser(id string) (*model.User, error) {
	user, exists := s.users.Get(id)
	if !exists {
		return nil, nil // Not found
	}
	return user, nil
}

// GetUsers returns all users
func (s *Store) GetUsers() []*model.User {
	return s.users.GetAll()
}

// CreateUser adds a new user to the store
func (s *Store) CreateUser(name, email string, role model.UserRole, website *string) *model.User {
	user := &model.User{
		Name:      name,
		Email:     email,
		Role:      role,
		CreatedAt: time.Now(),
		Website:   website,
	}

	id, createdUser := s.users.Create(user)
	createdUser.ID = id
	return createdUser
}

// === Todo Methods ===

// GetTodo retrieves a todo by global ID
func (s *Store) GetTodo(id string) (*model.Todo, error) {
	todo, exists := s.todos.Get(id)
	if !exists {
		return nil, nil // Not found
	}
	return todo, nil
}

// GetTodos returns all todos
func (s *Store) GetTodos() []*model.Todo {
	return s.todos.GetAll()
}

// CreateTodo adds a new todo to the store
func (s *Store) CreateTodo(title string, createdByID string, input *model.CreateTodoInput) *model.Todo {
	now := time.Now()
	todo := &model.Todo{
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

	id, createdTodo := s.todos.Create(todo)
	createdTodo.ID = id
	return createdTodo
}

// UpdateTodo updates an existing todo
func (s *Store) UpdateTodo(id string, input *model.UpdateTodoInput) (*model.Todo, error) {
	todo, exists := s.todos.Get(id)
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
	s.todos.Update(id, todo)
	return todo, nil
}

// DeleteTodo removes a todo from the store
func (s *Store) DeleteTodo(id string) bool {
	return s.todos.Delete(id)
}

// === Attachment Methods ===

// GetFileAttachment retrieves a file attachment by global ID
func (s *Store) GetFileAttachment(id string) (*model.FileAttachment, error) {
	attachment, exists := s.fileAttachments.Get(id)
	if !exists {
		return nil, nil
	}
	return attachment, nil
}

// GetLinkAttachment retrieves a link attachment by global ID
func (s *Store) GetLinkAttachment(id string) (*model.LinkAttachment, error) {
	attachment, exists := s.linkAttachments.Get(id)
	if !exists {
		return nil, nil
	}
	return attachment, nil
}

// CreateFileAttachment adds a new file attachment
func (s *Store) CreateFileAttachment(title, filename, mimeType string, size int) *model.FileAttachment {
	attachment := &model.FileAttachment{
		Title:     title,
		CreatedAt: time.Now(),
		Filename:  filename,
		MimeType:  mimeType,
		Size:      int32(size),
	}

	id, createdAttachment := s.fileAttachments.Create(attachment)
	createdAttachment.ID = id
	return createdAttachment
}

// CreateLinkAttachment adds a new link attachment
func (s *Store) CreateLinkAttachment(title, url string, description *string) *model.LinkAttachment {
	attachment := &model.LinkAttachment{
		Title:       title,
		CreatedAt:   time.Now(),
		URL:         url,
		Description: description,
	}

	id, createdAttachment := s.linkAttachments.Create(attachment)
	createdAttachment.ID = id
	return createdAttachment
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

// ============================================================================
// SUBSCRIPTION MANAGEMENT
// ============================================================================

// SubscribeToTodoEvents registers a channel to receive todo events.
// Returns a subscriber ID that should be used to unsubscribe.
func (s *Store) SubscribeToTodoEvents(ch chan *model.Todo) int {
	return s.todoSubscribers.Subscribe(ch)
}

// UnsubscribeFromTodoEvents removes a todo event subscriber.
func (s *Store) UnsubscribeFromTodoEvents(id int) {
	s.todoSubscribers.Unsubscribe(id)
}

// BroadcastTodoEvent sends a todo to all active subscribers.
// This should be called after any create, update, or delete operation.
func (s *Store) BroadcastTodoEvent(todo *model.Todo) {
	s.todoSubscribers.Broadcast(todo)
}

// SubscribeToUserEvents registers a channel to receive user events.
// Returns a subscriber ID that should be used to unsubscribe.
func (s *Store) SubscribeToUserEvents(ch chan *model.User) int {
	return s.userSubscribers.Subscribe(ch)
}

// UnsubscribeFromUserEvents removes a user event subscriber.
func (s *Store) UnsubscribeFromUserEvents(id int) {
	s.userSubscribers.Unsubscribe(id)
}

// BroadcastUserEvent sends a user to all active subscribers.
// This should be called after any create or update operation.
func (s *Store) BroadcastUserEvent(user *model.User) {
	s.userSubscribers.Broadcast(user)
}

// === Helper Functions ===

func strPtr(s string) *string {
	return &s
}
