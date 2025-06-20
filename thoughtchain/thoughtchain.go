package thoughtchain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"strings"
)

// ThoughtChain 思考链结构
type ThoughtChain struct {
	Steps  []string `json:"steps"`
	Answer string   `json:"answer"`
}

// BuildThoughtChainFlow 构建思考链流程（依赖模型接口）
func BuildThoughtChainFlow(ctx context.Context, model model.ChatModel) (compose.Runnable[map[string]any, *ThoughtChain], error) {
	promptTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(`你是一个逻辑清晰的分析助手。请按照以下要求处理用户问题：
1. 首先逐步思考解决问题的步骤；
2. 最后给出简洁的最终答案；
3. 输出格式必须为JSON，包含"steps"（思考步骤数组）和"answer"（最终答案）字段`),
		schema.UserMessage("用户问题：{query}"),
	)

	chain := compose.NewChain[map[string]any, *ThoughtChain]()
	chain.AppendChatTemplate(promptTemplate)
	chain.AppendChatModel(model)

	parseLambda := compose.InvokableLambda(
		func(ctx context.Context, msg *schema.Message) (*ThoughtChain, error) {
			var tc ThoughtChain
			if err := json.Unmarshal([]byte(msg.Content), &tc); err != nil {
				start := strings.Index(msg.Content, "{")
				end := strings.LastIndex(msg.Content, "}")
				if start >= 0 && end > start {
					jsonContent := msg.Content[start : end+1]
					if err := json.Unmarshal([]byte(jsonContent), &tc); err == nil {
						return &tc, nil
					}
				}
				return nil, fmt.Errorf("解析失败: %w\n原始内容: %s", err, msg.Content)
			}
			return &tc, nil
		},
	)

	chain.AppendLambda(parseLambda)
	return chain.Compile(ctx)
}
