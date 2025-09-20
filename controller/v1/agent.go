package v1

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/caiflower/ai-agent/controller"
	apiv1 "github.com/caiflower/ai-agent/model/api/v1"
	"github.com/caiflower/common-tools/pkg/logger"
	"github.com/caiflower/common-tools/web"
	"github.com/caiflower/common-tools/web/e"
	"github.com/tmaxmax/go-sse"
)

type agentController struct {
	SSEProvider sse.Provider `autowired:""`
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

func (c *agentController) Scheduling(request *apiv1.SchedulingRequest) (err e.ApiError) {
	sessionIds := []string{request.SessionId}

	go func() {
		cnt := 0
		for {
			time.Sleep(time.Second)

			if cnt == 3 {
				message := &sse.Message{
					Type: sse.Type("close"),
				}
				_ = c.SSEProvider.Publish(message, sessionIds)

				logger.Info("sessionIds %s closed", sessionIds)
				return
			}

			err1 := c.SSEProvider.Publish(generateRandomNumbers(request.SessionId), sessionIds)
			if err1 != nil {
				logger.Error("publish failed. Error: %v", err1)
			}
			cnt++
		}
	}()

	sseErr := c.beginSse(sessionIds, &request.Context)
	if sseErr != nil {
		return e.NewInternalError(sseErr)
	}

	return
}

func (c *agentController) beginSse(sessionIds []string, webCtx *web.Context) error {
	logger.Info("beginSse sessionIds %s", sessionIds)
	w, r := webCtx.GetResponseWriterAndRequest()
	sess, err := sse.Upgrade(w, r)
	if err != nil {
		logger.Error("upgrade sse failed. Error: %v", err)
		return err
	}

	sub := sse.Subscription{Client: sess, LastEventID: sess.LastEventID, Topics: sessionIds}

	err = c.SSEProvider.Subscribe(r.Context(), sub)
	if err != nil {
		logger.Error("sse subscribe failed. Error: %v", err)
		return err
	}

	return nil
}

func generateRandomNumbers(sessionId string) *sse.Message {
	message := &sse.Message{}
	message.AppendData(sessionId + ":" + strconv.FormatUint(rand.Uint64(), 10))

	return message
}
