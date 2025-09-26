package main

import (
	"context"
	"encoding/json"
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

func simpleAgent() {
	defer logger.DefaultLogger().Close()
	ctx := context.Background()

	callbacks.AppendGlobalHandlers(&loggerCallbacks{})

	userQuery := "我叫 zhangsan, 邮箱是 zhangsan@bytedance.com, 帮我推荐一处房产"

	// 1. create an instance of ChatTemplate as 1st Graph Node
	systemTpl := `你是一名房产经纪人，结合用户的薪酬和工作，使用 user_info API，为其提供相关的房产信息。邮箱是必须的`
	chatTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemTpl),
		schema.MessagesPlaceholder("message_histories", true),
		schema.UserMessage("{user_query}"),
	)

	chatTpl1 := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一名房产经纪人，结合用户的薪酬和工作, 为其提供相关的房产信息"),
		schema.ToolMessage("姓名:{name} 邮箱：{email} 公司: {company} 职位: {position} 薪水：{salary}", "user_info"),
		schema.MessagesPlaceholder("message_histories", true),
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
			NumPredict:    10000,      // 最大生成长度
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
		nodeKeyOfTemplate   = "template"
		nodeKeyOfTemplate1  = "template1"
		nodeKeyOfLambda     = "lambda"
		nodeKeyOfChatModel  = "chat_model"
		nodeKeyOfChatModel1 = "chat_model1"
		nodeKeyOfTools      = "tools"
	)

	// 6. create an instance of Graph
	// input type is 1st Graph Node's input type, that is ChatTemplate's input type: map[string]any
	// output type is last Graph Node's output type, that is ToolsNode's output type: []*schema.Message
	g := compose.NewGraph[map[string]any, *schema.Message]()

	// 7. add ChatTemplate into graph
	_ = g.AddChatTemplateNode(nodeKeyOfTemplate, chatTpl)

	// 8. add ChatModel into graph
	_ = g.AddChatModelNode(nodeKeyOfChatModel, toolModels)

	_ = g.AddChatModelNode(nodeKeyOfChatModel1, chatModel)

	// 9. add ToolsNode into graph
	_ = g.AddToolsNode(nodeKeyOfTools, toolsNode)

	_ = g.AddLambdaNode(nodeKeyOfLambda, compose.InvokableLambda[[]*schema.Message, map[string]any](func(ctx context.Context, input []*schema.Message) (output map[string]any, err error) {
		output = make(map[string]any)
		json.Unmarshal([]byte(input[0].Content), &output)

		return
	}))

	_ = g.AddChatTemplateNode(nodeKeyOfTemplate1, chatTpl1)

	// 10. add connection between nodes
	_ = g.AddEdge(compose.START, nodeKeyOfTemplate)

	_ = g.AddEdge(nodeKeyOfTemplate, nodeKeyOfChatModel)

	_ = g.AddEdge(nodeKeyOfChatModel, nodeKeyOfTools)

	_ = g.AddEdge(nodeKeyOfTools, nodeKeyOfLambda)

	_ = g.AddEdge(nodeKeyOfLambda, nodeKeyOfTemplate1)

	_ = g.AddEdge(nodeKeyOfTemplate1, nodeKeyOfChatModel1)

	_ = g.AddEdge(nodeKeyOfChatModel1, compose.END)

	// 9. compile Graph[I, O] to Runnable[I, O]
	r, err := g.Compile(ctx)
	if err != nil {
		logger.Error("Compile failed, err=%v", err)
		return
	}

	out, err := r.Invoke(ctx, map[string]any{
		"message_histories": []*schema.Message{},
		"user_query":        userQuery,
	})
	if err != nil {
		logger.Error("Invoke failed, err=%v", err)
		return
	}
	logger.Info("Generation: %v Messages", out)
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
