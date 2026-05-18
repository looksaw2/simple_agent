package main

import "fmt"

/*
* 一个简单的工具注册表
 */

// 工具注册表
type Registry struct {
	Tools map[string]Tool
}

// 初始化
func NewRegistry() *Registry {
	return &Registry{
		Tools: make(map[string]Tool),
	}
}

// 注册对应的表
func (r *Registry) Register(tool Tool) error {
	name := tool.Name()
	if _, ok := r.Tools[name]; ok {
		return fmt.Errorf("%v tool has already exists", name)
	}
	r.Tools[name] = tool
	return nil
}

// 得到对应的工具
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.Tools[name]
	return tool, ok
}

// 得到对应的List列表
func (r *Registry) List() []Tool {
	list := make([]Tool, 0, len(r.Tools))
	for _, tool := range r.Tools {
		list = append(list, tool)
	}
	return list
}

/*
转化成Deepseek认识的格式
[

	{
		"type" : "function",
		"function" : {
			"name" : name ,
			"description" : description,
			"parameters" : parameters
		}
	}
	..................

]
*/
func (r *Registry) ToDeepSeekFormat() []map[string]any {
	tools := make([]map[string]any, 0)
	for _, tool := range r.Tools {
		tools = append(tools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name(),
				"description": tool.Description(),
				"parameters":  tool.InputSchema(),
			},
		})
	}
	return tools
}
