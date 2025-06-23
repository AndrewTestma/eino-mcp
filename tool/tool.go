package tool

import (
	"context"
	mccp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"log"
	"net/http"
)

type Tool interface {
	tool.BaseTool
	Info(ctx context.Context) (*schema.ToolInfo, error)
}

type MCPToolManager struct {
	cli *client.Client
}

func NewMCPToolManager(ctx context.Context, serverURL string) (*MCPToolManager, error) {
	cli, err := client.NewSSEMCPClient(serverURL, client.WithHTTPClient(&http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}))
	if err != nil {
		return nil, err
	}
	if err := cli.Start(ctx); err != nil {
		cli.Close()
		return nil, err
	}
	initReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo:      mcp.Implementation{Name: "eino-mcp-client"},
		},
	}

	if _, err := cli.Initialize(ctx, initReq); err != nil {
		log.Fatalf("❌ MCP 协议初始化失败: %v", err)
	}
	return &MCPToolManager{cli: cli}, nil
}

func (m *MCPToolManager) GetTools(ctx context.Context) ([]Tool, error) {
	rawTools, err := mccp.GetTools(ctx, &mccp.Config{Cli: m.cli})
	if err != nil {
		return nil, err
	}
	tools := make([]Tool, 0, len(rawTools))
	for _, t := range rawTools {
		tools = append(tools, t) // 假设rawTools实现了Tool接口
	}
	return tools, nil
}

// Close 关闭MCP客户端
func (m *MCPToolManager) Close() {
	m.cli.Close()
}
