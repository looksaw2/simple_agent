package llm


//这些信息是谁发出的
type Role string

//一下包含4种角色
const(
	//System最高角色
	SystemRole Role = "system"
	//用户角色
	UserRole Role = "user"
	//助手的角色
	AssitantRole Role = "assitant"
	//Tool的角色
	TooRole Role = "tool"
)

//然后对应的消息的抽象
type Message struct {
	Role Role
	Content string
	ToolCalls []ToolCall 
}
//使用工具调用
type ToolCall struct {

}