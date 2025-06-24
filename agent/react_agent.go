package agent

import (
	"context"
	"eino-mcp/thoughtchain"
	"eino-mcp/tool"
	"fmt"
	"github.com/cloudwego/eino/components/model"
	tools "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"strings"
)

// ConversationAgent 对话代理
type ConversationAgent struct {
	model    model.ChatModel
	tools    []tool.Tool
	messages []*schema.Message
	ctx      context.Context
	flow     compose.Runnable[map[string]any, *thoughtchain.ThoughtChain] // 新增：思考链流程
}

// NewConversationAgent 创建对话代理
func NewConversationAgent(
	ctx context.Context,
	model model.ChatModel,
	toolManager tool.MCPToolManager, // 依赖工具接口
	flow compose.Runnable[map[string]any, *thoughtchain.ThoughtChain], // 思考链流程
) (*ConversationAgent, error) {
	tools, err := toolManager.GetTools(ctx)
	if err != nil {

		return nil, err
	}

	// 绑定工具到模型
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		toolInfos = append(toolInfos, info)
	}
	if err := model.BindTools(toolInfos); err != nil {
		return nil, err
	}

	return &ConversationAgent{
		model:    model,
		tools:    tools,
		messages: []*schema.Message{},
		ctx:      ctx,
		flow:     flow, // 保存思考链流程
	}, nil
}

// SendMessage 处理用户消息
func (a *ConversationAgent) SendMessage(input string) (string, error) {
	userMsg := &schema.Message{Role: schema.User, Content: input}
	a.messages = append(a.messages, userMsg)
	var baseTools []tools.BaseTool
	for _, tool := range a.tools {
		baseTool, ok := tool.(tools.BaseTool)
		if !ok {
			panic("tool is not a BaseTool")
		}
		baseTools = append(baseTools, baseTool)
	}
	agent, err := react.NewAgent(a.ctx, &react.AgentConfig{
		Model: a.model,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: baseTools, // 直接使用工具接口
		},
		MaxStep: 20,
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			// 调用思考链流程
			thoughtChain, err := a.flow.Invoke(ctx, map[string]any{"query": input[len(input)-1].Content})
			if err != nil {
				return append([]*schema.Message{
					schema.SystemMessage("你是一个能调用工具的智能助手。用户询问具体功能（如门票、酒店）时，必须使用\"vector_search\"工具搜索功能入口URL，并直接返回链接"),
				}, input...)
			}
			stepsPrompt := fmt.Sprintf("任务步骤：\n%s", strings.Join(thoughtChain.Steps, "\n"))
			systemMsg := schema.SystemMessage(fmt.Sprintf(`你是一个能调用工具的智能助手。请根据以下要求处理用户问题：
%s
请按步骤选择工具并展示思考过程`, stepsPrompt))
			return append([]*schema.Message{systemMsg}, input...)
		},
	})
	if err != nil {
		return "", err
	}

	result, err := agent.Generate(a.ctx, a.messages)
	if err != nil {
		return "", err
	}

	a.messages = append(a.messages, &schema.Message{
		Role:    schema.Assistant,
		Content: result.Content,
	})
	return result.Content, nil
}
