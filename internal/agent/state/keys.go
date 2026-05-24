package state

import (
	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
	corestate"github.com/looksaw/simple_agent_with_golang/internal/core/state"
)


const (
	MessagesKey corestate.Key[[]llm.Message] = "messages"
)