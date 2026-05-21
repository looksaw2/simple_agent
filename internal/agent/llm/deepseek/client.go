package deepseek

import (
	"context"
	"net/http"
	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
)

//目前先弄个全局变量，但是之后马上会变成从yaml文件中读取
var DEEPSEEKBASEURL = "https://api.deepseek.com"
//目前在lesson 3的时候是这样，之后会换
var MODEL = "deepseek-v4-pro"
// 具体的DeepSeek的Client定义
type Client struct{
	//Deepseek的API
	APiKey string
	//DeepSeek的BaseURL
	BaseURL string
	//使用的是Deepseek的什么模型
	Model string
	//通用的HTTP Client
	client *http.Client
}
//初始化Client(DeepSeek)
func NewClient(apiKey string , baseUrl string , model string ,httpClient *http.Client) *Client {
	//检验httpClient是不是empty的
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		APiKey: apiKey,
		BaseURL: baseUrl,
		Model: model,
		client: httpClient,
	}
}

func (c *Client) Generate(ctx context.Context,
	req *llm.Request) (*llm.Response, error) {
	return nil, nil
}

func (c *Client) Stream(ctx context.Context,
	req *llm.Request,
) (<-chan *llm.Chunk, error) {
	return nil, nil
}
