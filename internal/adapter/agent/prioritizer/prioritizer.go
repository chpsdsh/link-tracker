package prioritizer

import (
	"strings"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/agent"
)

type Prioritizer struct {
	Config config.AIAgentConfig
}

func (p Prioritizer) FindPriority(description string) agent.Priority {
	for _, w := range p.Config.HighPriorityKeyWords {
		if strings.Contains(description, w) {
			return agent.PriorityHigh
		}
	}

	for _, w := range p.Config.LowPriorityKeyWords {
		if strings.Contains(description, w) {
			return agent.PriorityLow
		}
	}
	return agent.PriorityMedium
}
