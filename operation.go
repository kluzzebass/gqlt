package gqlt

import (
	"fmt"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// OperationType represents the type of a GraphQL operation
type OperationType string

const (
	OperationTypeQuery        OperationType = "query"
	OperationTypeMutation     OperationType = "mutation"
	OperationTypeSubscription OperationType = "subscription"
)

// OperationInfo contains information about a GraphQL operation
type OperationInfo struct {
	Type OperationType
	Name string
}

// DetectOperationType parses a GraphQL document and detects the operation type.
// If operationName is provided, it finds that specific operation.
// If operationName is empty and there's only one operation, it uses that one.
// Returns an error if the operation can't be determined or doesn't exist.
func DetectOperationType(query string, operationName string) (*OperationInfo, error) {
	// Parse the GraphQL document without schema validation (syntax only)
	source := &ast.Source{
		Name:  "query",
		Input: query,
	}

	doc, gqlErr := parser.ParseQuery(source)
	if gqlErr != nil {
		return nil, fmt.Errorf("failed to parse GraphQL query: %w", gqlErr)
	}

	// If no operations found
	if len(doc.Operations) == 0 {
		return nil, fmt.Errorf("no operations found in query")
	}

	var targetOp *ast.OperationDefinition

	// If operation name is specified, find it
	if operationName != "" {
		for _, op := range doc.Operations {
			if op.Name == operationName {
				targetOp = op
				break
			}
		}
		if targetOp == nil {
			return nil, fmt.Errorf("operation '%s' not found in query", operationName)
		}
	} else {
		// No operation name specified
		if len(doc.Operations) > 1 {
			return nil, fmt.Errorf("query contains multiple operations, please specify --operation")
		}
		// Use the single operation
		for _, op := range doc.Operations {
			targetOp = op
			break
		}
	}

	// Determine operation type
	var opType OperationType
	switch targetOp.Operation {
	case ast.Query:
		opType = OperationTypeQuery
	case ast.Mutation:
		opType = OperationTypeMutation
	case ast.Subscription:
		opType = OperationTypeSubscription
	default:
		return nil, fmt.Errorf("unknown operation type: %v", targetOp.Operation)
	}

	return &OperationInfo{
		Type: opType,
		Name: targetOp.Name,
	}, nil
}
