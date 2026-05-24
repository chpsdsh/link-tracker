package summarizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/config"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

func TestLinksSummarizerSummarize(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.AIAgentConfig
		update      pkg.LinkUpdate
		expected    string
		expectedErr error
	}{
		{
			name: "too short update",
			cfg: config.AIAgentConfig{
				MinLength:       10,
				Threshold:       100,
				StopWords:       []string{"spam"},
				ExcludedAuthors: []string{"bot-user"},
			},
			update: pkg.LinkUpdate{
				Description: "short",
			},
			expected:    "",
			expectedErr: ErrTooShortUpdate,
		},
		{
			name: "update contains banned word",
			cfg: config.AIAgentConfig{
				MinLength:       1,
				Threshold:       100,
				StopWords:       []string{"spam"},
				ExcludedAuthors: []string{"bot-user"},
			},
			update: pkg.LinkUpdate{
				Description: "This update contains spam message",
			},
			expected:    "",
			expectedErr: ErrContainsBannedWords,
		},
		{
			name: "update contains banned author",
			cfg: config.AIAgentConfig{
				MinLength:       1,
				Threshold:       100,
				StopWords:       []string{"spam"},
				ExcludedAuthors: []string{"bot-user"},
			},
			update: pkg.LinkUpdate{
				Description: "Автор: bot-user обновил ссылку",
			},
			expected:    "",
			expectedErr: ErrContainsBannedAuthor,
		},
		{
			name: "valid update without summarization",
			cfg: config.AIAgentConfig{
				MinLength:       1,
				Threshold:       100,
				StopWords:       []string{"spam"},
				ExcludedAuthors: []string{"bot-user"},
			},
			update: pkg.LinkUpdate{
				Description: "Normal useful update",
			},
			expected:    "Normal useful update",
			expectedErr: nil,
		},
		{
			name: "valid update with empty filters",
			cfg: config.AIAgentConfig{
				MinLength:       1,
				Threshold:       100,
				StopWords:       nil,
				ExcludedAuthors: nil,
			},
			update: pkg.LinkUpdate{
				Description: "Normal update without any filters",
			},
			expected:    "Normal update without any filters",
			expectedErr: nil,
		},
		{
			name: "description length equals min length",
			cfg: config.AIAgentConfig{
				MinLength: 5,
				Threshold: 100,
			},
			update: pkg.LinkUpdate{
				Description: "12345",
			},
			expected:    "12345",
			expectedErr: nil,
		},
		{
			name: "description length equals threshold",
			cfg: config.AIAgentConfig{
				MinLength: 1,
				Threshold: 5,
			},
			update: pkg.LinkUpdate{
				Description: "12345",
			},
			expected:    "12345",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := LinksSummarizer{
				Config: tt.cfg,
			}

			actual, err := s.Summarize(tt.update)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Empty(t, actual)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
