package shared

type AddLinkRequest struct {
	Filters *[]string `json:"filters,omitempty"`
	Link    *string   `json:"link,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
}
