package shared

type LinkUpdate struct {
	Description string  `json:"description,omitempty"`
	Id          int64   `json:"id,omitempty"`
	TgChatIds   []int64 `json:"tgChatIds,omitempty"`
	Url         string  `json:"url,omitempty"`
}
