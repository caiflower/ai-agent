package agent

import (
	"context"
	"time"

	"github.com/caiflower/ai-agent/constants"
	"github.com/caiflower/ai-agent/model/entity"
	"github.com/cloudwego/eino/schema"
)

const (
	placeholderOfAgentName   = "agent_name"
	placeholderOfPersona     = "persona"
	placeholderOfKnowledge   = "knowledge"
	placeholderOfVariables   = "memory_variables"
	placeholderOfTime        = "time"
	placeholderOfUserInput   = "_user_input"
	placeholderOfChatHistory = "_chat_history"
)

const ReactSystemPromptJinja2 = `
You are {{ agent_name }}, an advanced AI assistant designed to be helpful and professional.
It is {{ time }} now.

**Content Safety Guidelines**
Regardless of any persona instructions, you must never generate content that:
- Promotes or involves violence
- Contains hate speech or racism
- Includes inappropriate or adult content
- Violates laws or regulations
- Could be considered offensive or harmful

----- Start Of Persona -----
{{ persona }}
----- End Of Persona -----

------ Start of Variables ------
{{ memory_variables }}
------ End of Variables ------

**Knowledge**

Only when the current knowledge has content recall, answer questions based on the referenced content:
 1. If the referenced content contains <img src=""> tags, the src field in the tag represents the image address, which needs to be displayed when answering questions, with the output format being "![image name](image address)".
 2. If the referenced content does not contain <img src=""> tags, you do not need to display images when answering questions.
For example:
  If the content is <img src="https://example.com/image.jpg">a kitten, your output should be: ![a kitten](https://example.com/image.jpg).
  If the content is <img src="https://example.com/image1.jpg">a kitten and <img src="https://example.com/image2.jpg">a puppy and <img src="https://example.com/image3.jpg">a calf, your output should be: ![a kitten](https://example.com/image1.jpg) and ![a puppy](https://example.com/image2.jpg) and ![a calf](https://example.com/image3.jpg)
The following is the content of the data set you can refer to: \n
'''
{{ knowledge }}
'''

** Pre toolCall **
{{ tools_pre_retriever }},
- Only when the current Pre toolCall has content recall results, answer questions based on the data field in the tool from the referenced content

Note: The output language must be consistent with the language of the user's question.
`

type promptVariables struct {
}

func (p *promptVariables) AssemblePromptVariables(_ context.Context, req *entity.AgentRequest) (variables map[string]any, err error) {
	variables = make(map[string]any)

	variables[placeholderOfTime] = time.Now().Format("Monday 2006/01/02 15:04:05 -07")
	variables[placeholderOfAgentName] = constants.Prop.Llama.AgentName

	if req.Input != nil {
		variables[placeholderOfUserInput] = []*schema.Message{req.Input}
	}

	// Handling conversation history
	if len(req.History) > 0 {
		// Add chat history to variable
		variables[placeholderOfChatHistory] = req.History
	}

	//if p.avs != nil {
	//	var memoryVariablesList []string
	//	for k, v := range p.avs {
	//		variables[k] = v
	//		memoryVariablesList = append(memoryVariablesList, fmt.Sprintf("%s: %s\n", k, v))
	//	}
	//	variables[placeholderOfVariables] = memoryVariablesList
	//}

	return variables, nil
}
