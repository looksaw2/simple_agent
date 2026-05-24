package nodes

import (
	"context"

	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
	"github.com/looksaw/simple_agent_with_golang/internal/core/state"
	agnetstate "github.com/looksaw/simple_agent_with_golang/internal/agent/state"
	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
	"github.com/looksaw/simple_agent_with_golang/internal/tool"
)

//这个是LLM的节点
type LLMNode struct {
	id types.NodeID
	model llm.Model
	systemPrompt string
	tools *tool.Registry
}
//初始化对应的LLM
func NewLLMNode(
	id types.NodeID,
	model llm.Model,
	systemPrompt string,
	tools *tool.Registry,
) *LLMNode {
	return &LLMNode{
		id: id,
		model: model,
		systemPrompt: systemPrompt,
		tools: tools,
	}
}
//得到具体的ID
func(n *LLMNode)ID() types.NodeID {
	return n.id
}

//llm节点的执行
func(n *LLMNode)Execute(ctx context.Context , input types.Map)(types.Map,error){
	//从节点里面取出来对应的信息
	messages := state.GetSafe[ [] llm.Message](
		input,
		agnetstate.MessagesKey,
	)
	//开始组装SystemPrompt
	if n.systemPrompt != "" {
		hasSystemPrompt := false
		for _ , msg := range messages {
			if msg.Role == llm.SystemRole {
				hasSystemPrompt = true
				break
			}
		}
		if !hasSystemPrompt {
			systemMessage := llm.Message{
				Role: llm.SystemRole,
				Content: n.systemPrompt,
			}
			messages = append(
				[]llm.Message{systemMessage},
				messages...,
			)
		}
	}
	// 构造 request
	req := &llm.Request{
		Messages: messages,
	}
	// 注入 tools schema
	if n.tools != nil {
		req.Tools = n.tools.ToDeepSeekFormat()
	}
	// 调用模型
	resp, err := n.model.Generate(
		ctx,
		req,
	)
	if err != nil {
		return nil, err
	}
	// append assistant message
	messages = append(
		messages,
		resp.Message,
	)
	// 返回 patch
	return types.Map{
		string(agnetstate.MessagesKey): messages,
	}, nil
}