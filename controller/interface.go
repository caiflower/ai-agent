package controller

import (
	apiv1 "github.com/caiflower/ai-agent/model/api/v1"
	"github.com/caiflower/common-tools/web/e"
)

type HealthController interface {
	DescribeHealth() string
}

type AgentController interface {
	Scheduling(request *apiv1.SchedulingRequest) (err e.ApiError)
}
