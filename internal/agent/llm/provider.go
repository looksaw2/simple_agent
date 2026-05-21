package llm

//定义有哪些provider

type Provider string

// 先定义一些简单的Provider吧
const (
	ProviderDeepSeek Provider = "deepseek"
	ProviderOpenai   Provider = "openai"
	ProviderClaude   Provider = "claude"
	ProviderGemini   Provider = "gemini"
	ProviderKimi     Provider = "kimi"
	ProviderGLM      Provider = "GLM"
	ProviderQWEN     Provider = "qwen"
)
