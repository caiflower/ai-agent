package chatmodel

import (
	"fmt"

	"github.com/cloudwego/eino/components/model"
)

type Builder func(config *Config) (model.ToolCallingChatModel, error)

type defaultFactory struct {
	protocol2Builder map[Protocol]Builder
}

func NewDefaultFactory() Factory {
	return &defaultFactory{
		protocol2Builder: map[Protocol]Builder{
			ProtocolOllama: ollamaBuilder,
		},
	}
}

func (f *defaultFactory) SupportProtocol(protocol Protocol) bool {
	_, found := f.protocol2Builder[protocol]
	return found
}

func (f *defaultFactory) CreateChatModel(protocol Protocol, config *Config) (model.ToolCallingChatModel, error) {
	if config == nil {
		return nil, fmt.Errorf("[CreateChatModel] config not provided")
	}

	builder, found := f.protocol2Builder[protocol]
	if !found {
		return nil, fmt.Errorf("[CreateChatModel] protocol not support, protocol=%s", protocol)
	}

	return builder(config)
}
