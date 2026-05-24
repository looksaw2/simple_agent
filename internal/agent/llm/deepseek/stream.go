package deepseek

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/looksaw/simple_agent_with_golang/internal/agent/llm"
)

// DeepSeek流式
type DeepSeekStream struct {
	body   io.ReadCloser
	reader *bufio.Reader
}

// 实现对应的关闭函数
func (s *DeepSeekStream) Close() error {
	return s.body.Close()
}

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
