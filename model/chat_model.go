package model

import (
	"context"
	"eino-mcp/config"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type ChatModel interface {
	model.BaseChatModel
	BindTools(tools []*schema.ToolInfo) error
}

type OpenAIChatModel struct {
	model.ChatModel // 内嵌原始模型实现
}

func NewOpenAIChatModel(ctx context.Context, cfg *config.OpenAIConfig) (ChatModel, error) {
	rawModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, err
	}
	return &OpenAIChatModel{
		ChatModel: rawModel,
	}, nil
}
