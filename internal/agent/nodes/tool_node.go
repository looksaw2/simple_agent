package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
	agentstate "github.com/looksaw/simple_agent_with_golang/internal/agent/state"
	corestate "github.com/looksaw/simple_agent_with_golang/internal/core/state"
	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
	"github.com/looksaw/simple_agent_with_golang/internal/tool"
)

//工具节点的定义
type ToolNode struct {
	id types.NodeID
	registry *tool.Registry
}
func NewToolNode(id types.NodeID, registry *tool.Registry) *ToolNode {
	return &ToolNode{
		id: id,
		registry: registry,
	}
}

//得到具体的ID
func(n *ToolNode)ID() types.NodeID {
	return n.id
}

func (n *ToolNode) Execute(
	ctx context.Context,
	input types.Map,
)(
	types.Map,
	error,
) {
	// 取 messages
	messages := corestate.GetSafe[
		[]llm.Message,
	](
		input,
		agentstate.MessagesKey,
	)

	if len(messages) == 0 {
		return nil, fmt.Errorf("messages is empty")
	}
	// 最后一个 message
	lastMessage := messages[len(messages)-1]
	// 必须有 tool call
	if len(lastMessage.ToolCalls) == 0 {
		return nil, fmt.Errorf("assistant message has no tool calls")
	}
	// 当前先只处理第一个 tool call
	toolCall := lastMessage.ToolCalls[0]
	toolImpl, exists := n.registry.Get(
		toolCall.Function.Name,
	)
	if !exists {
		return nil, fmt.Errorf(
			"tool [%s] not found",
			toolCall.Function.Name,
		)
	}
	var args map[string]any
	if err := json.Unmarshal(
		[]byte(toolCall.Function.Arguments),
		&args,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to parse tool arguments: %w",
			err,
		)
	}
	result, err := toolImpl.Call(
		ctx,
		args,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"tool execution failed: %w",
			err,
		)
	}
	// 构造 tool message
	toolMessage := llm.Message{
		Role: llm.ToolRole,
		Content: result,
		ToolCallID: toolCall.ID,
		Name: toolCall.Function.Name,
	}
	messages = append(
		messages,
		toolMessage,
	)
	return types.Map{
		string(agentstate.MessagesKey): messages,
	}, nil
}