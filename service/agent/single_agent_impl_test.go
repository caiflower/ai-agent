package agent

import (
	"io"
	"testing"

	mockchatmodel "github.com/caiflower/ai-agent/internal/mock/model"
	"github.com/caiflower/ai-agent/model/entity"
	chatmodel "github.com/caiflower/ai-agent/service/model"
	"github.com/caiflower/common-tools/pkg/bean"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAgentStreamExecute(t *testing.T) {
	agent := NewSingleAgent()
	ctl := gomock.NewController(t)
	factory := mockchatmodel.NewMockFactory(ctl)
	factory.EXPECT().CreateChatModel(chatmodel.ProtocolMock, &chatmodel.Config{}).Return(&chatmodel.MockChatModel{}, nil)
	bean.AddBean(agent)
	bean.AddBean(factory)
	bean.Ioc()

	sr, apiError := agent.StreamExecute(&entity.AgentRequest{
		Input:        schema.UserMessage("What's the weather like in Beijing?"),
		ChatProtocol: chatmodel.ProtocolMock,
	})
	assert.Equal(t, apiError, nil)

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

	assert.Equal(t, "the weather is good", message)
}
