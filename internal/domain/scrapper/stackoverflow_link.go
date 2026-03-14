package scrapper

import "fmt"

type StackOverflowLinkType int

const StackOverflowQuestion StackOverflowLinkType = iota

type StackOverflowLink struct {
	Type StackOverflowLinkType
	ID   string
}

func (s StackOverflowLink) ConvertToURL() string {
	var url string

	if s.Type == StackOverflowQuestion {
		url = fmt.Sprintf(
			"https://api.stackexchange.com/2.3/questions/%s?site=stackoverflow",
			s.ID,
		)
	}

	return url
}
