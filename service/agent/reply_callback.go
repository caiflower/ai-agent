package agent

import (
	"context"

	"github.com/caiflower/ai-agent/model/entity"
	"github.com/caiflower/common-tools/pkg/logger"
	"github.com/caiflower/common-tools/pkg/tools"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func newReplyCallback(_ context.Context, executeID string, _ map[string]struct{}) (clb callbacks.Handler,
	sr *schema.StreamReader[*entity.AgentRespEvent], sw *schema.StreamWriter[*entity.AgentRespEvent],
) {
	sr, sw = schema.Pipe[*entity.AgentRespEvent](10)

	rcc := &replyChunkCallback{
		sw:        sw,
		executeID: executeID,
		//returnDirectlyTools: returnDirectlyTools,
	}

	clb = callbacks.NewHandlerBuilder().
		OnStartFn(rcc.OnStart).
		OnEndFn(rcc.OnEnd).
		OnEndWithStreamOutputFn(rcc.OnEndWithStreamOutput).
		OnErrorFn(rcc.OnError).
		Build()

	return clb, sr, sw
}

type replyChunkCallback struct {
	sw                  *schema.StreamWriter[*entity.AgentRespEvent]
	executeID           string
	returnDirectlyTools map[string]struct{}
}

func (r *replyChunkCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	logger.Error("OnError - info=%v, error=%v", tools.ToJson(info), err)

	return ctx
}

func (r *replyChunkCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	logger.Info("OnStart - info=%v, input=%v", tools.ToJson(info), tools.ToJson(input))

	return ctx
}

func (r *replyChunkCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	logger.Info("OnEnd - info=%v, input=%v", tools.ToJson(info), tools.ToJson(output))

	switch info.Name {
	default:
		return ctx
	}
}

func (r *replyChunkCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput],
) context.Context {
	logger.Info("OnEndWithStreamOutput - info=%v, output=%v", tools.ToJson(info), tools.ToJson(output))

	switch info.Component {
	case components.ComponentOfChatModel:
		if info.Name != "ChatModel" {
			output.Close()
			return ctx
		}
		sr := schema.StreamReaderWithConvert(output, func(t callbacks.CallbackOutput) (*schema.Message, error) {
			cbOut := model.ConvCallbackOutput(t)
			return cbOut.Message, nil
		})

		r.sw.Send(&entity.AgentRespEvent{
			EventType:       entity.EventTypeOfChatModelAnswer,
			ChatModelAnswer: sr,
		}, nil)
		return ctx
	case compose.ComponentOfToolsNode:
		//toolsMessage, err := r.concatToolsNodeOutput(ctx, output)
		//if err != nil {
		//	r.sw.Send(nil, err)
		//	return ctx
		//}
		//
		//r.sw.Send(&entity.AgentEvent{
		//	EventType:    singleagent.EventTypeOfToolsMessage,
		//	ToolsMessage: toolsMessage,
		//}, nil)
		return ctx
	default:
		return ctx
	}
}
