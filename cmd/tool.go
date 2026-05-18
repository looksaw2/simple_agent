package main

import "context"

// Tool的定义
type Tool interface {
	//工具的名称
	Name() string
	//工具的描述
	Description() string
	//工具的Input Schema应该是什么
	InputSchema() map[string]any
	//调用这个工具
	Call(ctx context.Context, input map[string]any) (string, error)
}
