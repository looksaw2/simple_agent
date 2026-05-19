package tool

import "context"

type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]any
	Call(ctx context.Context, input map[string]any) (string, error)
}
