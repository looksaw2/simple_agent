package edge

import (
	"context"

	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
)

// 这个是对于边的抽象
type ConditionFunc func(ctx context.Context, currentOutput types.Map) (types.NodeID, error)

//有两者边，一种是无条件转移的边，另一种是有条件转移的边

// 无条件转移的边
type Edge struct {
	From types.NodeID
	To   types.NodeID
}

// 有条件转移的边
type ConditionalEdge struct {
	From          types.NodeID
	ConditionFunc ConditionFunc
}
