package bot

type ApiErrorResponse struct {
	Code             *string   `json:"code,omitempty"`
	Description      *string   `json:"description,omitempty"`
	ExceptionMessage *string   `json:"exceptionMessage,omitempty"`
	ExceptionName    *string   `json:"exceptionName,omitempty"`
	Stacktrace       *[]string `json:"stacktrace,omitempty"`
}
