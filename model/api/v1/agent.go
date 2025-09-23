package apiv1

import (
	"github.com/caiflower/ai-agent/model/api"
	"github.com/caiflower/ai-agent/model/entity"
	chatmodel "github.com/caiflower/ai-agent/service/model"
	"github.com/caiflower/common-tools/web"
)

type ChatRequest struct {
	api.Request
	web.Context
	Input        string             `verf:""`
	ChatProtocol chatmodel.Protocol `inList:"mock,ollama" verf:""`
}

type ChatEvent = entity.AgentRespEvent
