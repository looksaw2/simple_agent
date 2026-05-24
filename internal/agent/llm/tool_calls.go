package llm

// 使用工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments"`
}

type ToolCallDelta struct {
	ID             string
	NameDelta      string
	ArgumentsDelta string
}
