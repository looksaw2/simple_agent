package llm

import "context"

// 模型的公共定义
type Model interface {
	//发送消息
	Generate(ctx context.Context, req *Request) (*Response, error)
}

// Chunk的定义
type Chunk struct {
	RoleDelta     string
	ContentDelta  string
	ToolCallDelta *ToolCallDelta
	FinishReason  *string
}

// Stream的定义
type Stream interface {
	Recv() (*Chunk, error)
	Close() error
}

type StreamingModel interface {
	Model
	Stream(ctx context.Context, req *Request) (Stream, error)
}
