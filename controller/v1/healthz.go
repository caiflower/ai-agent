package v1

import "github.com/caiflower/ai-agent/controller"

type healthController struct {
}

func NewHealthController() controller.HealthController {
	return &healthController{}
}

func (c *healthController) DescribeHealth() string {
	return "healthz"
}
