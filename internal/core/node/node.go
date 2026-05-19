package node

import (
	"context"

	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
)

//对于Node的抽象
type Node interface {
	ID() types.NodeID
	Execute(ctx context.Context , input types.Map)(types.Map,error)
}