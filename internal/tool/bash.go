package tool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type BashTool struct{}

func (b *BashTool) Name() string {
	return "bash"
}

func (b *BashTool) Description() string {
	return "这是Bash的工具,执行和Bash有关的操作"
}

func (b *BashTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "Bash command to execute",
			},
		},
		"required": []string{"command"},
	}
}

func (b *BashTool) Call(ctx context.Context, input map[string]any) (string, error) {
	command, ok := input["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command is not exist or command is empty")
	}
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
