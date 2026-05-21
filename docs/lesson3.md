### lesson 3
开始进入更上层的领域，目前在lesson2中我们已经完成了对于Graph化的处理，但是在lesson2中我们只是仅仅的搭建了底层的结构，接下来我们需要搭建更高层的结构了，例如之前我们仅仅只是对Deepseek有一层Client的结构体但是这远远不够，例如以后还要QWEN，OPENAI，GEMINI，CLAUDE之类的模型呢？这里显然需要完成一个解耦，即具体的核心代码和厂商的无关，需要一个接口适配来满足对于不同厂商的不同的格式，并且在之前我们并没有完成大模型的流式传输但是对于更好的用户体验而言这个是必不可少的。其他的诸如memory也需要做一个抽象。那就让我们开始构建把。

首先，需要对于这些Deepseek，Qwen，Chatgpt之类的LLM提取共性，那首先提出Model.go和Message.go是模型需要满足什么功能和消息应该是什么格式，对于模型而言其基本应该满足基本的HTTP和流式传输，所以在Model.go中我们可以抽象出具体的模型应该满足什么样子的抽象，并且之前的Client也需要做对应的抽象，所以我们首要的目的是抽象出这个两个的接口的
```
internal/llm/message.go
internal/llm/model.go
```
首先，我们定义好了消息，其具体的结构如下:
```go
type Message struct {
	Role Role
	Content string
	ToolCalls []ToolCall 
}
```

那首先我们应该考虑角色，是什么样子的角色比较好，首先我们初略的就可以想到一个系统角色，一个用户角色，一个AI回复的角色·，还有由于我们采用了工具调用产生的工具角色，所以可以分成如下的角色:

![](lesson3_role.excalidraw)

此外Content顾名思义是每次传入的内容但是ToolCall怎么定义，通常来想一想我们是怎么定义和使用函数的，例如在Python里面，我们是这样定义和使用函数的
```python
def sum(a : float , b : float) ->float:
	return a + b
sum(1,2)
```
那由于LLM是一个语言(多模态)的模型，我们想要告诉大模型怎么去调用这个函数，就必须告诉其name,arguments。所以对应ToolCall而言，我们可以如下的定义
```go
type ToolCall struct {
	ID string
	Name string
	Arguments map[string]any
}
```
并且我们和大模型的交互是通过其和HTTP交互的，并且为了用户的体验我们需要使用SSE技术来防止用户等待，不过在lesson 3的时候先让我们简单一点，使用对应一个简单的定义来描述Http的Request和Response那么golang的定义如下:
```go
//接受LLM的Response的定义
type Response struct {
	Messages []Message
	FinishedReason string
}
//发送LLM的Request的定义
type Request struct {
	Messages []Message
	Tools []map[string]any
	Temperature float64
}
```
然后我们思考模型层的普遍定义，例如无论是Chatgpt，Deepseek，Claude Code，Gemini他们共同的能力是什么？对的，收发HTTP，对于这个样子来说，我们就可以来抽象一些模型共有的能力，便可以得到下面的一些共性的定义
```go

// 模型的公共定义
type Model interface {
	//发送消息
	Generate(ctx context.Context, req *Request) (*Response, error)
	//发送流式消息
	Stream(ctx context.Context, req *Request)(<- chan Chunk,error)
}
//Chunk的定义
type Chunk struct {
	ContentDelta string
	ToolCallDelat *ToolCall
}
```
然后我们完成了对应的抽象了，下面就需要实现对应的层级结构，即Deepseek,OpenAI都需要实现这个接口，然后我们在 `internal/llm/deepseek`新创建这个文件夹，并且如果以后需要继续使用添加的话，继续在下面添加文件就可以了(以后是Provider但目前让我们简单一点一步一步的迭代)，由于之前client是Deepseek的client，那接下来我们实现的也是Deepseek的接口，完成了简单的接口之后，我们首先取参考Deepseek的官网的golang和curl的方法示例，如下:
```go
package main

import (
  "fmt"
  "strings"
  "net/http"
  "io/ioutil"
)

func main() {

  url := "https://api.deepseek.com/chat/completions"
  method := "POST"

  payload := strings.NewReader(`{
  "messages": [
    {
      "content": "You are a helpful assistant",
      "role": "system"
    },
    {
      "content": "Hi",
      "role": "user"
    }
  ],
  "model": "deepseek-v4-pro",
  "thinking": {
    "type": "enabled"
  },
  "reasoning_effort": "high",
  "max_tokens": 4096,
  "response_format": {
    "type": "text"
  },
  "stop": null,
  "stream": false,
  "stream_options": null,
  "temperature": 1,
  "top_p": 1,
  "tools": null,
  "tool_choice": "none",
  "logprobs": false,
  "top_logprobs": null
}`)

  client := &http.Client {
  }
  req, err := http.NewRequest(method, url, payload)

  if err != nil {
    fmt.Println(err)
    return
  }
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Authorization", "Bearer <TOKEN>")

  res, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
    return
  }
  defer res.Body.Close()

  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println(string(body))
}
```
和curl相关的示例
```shell
curl -L -X POST 'https://api.deepseek.com/chat/completions' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
--data-raw '{
  "messages": [
    {
      "content": "You are a helpful assistant",
      "role": "system"
    },
    {
      "content": "Hi",
      "role": "user"
    }
  ],
  "model": "deepseek-v4-pro",
  "thinking": {
    "type": "enabled"
  },
  "reasoning_effort": "high",
  "max_tokens": 4096,
  "response_format": {
    "type": "text"
  },
  "stop": null,
  "stream": false,
  "stream_options": null,
  "temperature": 1,
  "top_p": 1,
  "tools": null,
  "tool_choice": "none",
  "logprobs": false,
  "top_logprobs": null
}'
```
所以可以抽象出来如下的deepseek相关的go程序,将DeepSeek的专用的Request和Response转化为我的LLM内部的Request和Response，做一个Adapter，这样我内部的LLM就不会需要担心和外部接入耦合的问题.
```go
//将Deepseek的模式转化为标准的格式
type DeepSeekChatRequest struct {
	Model string `json:"model"`
	Messages []DeepSeekChatMessage `json:"messages"`
	Tools []map[string]any `json:"tools,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}
//DeepSeek的Response
type DeepSeekChatResponse struct {
	Choices []struct {
		Message DeepSeekChatMessage `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}
//DeepSeek的Message定义
type DeepSeekChatMessage struct {
	Role string `json:"role"`
	Content string `json:"content,omitempty"`
	ToolCalls []DeepSeekToolCall `json:"tool_calls,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
	Name string `json:"name,omitempty"`
}
type DeepSeekToolCall struct {
	ID string `json:"id"`
	Type string `json:"type"`
	Function DeepSeekFunction `json:"function"`
}
type DeepSeekFunction struct {
	Name string `json:"name"`
	Arguments string `json:"arguments"`
}
//将标准的Request转化为DeepSeek的Request
func toDeepSeekRequest(
	req *llm.Request,
	defaultModel string,
) *DeepSeekChatRequest {
	model := req.Model
	if model == "" {
		model = defaultModel
	}
	msgs := make([]DeepSeekChatMessage,0,len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, toDeepSeekMessage(m))
	}
	return &DeepSeekChatRequest{
		Model: model,
		Messages: msgs,
		Tools: req.Tools,
		Temperature: req.Temperature,
	}
}
func toDeepSeekMessage(
	m llm.Message,
) DeepSeekChatMessage {
	toolCalls := make(
		[]DeepSeekToolCall,
		0,
		len(m.ToolCalls),
	)
	for _, tc := range m.ToolCalls {
		toolCalls = append(
			toolCalls,
			toDeepSeekToolCall(tc),
		)
	}
	return DeepSeekChatMessage{
		Role: string(m.Role),
		Content: m.Content,
		ToolCalls: toolCalls,
		ToolCallID: m.ToolCallID,
		Name: m.Name,
	}
}
func toDeepSeekToolCall(
	tc llm.ToolCall,
) DeepSeekToolCall {
	return DeepSeekToolCall{
		ID: tc.ID,
		Type: tc.Type,
		Function: DeepSeekFunction{
			Name: tc.Function.Name,
			Arguments: tc.Function.Arguments,
		},
	}
}
```
然后完成了具体的厂商和LLM的解耦