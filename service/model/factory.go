package chatmodel

import (
	"time"

	"github.com/cloudwego/eino/components/model"
)

type Config struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

//go:generate mockgen -destination ../../internal/mock/model/factory_mock.go -package chatmodel -source factory.go
type Factory interface {
	CreateChatModel(protocol Protocol, config *Config) (model.ToolCallingChatModel, error)
	SupportProtocol(protocol Protocol) bool
}
