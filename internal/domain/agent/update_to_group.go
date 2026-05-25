package agent

type UpdateToGroup struct {
	Description string
	URL         string
	Priority    Priority
	TgChatIDs   []int64
}
