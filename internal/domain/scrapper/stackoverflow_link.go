package scrapper

import "fmt"

type StackOverflowLinkOption int

const (
	StackOverflowLinkQuestion StackOverflowLinkOption = iota
	StackOverflowLinkAnswer
	StackOverflowLinkComment
)

type StackOverflowLink struct {
	ID string
}

func (s StackOverflowLink) ConvertToURL(option StackOverflowLinkOption) string {
	var url string
	switch option {
	case StackOverflowLinkQuestion:
		url = fmt.Sprintf(
			"https://api.stackexchange.com/2.3/questions/%s?site=stackoverflow&filter=withbody",
			s.ID,
		)
	case StackOverflowLinkAnswer:
		url = fmt.Sprintf(
			"https://api.stackexchange.com/2.3/questions/%s/answers?site=stackoverflow&filter=withbody",
			s.ID,
		)
	case StackOverflowLinkComment:
		url = fmt.Sprintf(
			"https://api.stackexchange.com/2.3/questions/%s/comments?site=stackoverflow&filter=withbody",
			s.ID,
		)
	}
	return url
}
