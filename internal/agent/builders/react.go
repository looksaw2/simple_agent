package builders

import (
	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
	"github.com/looksaw/simple_agent_with_golang/internal/agent/nodes"
	"github.com/looksaw/simple_agent_with_golang/internal/core/graph"
	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
	"github.com/looksaw/simple_agent_with_golang/internal/tool"
)


func BuildReAct(
	model llm.Model,
	registry *tool.Registry,
	systemPrompt string,
)(*graph.CompiledGraph , error){
	g := graph.NewStateGraph(nil)
	//初始化llmNode和ToolNode
	llmNode := nodes.NewLLMNode(
		"llm",
		model,
		systemPrompt,
		registry,
	)
	//初始化ToolNode
	toolNode := nodes.NewToolNode(
		"tool",
		registry,
	)
	g.AddNode(llmNode)
	g.AddNode(toolNode)
	//然后从Start节点到LLMNode边
	g.AddEdge(
		types.Start,
		llmNode.ID(),
	)
	//然后添加LLMNode到Tool的边
	g.AddConditionalEdge(
		llmNode.ID(),
		nodes.HasToolCall,
	)
	//LLMNode到Tool的边
	g.AddEdge(
		toolNode.ID(),
		llmNode.ID(),
	)
	return g.Compile()
}