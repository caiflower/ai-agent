package chatmodel

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type MockChatModel struct{}

func mockChatModelBuilder(config *Config) (model.ToolCallingChatModel, error) {
	return &MockChatModel{}, nil
}

func (m *MockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return schema.AssistantMessage("the weather is good", nil), nil
}

func (m *MockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](0)
	go func() {
		defer sw.Close()
		sw.Send(schema.AssistantMessage("the weather is", nil), nil)
		sw.Send(schema.AssistantMessage(" good", nil), nil)
	}()
	return sr, nil
}

func (m *MockChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	panic("implement me")
}
