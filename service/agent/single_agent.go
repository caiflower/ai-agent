package agent

import (
	entity "github.com/caiflower/ai-agent/model/entity"
	"github.com/cloudwego/eino/schema"
)

//go:generate mockgen -destination ../../internal/mock/sigle_agent_mock.go -package agent -source single_agent.go
type SingleAgent interface {
	StreamExecute(req *entity.AgentRequest) (*schema.StreamReader[*entity.AgentRespEvent], error)
}
