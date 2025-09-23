package v1

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/caiflower/ai-agent/service/agent"
	"github.com/caiflower/ai-agent/service/xsse"
	"github.com/caiflower/common-tools/pkg/bean"
	"github.com/caiflower/common-tools/pkg/logger"
	. "github.com/caiflower/common-tools/web/v1"
	"github.com/stretchr/testify/assert"
	"github.com/tmaxmax/go-sse"
)

func TestChat(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockServer := NewHttpServer(Config{
		Name: "mockSever",
		Port: 8081,
	})

	bean.AddBean(xsse.NewSSEProvider())
	bean.AddBean(agent.NewAgentRuntime())
	bean.AddBean(agent.NewSingleAgent())
	mockServer.AddController(NewAgentController())
	bean.Ioc()

	mockServer.Register(NewRestFul().Method(http.MethodGet).Version("v1").Controller("v1.agentController").Path("/chat").Action("Chat"))
	mockServer.StartUp()

	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:8081/v1/chat?input=%E4%BB%8B%E7%BB%8D%E4%B8%80%E4%B8%8B%E5%8C%97%E4%BA%AC", http.NoBody)
	r.Header.Set("X-User-Id", "test-user")
	conn := sse.NewConnection(r)

	message := ""
	conn.SubscribeToAll(func(event sse.Event) {
		switch event.Type {
		case EventTypeOfChatModelAnswer:
			message += event.Data
		case EventTypeOfChatError:
			cancel()
			logger.Error("chat failed. Error: %v", event.Data)
			return
		case EventTypeOfChatFinish:
			cancel()
			return
		}
	})

	if err := conn.Connect(); err != nil {
		assert.Contains(t, err.Error(), "context canceled")
	}
	fmt.Println(message)
}
