package apiv1

import (
	"github.com/caiflower/ai-agent/model/api"
	"github.com/caiflower/common-tools/web"
)

type SchedulingRequest struct {
	api.Request
	web.Context
}

type SchedulingResponse struct {
}
