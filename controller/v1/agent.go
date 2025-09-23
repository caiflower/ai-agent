package v1

import (
	"context"
	"io"

	"github.com/caiflower/ai-agent/controller"
	apiv1 "github.com/caiflower/ai-agent/model/api/v1"
	entity "github.com/caiflower/ai-agent/model/entity"
	"github.com/caiflower/ai-agent/service/agent"
	"github.com/caiflower/common-tools/pkg/logger"
	"github.com/caiflower/common-tools/pkg/safego"
	"github.com/caiflower/common-tools/web"
	"github.com/caiflower/common-tools/web/e"
	"github.com/cloudwego/eino/schema"
	"github.com/tmaxmax/go-sse"
)

const (
	EventTypeOfChatModelAnswer = "chat.answer"
	EventTypeOfChatError       = "chat.error"
	EventTypeOfChatFinish      = "chat.finish"
)

type agentController struct {
	SSEProvider  sse.Provider  `autowired:""`
	AgentRuntime agent.Runtime `autowired:""`
}

func NewAgentController() controller.AgentController {
	return &agentController{}
}

func (c *agentController) Close() {
	if c.SSEProvider != nil {
		if err := c.SSEProvider.Shutdown(context.Background()); err != nil {
			logger.Error("shutdown SSE provider failed. Error: %v", err)
		} else {
			logger.Info("shutdown SSE provider success.")
		}
	}
}

func (c *agentController) Chat(request *apiv1.ChatRequest) e.ApiError {
	var (
		topics = []string{request.RequestID}
	)

	sr, err := c.AgentRuntime.Run(&entity.AgentRequest{
		Input:        schema.UserMessage(request.Input),
		ChatProtocol: request.ChatProtocol,
	})
	if err != nil {
		logger.Error("agent run failed. Error: %v", err)
		return e.NewInternalError(err)
	}

	safego.Go(func() {
		for {
			chatEventRecv, recvErr := sr.Recv()
			if recvErr != nil {
				if recvErr == io.EOF {
					_ = c.SSEProvider.Publish(buildChatMessage(EventTypeOfChatFinish, "finish"), topics)
					break
				}
				_ = c.SSEProvider.Publish(buildChatMessage(EventTypeOfChatError, "chat failed"), topics)
				logger.Error("chat receive failed. Error: %v", recvErr)
				return
			}

			switch chatEventRecv.EventType {
			case entity.EventTypeOfChatModelAnswer:
				for {
					message, recvErr := chatEventRecv.ChatModelAnswer.Recv()
					if recvErr != nil {
						if recvErr == io.EOF {
							break
						}
						_ = c.SSEProvider.Publish(buildChatMessage(EventTypeOfChatError, "chat failed"), topics)
						logger.Error("chat receive failed. Error: %v", recvErr)
						return
					}
					_ = c.SSEProvider.Publish(buildChatAnswerMessage(message), topics)
				}
			default:
				logger.Warn("chat receive unknown event: %v", chatEventRecv.EventType)
			}
		}
	})

	sseErr := c.beginSse(topics, &request.Context)
	if sseErr != nil {
		return e.NewInternalError(sseErr)
	}

	return nil
}

func (c *agentController) beginSse(sessionIds []string, webCtx *web.Context) error {
	logger.Info("beginSse sessionIds %s", sessionIds)
	w, r := webCtx.GetResponseWriterAndRequest()
	sess, err := sse.Upgrade(w, r)
	if err != nil {
		logger.Error("upgrade xsse failed. Error: %v", err)
		return err
	}

	sub := sse.Subscription{Client: sess, LastEventID: sess.LastEventID, Topics: sessionIds}

	err = c.SSEProvider.Subscribe(r.Context(), sub)
	if err != nil {
		logger.Error("xsse subscribe failed. Error: %v", err)
		return err
	}

	return nil
}

func buildChatAnswerMessage(message *schema.Message) *sse.Message {
	msg := &sse.Message{
		Type: sse.Type(EventTypeOfChatModelAnswer),
	}
	msg.AppendData(message.Content)

	return msg
}

func buildChatMessage(_type string, message string) *sse.Message {
	msg := &sse.Message{
		Type: sse.Type(_type),
	}
	msg.AppendData(message)

	return msg
}
