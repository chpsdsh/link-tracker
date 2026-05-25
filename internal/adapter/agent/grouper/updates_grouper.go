//go:generate mockgen -source updates_grouper.go -destination=../mocks/updates_grouper_mocks.go -package=mocks
package grouper

import (
	"log/slog"
	"sync"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/agent"
)

type UpdatesHandler interface {
	FormAndSendUpdate(update agent.UpdateToSend) error
}

type UpdatesGrouper struct {
	TgChatsMap     map[int64]agent.GroupedUpdates
	mu             sync.Mutex
	UpdatesHandler UpdatesHandler
	BaseLogger     *slog.Logger
}

func NewUpdatesGrouper(baseLogger *slog.Logger, updatesHandler UpdatesHandler) *UpdatesGrouper {
	return &UpdatesGrouper{
		TgChatsMap:     make(map[int64]agent.GroupedUpdates),
		UpdatesHandler: updatesHandler,
		BaseLogger:     baseLogger,
		mu:             sync.Mutex{},
	}
}

func (u *UpdatesGrouper) Add(update agent.UpdateToGroup) {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, id := range update.TgChatIDs {
		group, ok := u.TgChatsMap[id]
		if !ok {
			u.TgChatsMap[id] = agent.GroupedUpdates{Updates: []agent.UpdateToGroup{update}, Priority: update.Priority}
			continue
		}

		group.Updates = append(group.Updates, update)
		if isHigherPriority(u.TgChatsMap[id].Priority, update.Priority) {
			group.Priority = update.Priority
		}

		u.TgChatsMap[id] = group
	}
}

func (u *UpdatesGrouper) Flush() {
	u.mu.Lock()
	defer u.mu.Unlock()
	for id, group := range u.TgChatsMap {
		if err := u.UpdatesHandler.FormAndSendUpdate(agent.UpdateToSend{TgChatID: id, GroupedUpdates: group}); err != nil {
			u.BaseLogger.Error("forming and sending update", slog.String("error", err.Error()))
		}
		delete(u.TgChatsMap, id)
	}
}

func isHigherPriority(currentPriority, priorityToCompare agent.Priority) bool {
	return priorityToCompare > currentPriority
}
