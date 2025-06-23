package config

import "github.com/spf13/viper"

type MCPConfig struct {
	ServerURL string `mapstructure:"server_url"`
}

type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
	Proxy   string `mapstructure:"proxy"` // 新增代理配置
}

type AppConfig struct {
	MCP    MCPConfig    `mapstructure:"mcp"`
	OpenAI OpenAIConfig `mapstructure:"openai"`
}

func LoadConfig() (*AppConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
