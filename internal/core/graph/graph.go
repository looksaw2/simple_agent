package graph

import (
	"fmt"
	"github.com/looksaw/simple_agent_with_golang/internal/core/edge"
	"github.com/looksaw/simple_agent_with_golang/internal/core/node"
	"github.com/looksaw/simple_agent_with_golang/internal/core/state"
	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
)

//定义图的拓扑和静态的组装
type StateGraph struct {
	//定义图的节点
	nodes map[types.NodeID]node.Node
	//定义图的无条件边
	edge map[types.NodeID]types.NodeID
	//定义有条件才转移的边
	conditionalEdge map[types.NodeID][]edge.ConditionalEdge
	//定义Reducer
	reducers map[string]state.Reducer
}

//初始化这个StateGraph
func NewStateGraph(reducers map[string]state.Reducer) *StateGraph {
	return &StateGraph{
		nodes: make(map[types.NodeID]node.Node),
		edge: make(map[types.NodeID]types.NodeID),
		conditionalEdge: make(map[types.NodeID][]edge.ConditionalEdge),
		reducers: reducers,
	}
}

//然后开始添加三个nodes，edge，conditionalEdge这三个需要被动态的添加进来
func(g *StateGraph)AddNode(node node.Node){
	g.nodes[node.ID()] = node
}
//然后添加无条件的边
func(g *StateGraph)AddEdge(from types.NodeID , to types.NodeID){
	g.edge[from] = to
}
//然后添加有条件的边
func(g *StateGraph)AddConditionalEdge(from types.NodeID , conditionalFunc edge.ConditionFunc){
	ce := edge.ConditionalEdge{
		From: from,
		ConditionFunc: conditionalFunc,
	}
	g.conditionalEdge[from] = append(g.conditionalEdge[from], ce)
}


/*
*
*	因为Agent的拓扑图在定义好之后应该不能随意的变动，所以抽象出一个CompiledGraph，将这个CompiledGraph
*   和StateGrapsh定义好之后隔离起来，防止对于StateGraph造成修改
*
* 
*/
//开始编译化整个状态图
func(g *StateGraph)Compile()(*CompiledGraph,error){
	if _ , hasStart := g.nodes[types.Start]; !hasStart {
		if _, hasCondStart := g.conditionalEdge[types.Start]; !hasCondStart {
			return nil, fmt.Errorf("graph must have a starting edge from types.START")
		}
	}
	return &CompiledGraph{sg: g} , nil
}



//对于死图的需要排除
type CompiledGraph struct {
	sg *StateGraph
}


//然后有得到对应的节点
func(cg *CompiledGraph)GetNode(id types.NodeID)(node.Node,bool){
	node , exist := cg.sg.nodes[id]
	return node , exist
}
//添加对应的存储
func(cg *CompiledGraph)NewStore() *state.Store {
	return state.NewStore(cg.sg.reducers)
}
//
func (cg *CompiledGraph) GetNormalEdge(from types.NodeID) (types.NodeID, bool) {
	to, exists := cg.sg.edge[from]
	return to, exists
}
//
func (cg *CompiledGraph) GetConditionalEdges(from types.NodeID) ([]edge.ConditionalEdge, bool) {
	edges, exists := cg.sg.conditionalEdge[from]
	return edges, exists
}