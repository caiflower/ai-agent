package agent

import (
	"errors"
	"runtime/debug"
	"time"

	entity "github.com/caiflower/ai-agent/model/entity"
	golocalv1 "github.com/caiflower/common-tools/pkg/golocal/v1"
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
}

func NewSingleAgent() SingleAgent {
	return &singleAgentImpl{}
}

func (sa *singleAgentImpl) StreamExecute(req *entity.AgentRequest) (*schema.StreamReader[*entity.AgentRespEvent], error) {
	var (
		g           = compose.NewGraph[*entity.AgentRequest, *schema.Message]()
		ctx         = golocalv1.GetContext()
		composeOpts []compose.Option
		pv          = promptVariables{}

		pt = prompt.FromMessages(
			schema.Jinja2,
			schema.SystemMessage(ReactSystemPromptJinja2),
			schema.MessagesPlaceholder(placeholderOfChatHistory, true),
			schema.MessagesPlaceholder(placeholderOfUserInput, false),
		)
		executeID = uuid.New()
	)

	//callback handle
	hdl, sr, sw := newReplyCallback(ctx, executeID.String(), nil)
	composeOpts = append(composeOpts, compose.WithCallbacks(hdl))

	chatModel, err := newOllamaModel(ctx)
	if err != nil {
		logger.Error("create model failed. Error: %v", err)
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
