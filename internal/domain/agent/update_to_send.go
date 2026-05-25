package agent

type UpdateToSend struct {
	TgChatID       int64
	GroupedUpdates GroupedUpdates
}
