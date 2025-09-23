package main

import (
	"context"
	"time"

	golocalv1 "github.com/caiflower/common-tools/pkg/golocal/v1"
	"github.com/caiflower/common-tools/pkg/logger"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/ollama/ollama/api"
)

func main() {

	ctx := context.Background()

	callbacks.AppendGlobalHandlers(&loggerCallbacks{})

	// 1. create an instance of ChatTemplate as 1st Graph Node
	systemTpl := `你是一名房产经纪人，结合用户的薪酬和工作，使用 user_info API，为其提供相关的房产信息。邮箱是必须的`
	chatTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemTpl),
		schema.MessagesPlaceholder("message_histories", true),
		schema.UserMessage("{user_query}"),
	)

	keepalive := 60 * time.Second
	chatModel, err := ollama.NewChatModel(golocalv1.GetContext(), &ollama.ChatModelConfig{
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

	// 3. create an instance of tool.InvokableTool for Intent recognition and execution
	userInfoTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "user_info",
			Desc: "根据用户的姓名和邮箱，查询用户的公司、职位、薪酬信息",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"name": {
					Type: "string",
					Desc: "用户的姓名",
				},
				"email": {
					Type: "string",
					Desc: "用户的邮箱",
				},
			}),
		},
		func(ctx context.Context, input *userInfoRequest) (output *userInfoResponse, err error) {
			return &userInfoResponse{
				Name:     input.Name,
				Email:    input.Email,
				Company:  "Bytedance",
				Position: "CEO",
				Salary:   "9999",
			}, nil
		})

	info, err := userInfoTool.Info(ctx)
	if err != nil {
		logger.Error("Get ToolInfo failed, err=%v", err)
		return
	}

	// 4. bind ToolInfo to ChatModel. ToolInfo will remain in effect until the next BindTools.
	toolModels, err := chatModel.WithTools([]*schema.ToolInfo{info})
	if err != nil {
		logger.Error("BindForcedTools failed, err=%v", err)
		return
	}

	// 5. create an instance of ToolsNode as 3rd Graph Node
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{userInfoTool},
	})
	if err != nil {
		return
	}

	const (
		nodeKeyOfTemplate  = "template"
		nodeKeyOfChatModel = "chat_model"
		nodeKeyOfTools     = "tools"
	)

	// 6. create an instance of Graph
	// input type is 1st Graph Node's input type, that is ChatTemplate's input type: map[string]any
	// output type is last Graph Node's output type, that is ToolsNode's output type: []*schema.Message
	g := compose.NewGraph[map[string]any, []*schema.Message]()

	// 7. add ChatTemplate into graph
	_ = g.AddChatTemplateNode(nodeKeyOfTemplate, chatTpl)

	// 8. add ChatModel into graph
	_ = g.AddChatModelNode(nodeKeyOfChatModel, toolModels)

	// 9. add ToolsNode into graph
	_ = g.AddToolsNode(nodeKeyOfTools, toolsNode)

	// 10. add connection between nodes
	_ = g.AddEdge(compose.START, nodeKeyOfTemplate)

	_ = g.AddEdge(nodeKeyOfTemplate, nodeKeyOfChatModel)

	_ = g.AddEdge(nodeKeyOfChatModel, nodeKeyOfTools)

	_ = g.AddEdge(nodeKeyOfTools, compose.END)

	// 9. compile Graph[I, O] to Runnable[I, O]
	r, err := g.Compile(ctx)
	if err != nil {
		logger.Error("Compile failed, err=%v", err)
		return
	}

	out, err := r.Invoke(ctx, map[string]any{
		"message_histories": []*schema.Message{},
		"user_query":        "我叫 zhangsan, 邮箱是 zhangsan@bytedance.com, 帮我推荐一处房产",
	})
	if err != nil {
		logger.Error("Invoke failed, err=%v", err)
		return
	}
	logger.Info("Generation: %v Messages", len(out))
	for _, msg := range out {
		logger.Info("    %v", msg)
	}
}

type userInfoRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userInfoResponse struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Company  string `json:"company"`
	Position string `json:"position"`
	Salary   string `json:"salary"`
}

type loggerCallbacks struct{}

func (l *loggerCallbacks) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	logger.Info("name: %v, type: %v, component: %v, input: %v", info.Name, info.Type, info.Component, input)
	return ctx
}

func (l *loggerCallbacks) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	logger.Info("name: %v, type: %v, component: %v, output: %v", info.Name, info.Type, info.Component, output)
	return ctx
}

func (l *loggerCallbacks) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	logger.Info("name: %v, type: %v, component: %v, error: %v", info.Name, info.Type, info.Component, err)
	return ctx
}

func (l *loggerCallbacks) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	return ctx
}

func (l *loggerCallbacks) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	return ctx
}
