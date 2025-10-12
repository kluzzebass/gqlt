package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	store *Store
}

// NewResolver creates a new Resolver with an initialized data store
func NewResolver() *Resolver {
	return &Resolver{
		store: NewStore(),
	}
}
