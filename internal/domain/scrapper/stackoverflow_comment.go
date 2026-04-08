package scrapper

type StackOverflowCommentsResponse struct {
	Items []StackOverflowComment `json:"items"`
}

type StackOverflowComment struct {
	CreationDate int64  `json:"creation_date"`
	Body         string `json:"body"`
	Owner        struct {
		DisplayName string `json:"display_name"`
	} `json:"owner"`
}
