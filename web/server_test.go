package web

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/caiflower/ai-agent/constants"
	"github.com/caiflower/ai-agent/controller/v1"
	mockchatmodel "github.com/caiflower/ai-agent/internal/mock/model"
	"github.com/caiflower/ai-agent/service/agent"
	chatmodel "github.com/caiflower/ai-agent/service/model"
	"github.com/caiflower/ai-agent/service/xsse"
	"github.com/caiflower/common-tools/pkg/bean"
	xhttp "github.com/caiflower/common-tools/pkg/http"
	"github.com/caiflower/common-tools/pkg/logger"
	"github.com/caiflower/common-tools/web/e"
	. "github.com/caiflower/common-tools/web/v1"
	"github.com/stretchr/testify/assert"
	"github.com/tmaxmax/go-sse"
	"go.uber.org/mock/gomock"
)

type CommonResponse struct {
	RequestId string
	Data      interface{} `json:",omitempty"`
	Error     *e.Error
}

func TestChat(t *testing.T) {
	ctl := gomock.NewController(t)

	mockServer := NewHttpServer(Config{
		Name: "mockSever",
		Port: 8081,
	})
	factory := mockchatmodel.NewMockFactory(ctl)
	factory.EXPECT().CreateChatModel(chatmodel.ProtocolMock, &chatmodel.Config{}).Return(&chatmodel.MockChatModel{}, nil)

	bean.AddBean(xsse.NewSSEProvider())
	bean.AddBean(agent.NewAgentRuntime())
	bean.AddBean(agent.NewSingleAgent())
	bean.AddBean(factory)
	mockServer.AddController(v1.NewAgentController())
	bean.Ioc()

	mockServer.AddInterceptor(NewUserInterceptor(), 0)
	mockServer.Register(NewRestFul().Method(http.MethodGet).Version("v1").Controller("v1.agentController").Path("/chat").Action("Chat"))
	mockServer.StartUp()
	time.Sleep(1 * time.Second)
	defer mockServer.Close()

	// v1.agentController.Chat /v1/chat
	chatV1(t)
}

func chatV1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := xhttp.NewHttpClient(xhttp.Config{})

	mockCompare(t,
		"Not login",
		c,
		http.MethodGet,
		"http://127.0.0.1:8081/v1/chat?input=%E4%BB%8B%E7%BB%8D%E4%B8%80%E4%B8%8B%E5%8C%97%E4%BA%AC&chatProtocol=mock",
		nil,
		nil,
		&CommonResponse{
			Error: &e.Error{
				Code:    constants.NotLoginError.Code,
				Type:    constants.NotLoginError.Type,
				Message: "not login",
			},
		})

	headers := make(map[string]string)
	headers["X-User-Id"] = "test-user"
	mockCompare(t,
		"Missing chatProtocol",
		c, http.MethodGet,
		"http://127.0.0.1:8081/v1/chat?input=%E4%BB%8B%E7%BB%8D%E4%B8%80%E4%B8%8B%E5%8C%97%E4%BA%AC",
		headers,
		nil,
		&CommonResponse{
			Error: &e.Error{
				Code:    e.InvalidArgument.Code,
				Type:    e.InvalidArgument.Type,
				Message: "ChatRequest.ChatProtocol is missing",
			},
		})

	mockCompare(t,
		"Missing input",
		c, http.MethodGet,
		"http://127.0.0.1:8081/v1/chat?chatProtocol=mock",
		headers,
		nil,
		&CommonResponse{
			Error: &e.Error{
				Code:    e.InvalidArgument.Code,
				Type:    e.InvalidArgument.Type,
				Message: "ChatRequest.Input is missing",
			},
		})

	mockCompare(t,
		"ChatProtocol not in list",
		c, http.MethodGet,
		"http://127.0.0.1:8081/v1/chat?input=%E4%BB%8B%E7%BB%8D%E4%B8%80%E4%B8%8B%E5%8C%97%E4%BA%AC&chatProtocol=test",
		headers,
		nil,
		&CommonResponse{
			Error: &e.Error{
				Code:    e.InvalidArgument.Code,
				Type:    e.InvalidArgument.Type,
				Message: "ChatRequest.ChatProtocol is not in [mock ollama]",
			},
		})

	req, _ := http.NewRequestWithContext(ctx,
		http.MethodGet,
		"http://127.0.0.1:8081/v1/chat?input=%E4%BB%8B%E7%BB%8D%E4%B8%80%E4%B8%8B%E5%8C%97%E4%BA%AC&chatProtocol=mock",
		http.NoBody)
	req.Header.Set("X-User-Id", "test-user")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	message := ""
	for ev, err := range sse.Read(res.Body, nil) {
		if err != nil {
			assert.Contains(t, err.Error(), "cancel")
			break
		}
		switch ev.Type {
		case v1.EventTypeOfChatModelAnswer:
			message += ev.Data
		case v1.EventTypeOfChatError:
			logger.Error("chat failed. Error: %v", ev.Data)
			cancel()
		case v1.EventTypeOfChatFinish:
			cancel()
		}
	}

	assert.Equal(t, message, "the weather is good")
}

func mockCompare(t *testing.T, testCaseName string, c xhttp.HttpClient, method string, url string, headers map[string]string, body interface{}, want *CommonResponse) {
	res := &CommonResponse{}
	response := &xhttp.Response{
		Data: res,
	}
	err := c.Do(method, "", url, xhttp.ContentTypeJson, body, nil, response, headers)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, want.Data, res.Data, testCaseName)
	assert.Equal(t, want.Error, res.Error, testCaseName)
}
