package scrapper

import "strings"

type LinkType int

const (
	GithubLinkType LinkType = iota
	StackOverflowLinkType
	UnknownLinkType
)

func GetLinkType(link string) LinkType {
	switch {
	case strings.Contains(link, "github"):
		return GithubLinkType
	case strings.Contains(link, "stackoverflow"):
		return StackOverflowLinkType
	default:
		return UnknownLinkType
	}
}
