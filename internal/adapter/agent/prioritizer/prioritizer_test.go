package prioritizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/agent"
)

func TestPrioritizingDescription(t *testing.T) {
	tests := []struct {
		name        string
		config      config.AIAgentConfig
		description string
		result      agent.Priority
	}{
		{
			name:        "high priority test",
			config:      config.AIAgentConfig{HighPriorityKeyWords: []string{"high", "high-priority"}},
			description: "High priority test",
			result:      agent.PriorityHigh,
		},
		{
			name:        "low priority test",
			config:      config.AIAgentConfig{HighPriorityKeyWords: []string{"high", "high-priority"}, LowPriorityKeyWords: []string{"low", "low-priority"}},
			description: "Low priority test",
			result:      agent.PriorityLow,
		},
		{
			name:        "medium priority test",
			config:      config.AIAgentConfig{HighPriorityKeyWords: []string{"high", "high-priority"}, LowPriorityKeyWords: []string{"low", "low-priority"}},
			description: "Medium priority test",
			result:      agent.PriorityMedium,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pr := Prioritizer{test.config}
			res := pr.FindPriority(test.description)
			assert.Equal(t, test.result, res)
		})
	}
}
