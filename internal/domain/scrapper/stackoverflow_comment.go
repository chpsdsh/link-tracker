package scrapper

type StackOverflowCommentsResponse struct {
	Items []StackOverflowComment `json:"items"`
}

type StackOverflowComment struct {
	LastActivityDate int64  `json:"last_activity_date"`
	CreationDate     int64  `json:"creation_date"`
	Body             string `json:"body"`
	Owner            struct {
		DisplayName string `json:"display_name"`
	}
}
