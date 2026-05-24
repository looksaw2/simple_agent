package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
)

// 目前先弄个全局变量，但是之后马上会变成从yaml文件中读取
var DEEPSEEKBASEURL = "https://api.deepseek.com"

// 目前在lesson 3的时候是这样，之后会换
var MODEL = "deepseek-v4-pro"

// 具体的DeepSeek的Client定义
type Client struct {
	//Deepseek的API
	APiKey string
	//DeepSeek的BaseURL
	BaseURL string
	//使用的是Deepseek的什么模型
	Model string
	//通用的HTTP Client
	client *http.Client
}

var _ llm.StreamingModel = (*Client)(nil)

// 初始化Client(DeepSeek)
func NewClient(apiKey string, baseUrl string, model string, httpClient *http.Client) *Client {
	//检验httpClient是不是empty的
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		APiKey:  apiKey,
		BaseURL: baseUrl,
		Model:   model,
		client:  httpClient,
	}
}

// 重要的将llm的Request转换为DeepSeek的request
func (c *Client) Generate(ctx context.Context,
	req *llm.Request) (*llm.Response, error) {
	//将LLM 的Request转换成DeepSeek
	dsReq := toDeepSeekRequest(req, c.Model)
	data, err := json.Marshal(dsReq)
	if err != nil {
		return nil, err
	}
	//向DeepSeek的HTTP端点发请求
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+"/chat/completions",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}
	//然后把API KEY加上去
	httpReq.Header.Set(
		"Authorization",
		"Bear "+c.APiKey,
	)
	//然后是JSON的输出
	httpReq.Header.Set(
		"Content-Type",
		"application/json",
	)
	//然后开始发数据了
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	//如果发生400以上的错误，目前直接报错，后面还会有处理措施的
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("deepseek api error : %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var dsResp DeepSeekChatResponse
	err = json.Unmarshal(body, &dsResp)
	if err != nil {
		return nil, err
	}
	return fromDeepSeekResponse(&dsResp), nil
}

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
	data, err := json.Marshal(dsReq)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+"/chat/completions",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set(
		"Authorization",
		"Bearer "+c.APiKey,
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
		body:   resp.Body,
		reader: bufio.NewReader(resp.Body),
	}, nil

}
