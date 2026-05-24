package nodes

import (
	"context"
	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
	agentstate "github.com/looksaw/simple_agent_with_golang/internal/agent/state"
	corestate "github.com/looksaw/simple_agent_with_golang/internal/core/state"
	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
)

func HasToolCall(ctx context.Context , input types.Map)(types.NodeID,error){
	messages := corestate.GetSafe[
		[]llm.Message,
	](
		input,
		agentstate.MessagesKey,
	)
	//没有消息回到End
	if len(messages) == 0 {
		return types.End, nil
	}
	lastMessage := messages[len(messages)-1]
	if len(lastMessage.ToolCalls) > 0 {
		return "tool", nil
	}
	// 没有 tool call
	return types.End, nil
}
