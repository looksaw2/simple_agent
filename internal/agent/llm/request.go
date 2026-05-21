package llm

// 发送LLM的Request的定义
type Request struct {
	Model       string
	Messages    []Message
	Tools       []map[string]any
	Temperature float64
}
