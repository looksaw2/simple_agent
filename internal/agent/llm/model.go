package llm

import "context"

// 模型的公共定义
type Model interface {
	//发送消息
	Generate(ctx context.Context, req *Request) (*Response, error)
	//发送流式消息
	Stream(ctx context.Context, req *Request) (<-chan Chunk, error)
}

// Chunk的定义
type Chunk struct {
	ContentDelta  string
	ToolCallDelat *ToolCall
}
