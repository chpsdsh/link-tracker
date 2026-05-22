package pkg

type ProcessedLinkUpdate struct {
	Description string  `json:"description,omitempty"`
	ID          int64   `json:"id,omitempty"`
	TgChatIDs   []int64 `json:"tgChatIds,omitempty"`
	Priority    string  `json:"priority,omitempty"`
}
