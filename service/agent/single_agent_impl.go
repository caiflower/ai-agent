package agent

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	"github.com/caiflower/ai-agent/constants"
	entity "github.com/caiflower/ai-agent/model/entity"
	"github.com/caiflower/ai-agent/service/model"
	"github.com/caiflower/common-tools/pkg/logger"
	"github.com/caiflower/common-tools/pkg/safego"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

const (
	KeyofChatModelNode   = "chat_model_node"
	keyOfPromptVariables = "prompt_variables"
	keyOfPromptTemplate  = "prompt_template"
)

type singleAgentImpl struct {
	Factory chatmodel.Factory `autowired:""`
}

func NewSingleAgent() SingleAgent {
	return &singleAgentImpl{}
}

func (sa *singleAgentImpl) StreamExecute(req *entity.AgentRequest) (*schema.StreamReader[*entity.AgentRespEvent], error) {
	var (
		g           = compose.NewGraph[*entity.AgentRequest, *schema.Message]()
		composeOpts []compose.Option
		pv          = promptVariables{}
		ctx         = context.Background()
		pt          = prompt.FromMessages(
			schema.Jinja2,
			schema.SystemMessage(ReactSystemPromptJinja2),
			schema.MessagesPlaceholder(placeholderOfChatHistory, true),
			schema.MessagesPlaceholder(placeholderOfUserInput, false),
		)
		executeID = uuid.New()
	)

	//callback handle
	hdl, sr, sw := newReplyCallback(executeID.String(), nil)
	composeOpts = append(composeOpts, compose.WithCallbacks(hdl))

	chatModel, err := sa.Factory.CreateChatModel(req.ChatProtocol, buildConfig(req))
	if err != nil {
		logger.Error("create cmodel failed. Error: %v", err)
		return nil, err
	}

	_ = g.AddLambdaNode(keyOfPromptVariables, compose.InvokableLambda[*entity.AgentRequest, map[string]any](pv.AssemblePromptVariables))
	_ = g.AddChatTemplateNode(keyOfPromptTemplate, pt)
	_ = g.AddChatModelNode(KeyofChatModelNode, chatModel, compose.WithNodeName("ChatModel"))

	_ = g.AddEdge(compose.START, keyOfPromptVariables)
	_ = g.AddEdge(keyOfPromptVariables, keyOfPromptTemplate)
	_ = g.AddEdge(keyOfPromptTemplate, KeyofChatModelNode)
	_ = g.AddEdge(KeyofChatModelNode, compose.END)

	runner, err := g.Compile(ctx)
	if err != nil {
		logger.Error("compile graph failed. Error: %v", err)
		return nil, err
	}

	safego.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("%s [ERROR] - Got a runtime error %s. %s\n%s", time.Now().Format("2006-01-02 15:04:05"), "StreamExecute", r, string(debug.Stack()))
				sw.Send(nil, errors.New("internal server error"))
			}

			sw.Close()
		}()

		_, _ = runner.Stream(ctx, req, composeOpts...)
	})

	return sr, nil
}

func buildConfig(req *entity.AgentRequest) (cfg *chatmodel.Config) {
	cfg = &chatmodel.Config{}
	switch req.ChatProtocol {
	case chatmodel.ProtocolOllama:
		cfg.BaseURL = constants.Prop.OLlama.Url
		cfg.Model = constants.Prop.OLlama.Model
	}
	return
}
