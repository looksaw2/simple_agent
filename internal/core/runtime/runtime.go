package runtime

import (
	"context"
	"fmt"

	"github.com/looksaw/simple_agent_with_golang/internal/core/graph"
	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
)

//拓扑图编排好了需要跑在runtime上面

// 这个是一个Runtime，编排好的图跑在上面
type Runtime struct {
	cg *graph.CompiledGraph
}

// 初始化这个Runtime
func NewRuntime(cg *graph.CompiledGraph) *Runtime {
	return &Runtime{
		cg: cg,
	}
}

// 开始事件循环
func (r *Runtime) Invoke(ctx context.Context, initialInput types.Map) (types.Map, error) {
	store := r.cg.NewStore()
	store.Patch(initialInput)
	currentNodeID, err := r.findNextNode(ctx, types.Start, initialInput)
	if err != nil {
		return nil, fmt.Errorf("failed to find init node %v", err)
	}
	for currentNodeID != types.End {
		node, exists := r.cg.GetNode(currentNodeID)
		if !exists {
			return nil, fmt.Errorf("current Node is not exist %v", exists)
		}
		outputPatch, err := node.Execute(ctx, store.Values())
		if err != nil {
			return nil, fmt.Errorf("node [%s] execution failed: %w", currentNodeID, err)
		}
		store.Patch(outputPatch)
		nextNodeID, err := r.findNextNode(ctx, currentNodeID, store.Values())
		if err != nil {
			return nil, fmt.Errorf("routing from [%s] failed: %w", currentNodeID, err)
		}
		currentNodeID = nextNodeID
	}
	return store.Values(), nil
}

// 寻找下一个节点
func (r *Runtime) findNextNode(
	ctx context.Context,
	from types.NodeID,
	currentOutput types.Map) (types.NodeID, error) {
	//优先匹配条件边
	if conditional, exist := r.cg.GetConditionalEdges(from); exist {
		for _, ce := range conditional {
			next, err := ce.ConditionFunc(ctx, currentOutput)
			if err == nil && next != "" {
				return next, nil
			}
		}
	}
	//开始匹配普通的边
	if next, exists := r.cg.GetNormalEdge(from); exists {
		return next, nil
	}
	return types.End, nil
}
