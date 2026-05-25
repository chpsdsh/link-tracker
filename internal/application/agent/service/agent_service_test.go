package service

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/agent/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/agent"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/pkg"
)

func TestFormatUpdate(t *testing.T) {
	tests := []struct {
		description string
		update      agent.UpdateToSend
		result      string
	}{
		{
			description: "single link update",
			update:      agent.UpdateToSend{GroupedUpdates: agent.GroupedUpdates{Updates: []agent.UpdateToGroup{{Description: "single link update", URL: "https://github/go"}}}},
			result:      "single link update\nСсылка: https://github/go\n",
		},
		{
			description: "multiple link updates",
			update: agent.UpdateToSend{
				GroupedUpdates: agent.GroupedUpdates{Updates: []agent.UpdateToGroup{
					{Description: "first link update", URL: "https://github/go"},
					{Description: "second link updates", URL: "https://github/go/new"},
				}},
			},
			result: "1) first link update\nСсылка: https://github/go\n2) second link updates\nСсылка: https://github/go/new\n",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			res := formatUpdate(test.update)
			assert.Equal(t, test.result, res)
		})
	}
}

func TestPriorityToString(t *testing.T) {
	tests := []struct {
		description string
		priority    agent.Priority
		result      string
		expectedErr error
	}{
		{
			description: "high priority",
			priority:    agent.PriorityHigh,
			result:      highPriority,
			expectedErr: nil,
		},
		{
			description: "medium priority",
			priority:    agent.PriorityMedium,
			result:      mediumPriority,
			expectedErr: nil,
		},
		{
			description: "low priority",
			priority:    agent.PriorityLow,
			result:      lowPriority,
			expectedErr: nil,
		},
		{
			description: "unknown priority",
			priority:    52,
			result:      "",
			expectedErr: ErrUnknownPriority,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			res, err := priorityToString(test.priority)
			assert.Equal(t, test.result, res)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestProcessLinkUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		description string
		update      pkg.LinkUpdate
		summarizer  func() UpdatesSummarizer
		prioritizer func() Prioritizer
		grouper     func() UpdatesGrouper
		result      error
	}{
		{
			description: "process success",
			update:      pkg.LinkUpdate{Description: "success", URL: "https://github/go", TgChatIDs: []int64{1, 2, 3}},
			summarizer: func() UpdatesSummarizer {
				summarizer := mocks.NewMockUpdatesSummarizer(ctrl)
				summarizer.EXPECT().Summarize(pkg.LinkUpdate{Description: "success", URL: "https://github/go", TgChatIDs: []int64{1, 2, 3}}).Return("success", nil)
				return summarizer
			},
			prioritizer: func() Prioritizer {
				prioritizer := mocks.NewMockPrioritizer(ctrl)
				prioritizer.EXPECT().
					FindPriority("success").
					Return(agent.PriorityMedium)
				return prioritizer
			},
			grouper: func() UpdatesGrouper {
				grouper := mocks.NewMockUpdatesGrouper(ctrl)
				grouper.EXPECT().
					Add(agent.UpdateToGroup{Description: "success", URL: "https://github/go", Priority: agent.PriorityMedium, TgChatIDs: []int64{1, 2, 3}})
				return grouper
			},
			result: nil,
		},
		{
			description: "process error",
			update:      pkg.LinkUpdate{Description: "failure", URL: "https://github/golang", TgChatIDs: []int64{1, 2, 6}},
			summarizer: func() UpdatesSummarizer {
				summarizer := mocks.NewMockUpdatesSummarizer(ctrl)
				summarizer.EXPECT().Summarize(pkg.LinkUpdate{Description: "failure", URL: "https://github/golang", TgChatIDs: []int64{1, 2, 6}}).Return("", errors.New("error"))
				return summarizer
			},
			prioritizer: func() Prioritizer {
				prioritizer := mocks.NewMockPrioritizer(ctrl)
				return prioritizer
			},
			grouper: func() UpdatesGrouper {
				grouper := mocks.NewMockUpdatesGrouper(ctrl)
				return grouper
			},
			result: ErrSummarizing,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			agentService := AgentService{
				Summarizer:  test.summarizer(),
				Prioritizer: test.prioritizer(),
				Grouper:     test.grouper(),
				BaseLogger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			}
			res := agentService.ProcessLinkUpdate(test.update)
			require.ErrorIs(t, res, test.result)
		})
	}
}

func TestFormAndSendUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		description string
		update      agent.UpdateToSend
		sender      func() Sender

		result error
	}{
		{
			description: "send success",
			update: agent.UpdateToSend{
				GroupedUpdates: agent.GroupedUpdates{Updates: []agent.UpdateToGroup{
					{Description: "first link update", URL: "https://github/go", TgChatIDs: []int64{1}},
					{Description: "second link updates", URL: "https://github/go/new", TgChatIDs: []int64{1}},
				},
					Priority: agent.PriorityHigh,
				},
				TgChatID: 1,
			},
			sender: func() Sender {
				s := mocks.NewMockSender(ctrl)
				s.EXPECT().SendLinkUpdate(pkg.ProcessedLinkUpdate{Description: "1) first link update\nСсылка: https://github/go\n2) second link updates\nСсылка: https://github/go/new\n",
					Priority:  highPriority,
					TgChatIDs: []int64{1}}, gomock.Any()).Return(nil)
				return s
			},
			result: nil,
		},
		{
			description: "send error",
			update: agent.UpdateToSend{
				GroupedUpdates: agent.GroupedUpdates{Updates: []agent.UpdateToGroup{
					{Description: "first link update", URL: "https://github/go", TgChatIDs: []int64{1}},
					{Description: "second link updates", URL: "https://github/go/new", TgChatIDs: []int64{1}},
				},
					Priority: agent.PriorityHigh,
				},
				TgChatID: 1,
			},
			sender: func() Sender {
				s := mocks.NewMockSender(ctrl)
				s.EXPECT().SendLinkUpdate(pkg.ProcessedLinkUpdate{Description: "1) first link update\nСсылка: https://github/go\n2) second link updates\nСсылка: https://github/go/new\n",
					Priority:  highPriority,
					TgChatIDs: []int64{1}}, gomock.Any()).Return(errors.New("error"))
				return s
			},
			result: ErrSendingUpdate,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			agentService := AgentService{
				Sender:     test.sender(),
				BaseLogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}
			res := agentService.FormAndSendUpdate(test.update)
			require.ErrorIs(t, res, test.result)
		})
	}
}
