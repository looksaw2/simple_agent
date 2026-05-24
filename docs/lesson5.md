### lesson 5
在lesson 4中我们实现了Stream流式的传输，但是从agent层到core层之间还缺少了一些必要的东西，所以和之前一样的，我们开始构建一些简单的节点，将对应的业务节点和core里面的节点联系起来，完成对应的一个简单的agent。我们首先得想一想我们需要什么样子的节点，从最简单的视角出发，我们的架构中牵扯最多的就是LLM和Tool之间的交互，因此，在这个前期的规划阶段，我们可以目前简单的抽象出两个节点，即LLMNode和ToolNode，至于节点的数量肯定远远不止这么点，但是由于我们正在起步阶段因此，在目前的视角下面先抽象和实现这两个节点。

首先让我们来思考LLNode的具体的功能是什么？如下图所示，一般来说LLMNode干的事情如下:
![](lesson5__llmNode_graph)
通过观察上面的LLM节点的工作特征我们可以抽象出三个LLMNode节点的能力:
|  序号 | 功能 |
| ---   | --- |
|  1 | 接受其他地方传过来的信息 |
|  2 | 发送给对应的LLM，收到信息处理 |
|  3 | 将这个信息发往别处  |

然后是ToolNode，他的具体的功能是什么？ToolNode节点的特性可以说是执行命令然后输出结果，其能力可以分为如下的图所示:

![](lesson4__toolNode_graph)
所以说可以抽象出如下的ToolNode的能力:
| 序号 | 能力 |
| ---  | ---  |
| 1 | 接受其他地方传过来的信息 |
| 2 | 寻找对应的工具执行 |
| 3 | 输出对应的结果 |


然后了解了上面的基本信息之后我们可以抽象出来如下的节点的golang的实现，当然目前还是比较的不完善，但之后可以迭代
```go
type LLMNode struct {
	id types.NodeID
	model llm.Model
	systemPrompt string
	tools *tool.Registry
}
```
之后将上面想到的LLMNode三种能力做一个实现，基本还是这三步其具体的golang的实现的代码如下：
```go
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
	messages = append(
		messages,
		resp.Message,
	)
	// 返回 patch
	return types.Map{
		string(agnetstate.MessagesKey): messages,
	}, nil
}
```

之后便是ToolNode节点，和之前LLM节点类似，在Tool节点我们也和之前LLMNode节点实现一致，实现对应的tool_node.go，当然对于一般的工具执行来说，可能会存在许多不可控的失败，但目前让我们先简单一点，先搓出来一个能用的Agent，之后再去考虑对于异常情况的处理，但不代表这个不重要，就像有人精辟的总结一样Agent是一个在不可靠的基座上面搭建一个可靠的系统，LLM是一定会出问题的，问题在于出了问题之后整个Agent系统能不能兜底。下面便是ToolNode的实现

```go
//工具节点的定义
type ToolNode struct {
	id types.NodeID
	registry *tool.Registry
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
```
之后还有跳转逻辑，即节点LLMNode怎跳到对应的ToolNode的节点以及之后ToolNode怎么跳回来的逻辑，由于之前在core文件夹下面实现了有条件的跳转因此我们还需要一个condition.go为了简单起见我们写一个简单的conditions.go
```go
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
```
最后是选择的模式，通常最后的成品会根据节点的状态来自动的选择对应的模式但是为了简单起见，我们使用了react模式和planner模式，其他的模式会在之后更新的lesson里面继续添加，但是就目前来说使用者两种模式,下面是react模式的具体的实现
```go
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
```
在这里可以看出我们之前的努力都是值得的，在这个需要的只是去定义这个图，然后agent的框架就搭建好了，在以后如果有时间的话甚至可以使用ts+react做一个前端然后用户通过拖曳组件来构建对应的agent流或者agent自进化生成这个图，当然这些都是后话，之后便是main.go的入口的实现。作为一个最小实现的agent助手，由于在lesson 5的时候还是没有前端，故目前把输入写死，测试一下对应的其他代码有无问题，其具体的代码如下:
```go
func main() {
	ctx := context.Background()
	//使用你自己的APIKEY
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		panic("DEEP SEEK API KEY must be valid......")
	}
	//初始化工具注册表
	registry := tool.NewRegistry()
	//注册一个内建工具(目前只有Bash)
	if err := registry.Register(builtin.NewBashTool()); err != nil {
		log.Fatal("failed to registr bash tool %v",err)
	}
	//初始化模型
	model := deepseek.NewClient(
		apiKey,
		deepseek.DEEPSEEKBASEURL,
		deepseek.MODEL,
		nil,
	)
	//构建Graph图
	cg ,err := builders.BuildReAct(
		model,
		registry,
		"You are an assitant, help me to do some thing useful",
	)
	if err != nil {
		log.Fatal("create graph is failed ..... %v",err)
	}
	//装载runtime
	rt := runtime.NewRuntime(cg)
	input := types.Map{
		string(agentstate.MessagesKey): []llm.Message{
			{
				Role: llm.UserRole,
				Content: "use bash tool to list current files",
			},
		},
	}
	//开始执行
	result , err := rt.Invoke(
		ctx,
		input,
	)
	if err != nil {
		log.Fatalf("invoke failed: %v", err)
	}
	//结果输出(这段AI生成的)
	messages := result[string(agentstate.MessagesKey)].([]llm.Message)
	fmt.Println("===== FINAL MESSAGES =====")
	for _, msg := range messages {
		fmt.Printf(
			"\n[%s]\n%s\n",
			msg.Role,
			msg.Content,
		)
		if len(msg.ToolCalls) > 0 {
			fmt.Println("Tool Calls:")
			for _, tc := range msg.ToolCalls {
				fmt.Printf(
					"- %s(%s)\n",
					tc.Function.Name,
					tc.Function.Arguments,
				)
			}
		}
	}
}
```
然后运行source .env && go run ./cmd/ 发现运行的结果为
```
[assistant]
当前目录下的文件和文件夹如下：

| 名称 | 类型 | 大小 |
|------|------|------|
| `.claude/` | 文件夹 | - |
| `.env` | 文件 | 59 B |
| `.git/` | 文件夹 | - |
| `.gitignore` | 文件 | 4 B |
| `Makefile` | 文件 | 65 B |
| `README.md` | 文件 | 1072 B |
| `cmd/` | 文件夹 | - |
| `docs/` | 文件夹 | - |
| `go.mod` | 文件 | 62 B |
| `img/` | 文件夹 | - |
| `internal/` | 文件夹 | - |
| `scripts/` | 文件夹 | - |
| `web/` | 文件夹 | - |

可以看到这是一个 **Go 语言项目**（有 `go.mod`、`cmd/`、`internal/` 等目录结构），同时包含 `Makefile`、`scripts/`、`web/`、`docs/` 等常见的项目目录。
```
可以发现我们做的这个agent已经运行起来了，项目处于一个初步可用的状态.