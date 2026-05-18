package scrapper

type StackOverflowQuestionResponse struct {
	Items []StackOverflowQuestion `json:"items"`
}

type StackOverflowQuestion struct {
	LastActivityDate int64 `json:"last_activity_date"`
}
