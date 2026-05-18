### Lesson 1

最近 AI Agent 非常火，为了弄清楚它的原理和运作方式，我会一步步从一个最小的可运行程序开始，逐步搭建一个具备初步功能的 Agent。

#### 思考：Agent 的最小定义是什么？

首先思考一个最简 Agent 需要具备什么能力：

1. 接收用户输入
2. 将输入发送给大模型
3. 根据大模型的返回内容，决定是继续调用工具，还是直接结束

基于上面的思考，我们可以抽象出如下状态机：

![状态转移](../img/lesson1_state_transform.svg)

Agent 从 `Start` 状态开始，随后转移到 `Thinking` 状态。`Thinking` 状态接收 LLM 的输出，决定下一步转移方向：

- **转向 Tool**：LLM 返回了工具调用请求，执行对应的工具
- **转向 Finish**：LLM 直接返回了最终回答，对话结束

搞清楚状态转移后，就可以开始写代码了。首先梳理我们需要的能力：

```
1. 收发 DeepSeek 消息（HTTP Client + JSON 序列化）
2. Bash 脚本执行（Tool）
```

---

#### 第一步：构建 DeepSeek Client

发送请求所需的必要参数：

```go
type DeepSeekClient struct {
    APIKey  string
    BaseURL string
    Model   string
}
```

消息载体结构体：

```go
type ChatMessage struct {
    Role       string     `json:"role"`
    Content    string     `json:"content,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
    Name       string     `json:"name,omitempty"`
}

type ToolCall struct {
    ID       string       `json:"id"`
    Type     string       `json:"type"`
    Function ToolFunction `json:"function"`
}

type ToolFunction struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

标准的 DeepSeek Chat API 交互：

```go
type ChatRequest struct {
    Model    string           `json:"model"`
    Messages []ChatMessage    `json:"messages"`
    Tools    []map[string]any `json:"tools,omitempty"`
}

type ChatResponse struct {
    Choices []struct {
        Message ChatMessage `json:"message"`
    } `json:"choices"`
}

func (c *DeepSeekClient) Chat(
    ctx context.Context,
    messages []ChatMessage,
    tools []map[string]any,
) (*ChatResponse, error) {
    reqBody := ChatRequest{
        Model:    c.Model,
        Messages: messages,
        Tools:    tools,
    }
    data, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        c.BaseURL+"/chat/completions",
        bytes.NewBuffer(data),
    )
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("deepseek api error: %s", string(body))
    }
    var result ChatResponse
    err = json.Unmarshal(body, &result)
    if err != nil {
        return nil, err
    }
    return &result, nil
}
```

---

#### 第二步：抽象 Tool 层

考虑到后续会添加更多工具，这里引入接口抽象：

```go
type Tool interface {
    Name() string                                    // 工具名称
    Description() string                             // 工具描述
    InputSchema() map[string]any                     // 参数的 JSON Schema
    Call(ctx context.Context, input map[string]any) (string, error) // 执行调用
}
```

配套的工具注册表：

```go
type Registry struct {
    Tools map[string]Tool
}

func NewRegistry() *Registry {
    return &Registry{
        Tools: make(map[string]Tool),
    }
}

func (r *Registry) Register(tool Tool) error {
    name := tool.Name()
    if _, ok := r.Tools[name]; ok {
        return fmt.Errorf("%v tool has already exists", name)
    }
    r.Tools[name] = tool
    return nil
}

func (r *Registry) Get(name string) (Tool, bool) {
    tool, ok := r.Tools[name]
    return tool, ok
}

func (r *Registry) List() []Tool {
    list := make([]Tool, 0, len(r.Tools))
    for _, tool := range r.Tools {
        list = append(list, tool)
    }
    return list
}

// 将已注册工具转换为 DeepSeek API 要求的格式
func (r *Registry) ToDeepSeekFormat() []map[string]any {
    tools := make([]map[string]any, 0)
    for _, tool := range r.Tools {
        tools = append(tools, map[string]any{
            "type": "function",
            "function": map[string]any{
                "name":        tool.Name(),
                "description": tool.Description(),
                "parameters":  tool.InputSchema(),
            },
        })
    }
    return tools
}
```

---

#### 第三步：实现 Bash Tool

Bash 工具的 Input Schema 定义 —— 我们期望 LLM 返回如下 JSON：

```json
{
    "type": "object",
    "properties": {
        "command": {
            "type": "string",
            "description": "需要执行的 Bash 脚本"
        }
    },
    "required": ["command"]
}
```

代码实现：

```go
type BashTool struct{}

func (b *BashTool) Name() string {
    return "bash"
}

func (b *BashTool) Description() string {
    return "执行 Bash 命令"
}

func (b *BashTool) InputSchema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "command": map[string]any{
                "type":        "string",
                "description": "Bash command to execute",
            },
        },
        "required": []string{"command"},
    }
}

func (b *BashTool) Call(ctx context.Context, input map[string]any) (string, error) {
    command, ok := input["command"].(string)
    if !ok || command == "" {
        return "", fmt.Errorf("command is not exist or command is empty")
    }
    cmd := exec.CommandContext(ctx, "bash", "-c", command)
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    cmd.Stdout = &stdout
    err := cmd.Run()
    output := stdout.String()
    if stderr.Len() > 0 {
        if output != "" {
            output += "\n[STDERR]:\n" + stderr.String()
        } else {
            output = stderr.String()
        }
    }
    if output == "" && err == nil {
        output = "bash exec success"
    }
    return output, err
}
```

---

#### 第四步：构建 Agent 状态机

定义状态与事件：

```go
type StateName string

const (
    StartState    StateName = "start"
    ThinkingState StateName = "thinking"
    ToolState     StateName = "tool"
    FinishState             = "finish"
)

type Event string

const (
    EventInit             Event = "INIT"
    EventStartToThinking  Event = "START_TO_THINKING"
    EventThinkingToTool   Event = "THINKING_TO_TOOL"
    EventToolToThinking   Event = "TOOL_TO_THINKING"
    EventThinkingToFinish Event = "THINKING_TO_FINISH"
)
```

状态转移结构体：

```go
type Transition struct {
    From  StateName
    Event Event
    To    StateName
}
```

Agent 结构体与状态转移方法：

```go
type Agent struct {
    CurrentState StateName
    Transitions  []Transition
    Handlers     map[StateName]StateHandler
    Client       *DeepSeekClient
    Registry     *Registry
    Message      []ChatMessage
}

func (a *Agent) RegisterHandler(h StateHandler) {
    a.Handlers[h.Name()] = h
}

func (a *Agent) Transition(event Event) error {
    for _, t := range a.Transitions {
        if a.CurrentState == t.From && event == t.Event {
            a.CurrentState = t.To
            return nil
        }
    }
    return fmt.Errorf("invalid transition from %v event %v", a.CurrentState, event)
}
```

最后，Agent 的主循环 —— `Run` 方法：

```go
func (a *Agent) Run(ctx context.Context, userInput string) error {
    // 初始化：将用户输入加入消息列表
    a.Message = append(a.Message, ChatMessage{
        Role:    "user",
        Content: userInput,
    })
    // 进入 Thinking 状态
    err := a.Transition(EventStartToThinking)
    if err != nil {
        return err
    }
    // 事件循环
    for {
        handler, ok := a.Handlers[a.CurrentState]
        if !ok {
            return fmt.Errorf("handler not found: %s", a.CurrentState)
        }
        err := handler.Execute(ctx, a)
        if err != nil {
            if errors.Is(err, ErrAgentFinished) {
                return nil
            }
            return err
        }
    }
}
```

---

#### 小结

至此，一个最小可运行的 Agent 就搭建完成了。它的运转逻辑是：

1. 用户输入 → `Start` 状态
2. 进入 `Thinking` 状态 → 调用 DeepSeek API
3. LLM 返回工具调用 → 进入 `Tool` 状态 → 执行工具 → 回到 `Thinking` 状态
4. LLM 返回文本回答 → 进入 `Finish` 状态 → 结束

虽然目前只有一个 Bash 工具，但 `Tool` 接口和 `Registry` 已经为后续扩展做好了准备。下一课将为 Agent 添加更多工具，让它能做更复杂的事情。
