package agent

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/caiflower/ai-agent/model/entity"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
)

const (
	nodeOfModel  = "model"
	nodeOfPrompt = "prompt"
)

type mockChatModel struct{}

func (m *mockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return schema.AssistantMessage("the weather is good", nil), nil
}

func (m *mockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](0)
	go func() {
		defer sw.Close()
		sw.Send(schema.AssistantMessage("the weather is", nil), nil)
		sw.Send(schema.AssistantMessage("good", nil), nil)
	}()
	return sr, nil
}

func (m *mockChatModel) BindTools(tools []*schema.ToolInfo) error {
	panic("implement me")
}

func TestM(m *testing.T) {
	ctx := context.Background()
	g := compose.NewGraph[map[string]any, *schema.Message]()

	pt := prompt.FromMessages(
		schema.FString,
		schema.UserMessage("what's the weather in {location}?"),
	)

	_ = g.AddChatTemplateNode(nodeOfPrompt, pt)
	_ = g.AddChatModelNode(nodeOfModel, &mockChatModel{}, compose.WithNodeName("ChatModel"))
	_ = g.AddEdge(compose.START, nodeOfPrompt)
	_ = g.AddEdge(nodeOfPrompt, nodeOfModel)
	_ = g.AddEdge(nodeOfModel, compose.END)

	r, err := g.Compile(ctx, compose.WithMaxRunSteps(10))
	if err != nil {
		panic(err)
	}

	in := map[string]any{"location": "beijing"}
	ret, err := r.Invoke(ctx, in)
	fmt.Println("invoke result: ", ret)

	// stream
	s, err := r.Stream(ctx, in)
	if err != nil {
		panic(err)
	}

	defer s.Close()
	for {
		chunk, err := s.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		fmt.Println("stream chunk: ", chunk)
	}
}

func TestAgentStreamExecute(t *testing.T) {
	agent := NewSingleAgent()
	sr, apiError := agent.StreamExecute(&entity.AgentRequest{
		Input: schema.UserMessage("介绍一下北京"),
		//History: []*schema.Message{schema.UserMessage("介绍一下天津")},
	})
	assert.Equal(t, apiError, nil, "failed")

	agentEventSr, err := sr.Recv()
	if err != nil {
		if err == io.EOF {
			return
		}
	}

	message := ""
	for {
		chunk, err := agentEventSr.ChatModelAnswer.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			assert.Equal(t, err, nil)
			return
		}

		message += chunk.Content
	}

	fmt.Println("result: ", message)
}
