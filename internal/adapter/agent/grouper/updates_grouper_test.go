package grouper

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/adapter/agent/mocks"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/agent"
)

func TestUpdatesGrouperAdd(t *testing.T) {
	tests := []struct {
		name     string
		updates  []agent.UpdateToGroup
		expected map[int64]agent.GroupedUpdates
	}{
		{
			name: "add single update for one chat",
			updates: []agent.UpdateToGroup{
				{
					Description: "critical update",
					TgChatIDs:   []int64{111},
					Priority:    agent.PriorityHigh,
				},
			},
			expected: map[int64]agent.GroupedUpdates{
				111: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "critical update",
							TgChatIDs:   []int64{111},
							Priority:    agent.PriorityHigh,
						},
					},
					Priority: agent.PriorityHigh,
				},
			},
		},
		{
			name: "add multiple updates for same chat and keep highest priority",
			updates: []agent.UpdateToGroup{
				{
					Description: "docs update",
					TgChatIDs:   []int64{111},
					Priority:    agent.PriorityLow,
				},
				{
					Description: "regular update",
					TgChatIDs:   []int64{111},
					Priority:    agent.PriorityMedium,
				},
				{
					Description: "critical update",
					TgChatIDs:   []int64{111},
					Priority:    agent.PriorityHigh,
				},
			},
			expected: map[int64]agent.GroupedUpdates{
				111: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "docs update",
							TgChatIDs:   []int64{111},
							Priority:    agent.PriorityLow,
						},
						{
							Description: "regular update",
							TgChatIDs:   []int64{111},
							Priority:    agent.PriorityMedium,
						},
						{
							Description: "critical update",
							TgChatIDs:   []int64{111},
							Priority:    agent.PriorityHigh,
						},
					},
					Priority: agent.PriorityHigh,
				},
			},
		},
		{
			name: "add one update for multiple chats",
			updates: []agent.UpdateToGroup{
				{
					Description: "urgent update",
					TgChatIDs:   []int64{111, 222},
					Priority:    agent.PriorityHigh,
				},
			},
			expected: map[int64]agent.GroupedUpdates{
				111: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "urgent update",
							TgChatIDs:   []int64{111, 222},
							Priority:    agent.PriorityHigh,
						},
					},
					Priority: agent.PriorityHigh,
				},
				222: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "urgent update",
							TgChatIDs:   []int64{111, 222},
							Priority:    agent.PriorityHigh,
						},
					},
					Priority: agent.PriorityHigh,
				},
			},
		},
		{
			name: "add updates for different chats",
			updates: []agent.UpdateToGroup{
				{
					Description: "minor update",
					TgChatIDs:   []int64{111},
					Priority:    agent.PriorityLow,
				},
				{
					Description: "breaking update",
					TgChatIDs:   []int64{222},
					Priority:    agent.PriorityHigh,
				},
			},
			expected: map[int64]agent.GroupedUpdates{
				111: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "minor update",
							TgChatIDs:   []int64{111},
							Priority:    agent.PriorityLow,
						},
					},
					Priority: agent.PriorityLow,
				},
				222: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "breaking update",
							TgChatIDs:   []int64{222},
							Priority:    agent.PriorityHigh,
						},
					},
					Priority: agent.PriorityHigh,
				},
			},
		},
		{
			name: "add update without chat ids",
			updates: []agent.UpdateToGroup{
				{
					Description: "update without chats",
					TgChatIDs:   nil,
					Priority:    agent.PriorityMedium,
				},
			},
			expected: map[int64]agent.GroupedUpdates{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockUpdatesHandler(ctrl)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			grouper := NewUpdatesGrouper(logger, handler)

			for _, update := range tt.updates {
				grouper.Add(update)
			}

			assert.Equal(t, tt.expected, grouper.TgChatsMap)
		})
	}
}

func TestUpdatesGrouperFlush(t *testing.T) {
	tests := []struct {
		name           string
		initialMap     map[int64]agent.GroupedUpdates
		expectedSends  []agent.UpdateToSend
		handlerReturns error
	}{
		{
			name: "flush single group",
			initialMap: map[int64]agent.GroupedUpdates{
				111: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "critical update",
							TgChatIDs:   []int64{111},
							Priority:    agent.PriorityHigh,
						},
					},
					Priority: agent.PriorityHigh,
				},
			},
			expectedSends: []agent.UpdateToSend{
				{
					TgChatID: 111,
					GroupedUpdates: agent.GroupedUpdates{
						Updates: []agent.UpdateToGroup{
							{
								Description: "critical update",
								TgChatIDs:   []int64{111},
								Priority:    agent.PriorityHigh,
							},
						},
						Priority: agent.PriorityHigh,
					},
				},
			},
			handlerReturns: nil,
		},
		{
			name: "flush multiple groups",
			initialMap: map[int64]agent.GroupedUpdates{
				111: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "docs update",
							TgChatIDs:   []int64{111},
							Priority:    agent.PriorityLow,
						},
					},
					Priority: agent.PriorityLow,
				},
				222: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "breaking update",
							TgChatIDs:   []int64{222},
							Priority:    agent.PriorityHigh,
						},
					},
					Priority: agent.PriorityHigh,
				},
			},
			expectedSends: []agent.UpdateToSend{
				{
					TgChatID: 111,
					GroupedUpdates: agent.GroupedUpdates{
						Updates: []agent.UpdateToGroup{
							{
								Description: "docs update",
								TgChatIDs:   []int64{111},
								Priority:    agent.PriorityLow,
							},
						},
						Priority: agent.PriorityLow,
					},
				},
				{
					TgChatID: 222,
					GroupedUpdates: agent.GroupedUpdates{
						Updates: []agent.UpdateToGroup{
							{
								Description: "breaking update",
								TgChatIDs:   []int64{222},
								Priority:    agent.PriorityHigh,
							},
						},
						Priority: agent.PriorityHigh,
					},
				},
			},
			handlerReturns: nil,
		},
		{
			name: "flush clears map even if handler returns error",
			initialMap: map[int64]agent.GroupedUpdates{
				111: {
					Updates: []agent.UpdateToGroup{
						{
							Description: "critical update",
							TgChatIDs:   []int64{111},
							Priority:    agent.PriorityHigh,
						},
					},
					Priority: agent.PriorityHigh,
				},
			},
			expectedSends: []agent.UpdateToSend{
				{
					TgChatID: 111,
					GroupedUpdates: agent.GroupedUpdates{
						Updates: []agent.UpdateToGroup{
							{
								Description: "critical update",
								TgChatIDs:   []int64{111},
								Priority:    agent.PriorityHigh,
							},
						},
						Priority: agent.PriorityHigh,
					},
				},
			},
			handlerReturns: errors.New("send error"),
		},
		{
			name:           "flush empty map",
			initialMap:     map[int64]agent.GroupedUpdates{},
			expectedSends:  nil,
			handlerReturns: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := mocks.NewMockUpdatesHandler(ctrl)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			grouper := NewUpdatesGrouper(logger, handler)
			grouper.TgChatsMap = tt.initialMap

			for _, expected := range tt.expectedSends {
				handler.
					EXPECT().
					FormAndSendUpdate(expected).
					Return(tt.handlerReturns)
			}

			grouper.Flush()

			assert.Empty(t, grouper.TgChatsMap)
		})
	}
}
