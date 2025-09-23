package entity

import (
	"github.com/cloudwego/eino/schema"
)

type AgentRequest struct {
	Input   *schema.Message
	History []*schema.Message
}

type EventType string

const (
	EventTypeOfChatModelAnswer        EventType = "chatmodel_answer"
	EventTypeOfToolsAsChatModelStream EventType = "tools_as_chatmodel_answer"
	EventTypeOfToolMidAnswer          EventType = "tool_mid_answer"
	EventTypeOfToolsMessage           EventType = "tools_message"
	EventTypeOfFuncCall               EventType = "func_call"
	EventTypeOfSuggest                EventType = "suggest"
	EventTypeOfKnowledge              EventType = "knowledge"
	EventTypeOfInterrupt              EventType = "interrupt"
)

type AgentRespEvent struct {
	EventType       EventType
	ChatModelAnswer *schema.StreamReader[*schema.Message]
}
