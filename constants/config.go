package constants

import (
	"github.com/caiflower/common-tools/global/config"
)

var DefaultConfig = config.DefaultConfig{}
var Prop = Config{}

func InitConfig() {
	if err := config.LoadDefaultConfig(&DefaultConfig); err != nil {
		panic(err)
	}
	if err := config.LoadYamlFile("config.yaml", &Prop); err != nil {
		panic(err)
	}
}

type Config struct {
	Prompt PromptConfig `yaml:"prompt"`
	OLlama OLlamaConfig `yaml:"ollama"`
}

type PromptConfig struct {
	AgentName string `yaml:"agentName" default:"全能助手"`
}

type OLlamaConfig struct {
	Url   string `yaml:"url"`
	Model string `yaml:"model"`
}
