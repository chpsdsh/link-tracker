package scrapper

import "errors"

var (
	ErrChatNotFound               = errors.New("chat not found")
	ErrLinkExists                 = errors.New("link already tracked")
	ErrIncorrectRequestParameters = errors.New("incorrect request parameters")
	ErrChatAlreadyExists          = errors.New("chat already exists")
	ErrLinkNotExists              = errors.New("link not exists")
	ErrInternalError              = errors.New("internal server error")
	ErrNotGitHubURL               = errors.New("not github")
	ErrInvalidGitHubURL           = errors.New("invalid github url")
	ErrNotStackOverflow           = errors.New("not StackOverflow")
	ErrInvalidStackOverflowURL    = errors.New("invalid StackOverflow url")
	ErrUnsupportedGithubURL       = errors.New("unsupported github url")
	ErrNotURL                     = errors.New("not url")
)
