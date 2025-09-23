package apiv1

import (
	"github.com/caiflower/ai-agent/model/api"
	"github.com/caiflower/ai-agent/model/entity"
	"github.com/caiflower/common-tools/web"
)

type ChatRequest struct {
	api.Request
	web.Context
	Input string `json:"input"`
}

type ChatEvent = entity.AgentRespEvent
