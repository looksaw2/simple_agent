package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/looksaw/simple_agent_with_golang/internal/agent/builders"
	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm/deepseek"
	agentstate "github.com/looksaw/simple_agent_with_golang/internal/agent/state"
	"github.com/looksaw/simple_agent_with_golang/internal/core/runtime"
	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
	"github.com/looksaw/simple_agent_with_golang/internal/tool"
	"github.com/looksaw/simple_agent_with_golang/internal/tool/builtin"
)

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
		log.Fatalf("failed to register bash tool: %v", err)
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
		log.Fatalf("create graph failed: %v", err)
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
