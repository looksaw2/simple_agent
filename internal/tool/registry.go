package tool

import "fmt"

type Registry struct {
	Tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		Tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) error {
	name := t.Name()
	if _, ok := r.Tools[name]; ok {
		return fmt.Errorf("%v tool has already exists", name)
	}
	r.Tools[name] = t
	return nil
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.Tools[name]
	return t, ok
}

func (r *Registry) List() []Tool {
	list := make([]Tool, 0, len(r.Tools))
	for _, t := range r.Tools {
		list = append(list, t)
	}
	return list
}

func (r *Registry) ToDeepSeekFormat() []map[string]any {
	tools := make([]map[string]any, 0)
	for _, t := range r.Tools {
		tools = append(tools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name(),
				"description": t.Description(),
				"parameters":  t.InputSchema(),
			},
		})
	}
	return tools
}
