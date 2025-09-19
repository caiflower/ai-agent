package caller

import (
	"github.com/caiflower/common-tools/cluster"
	"github.com/caiflower/common-tools/pkg/logger"
)

type DefaultCaller struct {
	cluster.DefaultCaller
}

func (c *DefaultCaller) MasterCall() {
	logger.Info("MasterCall")
}

func (c *DefaultCaller) SlaverCall(leaderName string) {
	logger.Info("SlaverCall leaderName=%v", leaderName)
}
