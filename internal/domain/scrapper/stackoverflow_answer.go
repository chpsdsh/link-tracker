package scrapper

type StackOverflowAnswersResponse struct {
	Items []StackOverflowAnswer `json:"items"`
}

type StackOverflowAnswer struct {
	LastActivityDate int64  `json:"last_activity_date"`
	CreationDate     int64  `json:"creation_date"`
	Body             string `json:"body"`
	Owner            struct {
		DisplayName string `json:"display_name"`
	} `json:"owner"`
}
