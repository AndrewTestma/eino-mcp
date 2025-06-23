package main

import (
	"context"
	"eino-mcp/agent"
	"eino-mcp/config"
	"eino-mcp/model"
	"eino-mcp/thoughtchain"
	"eino-mcp/tool"
	"fmt"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化上下文
	ctx := context.Background()

	// 初始化模型
	openaiModel, err := model.NewOpenAIChatModel(ctx, &config.OpenAIConfig{
		APIKey:  cfg.OpenAI.APIKey,
		BaseURL: cfg.OpenAI.BaseURL,
		Model:   cfg.OpenAI.Model,
	})
	if err != nil {
		panic(fmt.Sprintf("初始化模型失败: %v", err))
	}

	// 初始化MCP工具管理器
	toolManager, err := tool.NewMCPToolManager(ctx, cfg.MCP.ServerURL)
	if err != nil {
		panic(fmt.Sprintf("初始化工具管理器失败: %v", err))
	}
	defer toolManager.Close()

	// 初始化思考链流程
	flow, err := thoughtchain.BuildThoughtChainFlow(ctx, openaiModel)
	if err != nil {
		panic(fmt.Sprintf("构建思考链流程失败: %v", err))
	}

	// 初始化对话代理
	convAgent, err := agent.NewConversationAgent(ctx, openaiModel, *toolManager, flow)
	if err != nil {
		panic(fmt.Sprintf("初始化对话代理失败: %v", err))
	}

	// 示例：使用对话代理
	fmt.Println("智能助手已启动，输入'退出'结束对话")
	for {
		fmt.Print("用户: ")
		var input string
		fmt.Scanln(&input)

		if input == "退出" {
			break
		}

		// 使用对话代理处理消息（支持工具调用）
		response, err := convAgent.SendMessage(input)
		if err != nil {
			fmt.Printf("对话错误: %v\n", err)
			continue
		}
		fmt.Printf("助手: %s\n", response)

		// 示例：使用思考链分析
		thoughtChain, err := flow.Invoke(ctx, map[string]any{"query": input})
		if err != nil {
			fmt.Printf("思考链错误: %v\n", err)
			continue
		}
		fmt.Println("\n=== 思考链分析 ===")
		for i, step := range thoughtChain.Steps {
			fmt.Printf("%d. %s\n", i+1, step)
		}
		fmt.Printf("最终结论: %s\n\n", thoughtChain.Answer)
	}
}
