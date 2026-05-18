package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

/*
* 这个是关于Bash的
 */

type BashTool struct{}

func (b *BashTool) Name() string {
	return "bash"
}

func (b *BashTool) Description() string {
	return "这是Bash的工具,执行和Bash有关的操作"
}

/*
* 首先对于Bash而言我的输入实际上是一个Object，然后这个Object里面有command字段，这个字段可以是需要执行的命令，
* 并且这个字段是必须的因此，我可以写下
* {
* 	"type" : "object",
*   "properties" : {
*		"command" : {
*			"type" : "string",
*           "description" : "这里请填写需要执行的Bash脚本"
*		}
*	},
*   "required" : ["command"]
* }
 */
func (b *BashTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "Bash command to execute",
				"example":     []string{"ls", "cat", "tail", "pwd"},
			},
		},
		"required": []string{"command"},
	}
}

/*
* DeepSeek大概会返回
* {
*  		"command" : "echo 'Hello World'"
* }
 */
func (b *BashTool) Call(ctx context.Context, input map[string]any) (string, error) {
	//开始解析对应的命令
	command, ok := input["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command is not exist or command is empty %v....", ok)
	}
	//开始命令的执行
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
