# Simple Agent with Go

从零开始，用 Go 构建一个最小化的 AI Agent，逐步理解 Agent 的底层原理。

## 项目概述

这是一个教学项目，通过有限状态机（Finite State Machine）驱动 Agent 的行为，对接 DeepSeek API 实现 LLM 调用和工具调用（Tool Calling）。

核心组件：

- **状态机**：`Start → Thinking → Tool/Finish → Finish`，清晰的状态转移模型
- **DeepSeek Client**：封装对 DeepSeek Chat API 的 HTTP 调用
- **Tool Registry**：可扩展的工具注册与调用框架
- **Bash Tool**：内置首个工具，支持执行 shell 命令

## 项目结构

```
.
├── cmd/
│   ├── main.go      # Agent 状态机、Handler、主入口
│   ├── client.go    # DeepSeek API 客户端
│   ├── tool.go      # Tool 接口定义
│   ├── registry.go  # 工具注册表
│   └── bash.go      # Bash 工具实现
├── docs/
│   └── lesson1.md   # Lesson 1：从零构建最小 Agent
├── img/
│   └── lesson1_state_transform.svg  # 状态转移图
└── go.mod
```

## 快速开始

```bash
# 设置 API Key
export DEEPSEEK_API_KEY="your-api-key"

# 运行
go run ./cmd/...
```

## 学习路线

- **Lesson 1**：搭建最小可运行 Agent —— 状态机、工具抽象、Bash 工具
- 更多 Lesson 持续施工中...

## License

MIT
