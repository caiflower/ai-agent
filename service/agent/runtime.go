package agent

import (
	apiv1 "github.com/caiflower/ai-agent/model/api/v1"
	"github.com/caiflower/ai-agent/model/entity"
	"github.com/cloudwego/eino/schema"
)

type Runtime interface {
	Run(*entity.AgentRequest) (*schema.StreamReader[*apiv1.ChatEvent], error)
}
