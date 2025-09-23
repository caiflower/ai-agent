package chatmodel

import (
	"time"

	golocalv1 "github.com/caiflower/common-tools/pkg/golocal/v1"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/ollama/ollama/api"
)

func ollamaBuilder(config *Config) (model.ToolCallingChatModel, error) {
	keepalive := 60 * time.Second
	m, err := ollama.NewChatModel(golocalv1.GetContext(), &ollama.ChatModelConfig{
		// 基础配置
		BaseURL: config.BaseURL, // Ollama 服务地址
		Timeout: config.Timeout, // 请求超时时间

		// 模型配置
		Model: config.Model, // 模型名称
		//Format:    json.RawMessage(`"json"`), // 输出格式（可选）
		KeepAlive: &keepalive, // 保持连接时间

		// 模型参数
		Options: &api.Options{
			Runner: api.Runner{
				NumCtx: 4096, // 上下文窗口大小
				//NumGPU:    1,    // GPU 数量
				NumThread: 4, // CPU 线程数
			},
			Temperature:   0.7,        // 温度
			TopP:          0.9,        // Top-P 采样
			TopK:          40,         // Top-K 采样
			Seed:          42,         // 随机种子
			NumPredict:    10000,      // 最大生成长度
			Stop:          []string{}, // 停止词
			RepeatPenalty: 1.1,        // 重复惩罚
		},
	})

	return m, err
}
