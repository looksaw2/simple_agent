package llm

// 这些信息是谁发出的
type Role string

// 一下包含4种角色
const (
	//System最高角色
	SystemRole Role = "system"
	//用户角色
	UserRole Role = "user"
	//助手的角色
	AssistantRole Role = "assistant"
	//Tool的角色
	TooRole Role = "tool"
)

// 然后对应的消息的抽象
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// 使用工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"argments"`
}
