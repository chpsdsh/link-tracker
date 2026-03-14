package shared

type LinkUpdate struct {
	Description string  `json:"description,omitempty"`
	ID          int64   `json:"id,omitempty"`
	TgChatIDs   []int64 `json:"tgChatIds,omitempty"`
	URL         string  `json:"url,omitempty"`
}
