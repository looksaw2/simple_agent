package llm

// 接受LLM的Response的定义
type Response struct {
	Messages       Message
	FinishedReason string
}
