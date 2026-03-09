package ollama

import "ai-server/modules/types"

type Ai types.Ai

func (ai *Ai) Send() {
	ai.createRequest().fetch()
}
