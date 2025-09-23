package agent

import (
	"errors"
	"io"

	apiv1 "github.com/caiflower/ai-agent/model/api/v1"
	"github.com/caiflower/ai-agent/model/entity"
	"github.com/caiflower/common-tools/pkg/logger"
	"github.com/cloudwego/eino/schema"
)

type agentRuntime struct {
	SingleAgent SingleAgent `autowired:""`
}

func NewAgentRuntime() Runtime {
	return &agentRuntime{}
}

func (r *agentRuntime) Run(request *entity.AgentRequest) (*schema.StreamReader[*apiv1.ChatEvent], error) {
	//var (
	//	mainChan = make(chan *entity.AgentRespEvent, 100)
	//)

	singleAgentSr, err := r.SingleAgent.StreamExecute(request)
	if err != nil {
		logger.Error("runtime run failed. Error: %v", err)
		return nil, err
	}

	return singleAgentSr, nil
	//sr, sw := schema.Pipe[*apiv1.ChatEvent](10)
	//
	//safego.Go(func() {
	//	r.pull(mainChan, singleAgentSr)
	//})
	//
	//safego.Go(func() {
	//	r.push(mainChan)
	//})
	//
	//return sr, nil
}

func (r *agentRuntime) pull(ch chan<- *entity.AgentRespEvent, sr *schema.StreamReader[*apiv1.ChatEvent]) {
	for {
		recv, recvErr := sr.Recv()
		if recvErr != nil {
			if errors.Is(recvErr, io.EOF) {
				return
			}
		}
		switch recv.EventType {
		case entity.EventTypeOfChatModelAnswer:
			for {
				message, recvErr := recv.ChatModelAnswer.Recv()
				if recvErr != nil {

				}
				logger.Info("message: %v", message)
			}
		default:
			logger.Warn("unknown eventType: %s", recv.EventType)
		}
	}

}

func (r *agentRuntime) push(ch <-chan *entity.AgentRespEvent) {

}
