package v1

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/caiflower/ai-agent/controller"
	apiv1 "github.com/caiflower/ai-agent/model/api/v1"
	"github.com/caiflower/common-tools/web/e"
	"github.com/tmaxmax/go-sse"
)

var (
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
)

type agentController struct {
}

func NewAgentController() controller.AgentController {
	return &agentController{}
}

func (c *agentController) Scheduling(request *apiv1.SchedulingRequest) (err e.ApiError) {
	sseSever := newSSE(request)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer sseSever.Shutdown(context.Background())

	go func() {
		cnt := 0
		for {
			select {
			case <-ctx.Done():

			default:
				time.Sleep(time.Second)
				if cnt == 3 {
					message := &sse.Message{
						Type: sse.Type("close"),
					}
					_ = sseSever.Publish(message, request.UserId)
					return
				}

				cnt++
				err1 := sseSever.Publish(generateRandomNumbers(), request.UserId)
				if err1 != nil {
					logger.Error("publish err", err1)
				}
			}

		}
	}()

	sseSever.ServeHTTP(request.GetResponseWriterAndRequest())
	cancelFunc()

	return nil
}

func newSSE(request *apiv1.SchedulingRequest) *sse.Server {
	rp, _ := sse.NewValidReplayer(time.Minute*5, true)
	rp.GCInterval = time.Minute

	return &sse.Server{
		Provider: &sse.Joe{Replayer: rp},
		// If you are using a 3rd party library to generate a per-request logger, this
		// can just be a simple wrapper over it.
		Logger: func(r *http.Request) *slog.Logger {
			return logger.With("userId", request.UserId).With("reqId", request.RequestId)
		},
		OnSession: func(w http.ResponseWriter, r *http.Request) (topics []string, permitted bool) {
			// the shutdown message is sent on the default topic
			return []string{request.UserId}, true
		},
	}
}

func generateRandomNumbers() *sse.Message {
	message := &sse.Message{}
	count := 1 + rand.Intn(5)

	for i := 0; i < count; i++ {
		message.AppendData(strconv.FormatUint(rand.Uint64(), 10))
	}

	return message
}
