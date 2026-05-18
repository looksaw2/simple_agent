package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

/*
 *Lesson 1
 */
// 类型的名称
type StateName string

var ErrAgentFinished = errors.New("agent is finished ....")

// 类型的定义
const (
	StartState    StateName = "start"
	ThinkingState StateName = "thinking"
	ToolState     StateName = "tool"
	FinishState             = "finish"
)

// 然后开始定义事件之间的转移
type Event string

// 然后定义其中的事件
const (
	EventInit             Event = "INIT"
	EventStartToThinking  Event = "START_TO_THINKING"
	EventThinkingToTool   Event = "THINKING_TO_TOOL"
	EventToolToThinking   Event = "TOOL_TO_THINKING"
	EventThinkingToFinish Event = "THINKING_TO_FINISH"
)

// 然后定义转移
type Transition struct {
	From  StateName
	Event Event
	To    StateName
}

// 然后开始定义结构体
type State struct {
	StateName StateName
	Message   string
}

// 然后开始初始化State
func NewState(stateName StateName, message string) State {
	return State{
		StateName: stateName,
		Message:   message,
	}
}

// 然后初始化Agent
type Agent struct {
	CurrentState StateName
	Transitions  []Transition
	Handlers     map[StateName]StateHandler
	Client       *DeepSeekClient
	Registry     *Registry
	Message      []ChatMessage
}

func (a *Agent) RegisterHandler(
	h StateHandler,
) {
	a.Handlers[h.Name()] = h
}

// 然后我们实现状态转移
func (a *Agent) Transition(event Event) error {
	for _, t := range a.Transitions {
		if a.CurrentState == t.From && event == t.Event {
			a.CurrentState = t.To
			return nil
		}
	}
	return fmt.Errorf("invalid transition from %v event %v", a.CurrentState, event)
}

// 然后我们初始化状态机
func NewAgent(
	client *DeepSeekClient,
	registry *Registry,
) *Agent {
	a := &Agent{
		CurrentState: StartState,
		Client:       client,
		Registry:     registry,
		Handlers:     make(map[StateName]StateHandler),
		Transitions: []Transition{
			{
				From:  StartState,
				Event: EventStartToThinking,
				To:    ThinkingState,
			},
			{
				From:  ThinkingState,
				Event: EventThinkingToTool,
				To:    ToolState,
			},
			{
				From:  ToolState,
				Event: EventToolToThinking,
				To:    ThinkingState,
			},
			{
				From:  ThinkingState,
				Event: EventThinkingToFinish,
				To:    FinishState,
			},
		},
	}
	a.Registry.Register(&BashTool{})
	a.RegisterHandler(
		&ThinkingStateHandler{},
	)

	a.RegisterHandler(
		&ToolStateHandler{},
	)

	a.RegisterHandler(
		&FinishStateHandler{},
	)
	return a
}

type StateHandler interface {
	Name() StateName
	Execute(ctx context.Context, a *Agent) error
}

// Thinking的Handler
type ThinkingStateHandler struct{}

func (h *ThinkingStateHandler) Name() StateName {
	return ThinkingState
}
func (h *ThinkingStateHandler) Execute(ctx context.Context, a *Agent) error {
	resp, err := a.Client.Chat(
		ctx,
		a.Message,
		a.Registry.ToDeepSeekFormat(),
	)
	if err != nil {
		return err
	}
	msg := resp.Choices[0].Message
	a.Message = append(a.Message, msg)
	if len(msg.ToolCalls) > 0 {
		return a.Transition(EventThinkingToTool)
	}
	println(msg.Content)
	return a.Transition(EventThinkingToFinish)
}

// Tool的Handler
type ToolStateHandler struct{}

func (h *ToolStateHandler) Name() StateName {
	return ToolState
}
func (h *ToolStateHandler) Execute(ctx context.Context, a *Agent) error {
	last := a.Message[len(a.Message)-1]
	for _, tc := range last.ToolCalls {
		tool, ok := a.Registry.Tools[tc.Function.Name]
		if !ok {
			return fmt.Errorf("tool %s not found .......", tc.Function.Name)
		}
		var args map[string]any
		err := json.Unmarshal([]byte(tc.Function.Arguments), &args)
		if err != nil {
			return err
		}
		result, err := tool.Call(ctx, args)
		if err != nil {
			result = err.Error()
		}
		a.Message = append(a.Message, ChatMessage{
			Role:       "tool",
			Name:       tc.Function.Name,
			ToolCallID: tc.ID,
			Content:    result,
		})
	}
	return a.Transition(EventToolToThinking)
}

type FinishStateHandler struct{}

func (h *FinishStateHandler) Name() StateName {
	return FinishState
}

func (h *FinishStateHandler) Execute(
	ctx context.Context,
	a *Agent,
) error {
	return ErrAgentFinished
}

// 开始运行
func (a *Agent) Run(ctx context.Context, userInput string) error {
	//初始化
	a.Message = append(a.Message, ChatMessage{
		Role:    "user",
		Content: userInput,
	})
	//完成初始化，进入Thinking状态
	err := a.Transition(EventStartToThinking)
	if err != nil {
		return err
	}
	//然后开始循环
	for {
		handler, ok := a.Handlers[a.CurrentState]
		if !ok {
			return fmt.Errorf(
				"handler not found: %s",
				a.CurrentState,
			)
		}
		err := handler.Execute(ctx, a)
		if err != nil {
			if errors.Is(
				err,
				ErrAgentFinished,
			) {
				return nil
			}
			return err
		}
	}
}

func main() {
	apikey := "xxxxxxxxxx"
	client := NewDeepSeekClient(apikey)
	registry := NewRegistry()
	agent := NewAgent(client, registry)
	userInput := "DeepSeek你好，你是一个Agent助手，请帮我完成一些任务,查看当前目录"
	err := agent.Run(context.Background(), userInput)
	if err != ErrAgentFinished && err != nil {
		fmt.Printf("err is %v\n", err)
	}
	fmt.Println("Finished ...........")

}
