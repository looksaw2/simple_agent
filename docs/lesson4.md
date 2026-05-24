### lesson 4

在lesson 3上面我忘记将Generate的逻辑写了，下面来跟着SSE一起讲吧，从一般的意义上来说，理想的情况是本地的Agent发送HTTP报文给LLM然后LLM将信息发回来。看起来一切都很美好

![](lesson4_send_http)
但是和可惜的是LLM很多时候都是一个一个字蹦出来的，这是为什么，这又要回到经典的transformer架构了，首先回想一下我们是怎么做transformer的，首先我们准备一段需要input的数据，然后做一个embedding然后打个不严谨仅供参考比方是例如输入['Hello']就会变成['H','e','l','l','o']然后变成类似[5,4096]之类的结构，然后或早或晚的做一个位置编码，然后和 $W_{Q}$ , $W_{K}$ 矩阵相乘，然后再和 $QK^{T}$ 之后加上掩码，最后在和 $V$ 相乘得到结果，但是在训练的时候从 $QK^{T}$ 可以拿到全部的信息，但是在推理的时候可是没有这种，后一个的结果依赖前一个的输出 $ x_{i+1} = f(x_{i}) $ 导致不严谨的说大模型没法输出Hello，他只能H -> He -> Hel -> Hell -> Hello 当然实际情况不是这样，但是对于上层工程师我们可以简单的认为就是这样，这样就带来了一个问题，用户的体验需要怎么优化？假如我们还采取上面之前的方式的话，就会发生如下的现象:
![](lesson4_without_stream)

显然这样对于用户的体验来说是非常糟糕的，用户肯定会[黑人问号]，并且如果是工具调用的话，你必须等所有的工具返回之后才可以开始调用，这个是不好的，用户肯定想的是，为什么不能第一个工具输出结束之后就马上开始执行？这的确，所以SSE应运而生，通过一个流式传输试图解决上面的问题,在之前的代码里面我们试图通过channel来传输，但是这样的方式语义较弱，所以鉴于上述的思考，我们得出了如下的代码
```go
// 模型的公共定义
type Model interface {
	//发送消息
	Generate(ctx context.Context, req *Request) (*Response, error)
}
// Chunk的定义
type Chunk struct {
	ContentDelta  string
	ToolCallDelat *ToolCall
	FinishReason string
}
//Stream的定义
type Stream interface {
	Recv() (*Chunk,error)
	Close() error
}
type StreamingModel interface {
	Model
	Stream(ctx context.Context,req *Request)(Stream,error)
}
```
好了，那SSE是什么？SSE的全称中文名是服务器推送事件，是一种服务器向客户端不断的推送的一种方式，就传统的HTTP而言，在很早之前HTTP是一个短链接的模式，他具体的模式如下
![](lesson4_http_early)
但是伴随这样的设计模式一次HTTP报文的交换就需要一次TCP的握手和挥手，这样在某些场景造成极大的浪费，所以HTTP的长连接应运而生，通过这样的一个头标记
```
Connection: keep-alive
```
来保存在TCP握手之后客户端可以和服务器在不关闭TCP的连接下面不断的发生交互，但即便是这样HTTP还是一个半双工的协议，对于LLM而言很多的时候是H -> e -> l -> l -> o -> EOF 这样的发送信息，因此这样的SSE就比较符合这样的协议，对于SSE来说是是在HTTP的头里面有
```
Content-Type: text/event-stream
```
然后来发送对应的Stream流并且由于服务器直接关闭 TCP 连接会导致客户端自动重连，这往往不是期望的行为。因此，跨语言的最佳实践是：先发送应用层信号，让客户端主动关闭，所以说关闭的信号是我们自己在应用层自定义的，并且由于这样SSE短短续续的特性，因此对应的设计模式也会转向对应的事件驱动的设计下面我们来对其进行具体的修改
```go
type DeepSeekChatRequest struct {
	Model string `json:"model"`
	Messages []DeepSeekChatMessage `json:"messages"`
	Tools []map[string]any `json:"tools,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	//是否使用Stream流式传输
	Stream bool `json:"stream,omitempty"`
}
```
添加了对应的流式传输，然后研究Deepseek的流式传输的返回的结果可以得到如下的返回的结构体
```
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion.chunk",
  "created": 1703123456,
  "model": "deepseek-chat",
  "choices": [
    {
      "index": 0,
      "delta": {
        "role": "assistant",      // 仅在第一个块中出现
        "content": "你"            // 增量内容，每次只吐几个字符
      },
      "finish_reason": null       // 中间块为 null，最后一块为 "stop"
    }
  ]
}
```

```go
//Deepseek的流式传输
type DeepSeekStreamResponse struct {
	ID string `json:"id"`
	Object string `json:"object"`
	Model string `json:"model"`

}
//DeepSeek的流式Choice
type DeepSeekStreamChoice struct {
	Index int `json:"index"`
	Delta DeepSeekDelta `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

//DeepSeek的Delta
type DeepSeekDelta struct {
	Role  string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}
```
然后很重要的一点是流式的增量不仅仅是text文本的增量相反的是struct的增量，是结构的增量，这就给我们带来了巨量的好处，即对于ToolCall的假设有3个ToolCall我们并不需要等待全部的ToolCall到来完，反而是可以使用对应的ToolCall增量来说明对应的对应的Tool已经可以开始执行了，减少ToolCall的等待时间，并且就这样缩短用户的等待。这样我们就可以实现下面的代码

```go
// Stream接口的实现
func (c *Client) Stream(ctx context.Context,
	req *llm.Request,
) (llm.Stream, error) {
	//发出DeepSeek的消息
	dsReq := toDeepSeekRequest(
		req,
		c.Model,
	)
	//将流式标志位置为True
	dsReq.Stream = true
	//然后序列化开始发请求
	data , err := json.Marshal(dsReq)
	if err != nil {
		return nil ,err
	}
	httpReq , err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL + "/chat/completions",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil ,err
	}
	httpReq.Header.Set(
		"Authorization",
		"Bearer " + c.APiKey,
	)
	httpReq.Header.Set(
		"Content-Type",
		"application/json",
	)
	httpReq.Header.Set(
		"Accept",
		"text/event-stream",
	)
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf(
			"deepseek stream error: %s",
			string(body),
		)
	}
	return &DeepSeekStream{
		body: resp.Body,
		reader: bufio.NewReader(resp.Body),
	}, nil
}
```
然后返回一个stream的事件流之后，应该是不断的得到增量然后根据增量Delta的完整程度来判断对应的需要采取的动作，并且此外，就接口的实现的需要来说，接口需要实现Recv方法，因此，让我们开始实现这个Recv方法吧

```go
// 实现对应的接受函数
func (s *DeepSeekStream) Recv() (*llm.Chunk, error) {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		//是空的消息头
		if line == "" {
			continue
		}
		//如果来的数据没有data这个字段
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		//剩下的便是有Data字段的
		payload := strings.TrimPrefix(
			line,
			"data: ",
		)
		//是不是收到了DONE
		if payload == "[Done]" {
			return nil, io.EOF
		}
		//下面开始解析
		var resp DeepSeekStreamResponse
		err = json.Unmarshal(
			[]byte(payload),
			&resp,
		)
		if err != nil {
			return nil, err
		}
		if len(resp.Choices) == 0 {
			continue
		}
		choice := resp.Choices[0]
		chunk := &llm.Chunk{
			RoleDelta:    choice.Delta.Role,
			ContentDelta: choice.Delta.Content,
			FinishReason: choice.FinishReason,
		}
		if len(choice.Delta.ToolCalls) > 0 {
			tc := choice.Delta.ToolCalls[0]
			chunk.ToolCallDelta = &llm.ToolCallDelta{
				ID: tc.ID,
				NameDelta: tc.Function.Name,
				ArgumentsDelta: tc.Function.Arguments,
			}
		}
		return chunk, nil
	}
}
```
然后不少人肯定会有疑问，即来的chunk不完整怎么办？这个时候就体现出我们抽象的作用了,在目前这一层，我们只负责组装，至于具体的拼接的逻辑交到下一层。这些具体的实现放后面的lesson吧。