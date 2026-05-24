package deepseek

import "github.com/looksaw/simple_agent_with_golang/internal/agent/llm"

// 将Deepseek的模式转化为标准的格式
type DeepSeekChatRequest struct {
	Model       string                `json:"model"`
	Messages    []DeepSeekChatMessage `json:"messages"`
	Tools       []map[string]any      `json:"tools,omitempty"`
	Temperature float64               `json:"temperature,omitempty"`
	//是否使用Stream流式传输
	Stream bool `json:"stream,omitempty"`
}

// DeepSeek的Response
type DeepSeekChatResponse struct {
	Choices []struct {
		Message      DeepSeekChatMessage `json:"message"`
		FinishReason string              `json:"finish_reason"`
	} `json:"choices"`
}

// DeepSeek的Message定义
type DeepSeekChatMessage struct {
	Role             string             `json:"role"`
	Content          string             `json:"content,omitempty"`
	ReasoningContent string             `json:"reasoning_content,omitempty"`
	ToolCalls        []DeepSeekToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string             `json:"tool_call_id,omitempty"`
	Name             string             `json:"name,omitempty"`
}
type DeepSeekToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function DeepSeekFunction `json:"function"`
}
type DeepSeekFunction struct {
	Name      string            `json:"name"`
	Arguments string `json:"arguments"`
}

// Deepseek的流式传输
type DeepSeekStreamResponse struct {
	ID      string                 `json:"id"`
	Created int64                  `json:"created"`
	Object  string                 `json:"object"`
	Model   string                 `json:"model"`
	Choices []DeepSeekStreamChoice `json:"choices"`
}

// DeepSeek的流式Choice
type DeepSeekStreamChoice struct {
	Index        int           `json:"index"`
	Delta        DeepSeekDelta `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
}

// DeepSeek的Delta
type DeepSeekDelta struct {
	Role      string                   `json:"role,omitempty"`
	Content   string                   `json:"content,omitempty"`
	ToolCalls []DeepSeekToolCallsDelta `json:"tool_calls,omitempty"`
}

// ToolCallsDelta
type DeepSeekToolCallsDelta struct {
	ID       string                `json:"id"`
	Type     string                `json:"type"`
	Function DeepSeekFunctionDelta `json:"function"`
}

// DeepSeek的函数Delta
type DeepSeekFunctionDelta struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// 将标准的Request转化为DeepSeek的Request
func toDeepSeekRequest(
	req *llm.Request,
	defaultModel string,
) *DeepSeekChatRequest {
	model := req.Model
	if model == "" {
		model = defaultModel
	}
	msgs := make([]DeepSeekChatMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, toDeepSeekMessage(m))
	}
	return &DeepSeekChatRequest{
		Model:       model,
		Messages:    msgs,
		Tools:       req.Tools,
		Temperature: req.Temperature,
	}
}
func toDeepSeekMessage(
	m llm.Message,
) DeepSeekChatMessage {
	toolCalls := make(
		[]DeepSeekToolCall,
		0,
		len(m.ToolCalls),
	)
	for _, tc := range m.ToolCalls {
		toolCalls = append(
			toolCalls,
			toDeepSeekToolCall(tc),
		)
	}
	return DeepSeekChatMessage{
		Role:             string(m.Role),
		Content:          m.Content,
		ReasoningContent: m.ReasoningContent,
		ToolCalls:        toolCalls,
		ToolCallID:       m.ToolCallID,
		Name:             m.Name,
	}
}
func toDeepSeekToolCall(
	tc llm.ToolCall,
) DeepSeekToolCall {
	return DeepSeekToolCall{
		ID:   tc.ID,
		Type: tc.Type,
		Function: DeepSeekFunction{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		},
	}
}
func fromDeepSeekResponse(
	resp *DeepSeekChatResponse,
) *llm.Response {

	if len(resp.Choices) == 0 {
		return &llm.Response{}
	}
	choice := resp.Choices[0]
	return &llm.Response{
		Message:        fromDeepSeekMessage(choice.Message),
		FinishedReason: choice.FinishReason,
	}
}
func fromDeepSeekMessage(
	m DeepSeekChatMessage,
) llm.Message {
	toolCalls := make(
		[]llm.ToolCall,
		0,
		len(m.ToolCalls),
	)
	for _, tc := range m.ToolCalls {
		toolCalls = append(
			toolCalls,
			fromDeepSeekToolCall(tc),
		)
	}

	return llm.Message{
		Role:             llm.Role(m.Role),
		Content:          m.Content,
		ReasoningContent: m.ReasoningContent,
		ToolCalls:        toolCalls,
		ToolCallID:       m.ToolCallID,
		Name:             m.Name,
	}
}
func fromDeepSeekToolCall(
	tc DeepSeekToolCall,
) llm.ToolCall {

	return llm.ToolCall{
		ID:   tc.ID,
		Type: tc.Type,
		Function: llm.FunctionCall{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		},
	}
}
