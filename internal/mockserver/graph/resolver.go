package graph

import "github.com/kluzzebass/gqlt/internal/mockserver/graph/model"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	todos []*model.Todo
}

// NewResolver creates a new Resolver with initialized fields
func NewResolver() *Resolver {
	return &Resolver{
		todos: []*model.Todo{},
	}
}
