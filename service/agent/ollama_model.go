package agent

import (
	"context"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/ollama/ollama/api"
)

func newOllamaModel(ctx context.Context) (model.BaseChatModel, error) {
	keepalive := 60 * time.Second
	m, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		// 基础配置
		BaseURL: "http://ollama-svc.ollama.svc.cluster.local:80", // Ollama 服务地址
		Timeout: 30 * time.Second,                                // 请求超时时间

		// 模型配置
		Model: "Qwen3-0.6B:latest", // 模型名称
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
			NumPredict:    100,        // 最大生成长度
			Stop:          []string{}, // 停止词
			RepeatPenalty: 1.1,        // 重复惩罚
		},
	})

	return m, err
}
