package scrapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStackOverflowLink_ConvertToURL(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		option   StackOverflowLinkOption
		expected string
	}{
		{
			name:     "question url",
			id:       "123",
			option:   StackOverflowLinkQuestion,
			expected: "https://api.stackexchange.com/2.3/questions/123?site=stackoverflow&filter=withbody",
		},
		{
			name:     "answers url",
			id:       "123",
			option:   StackOverflowLinkAnswer,
			expected: "https://api.stackexchange.com/2.3/questions/123/answers?site=stackoverflow&filter=withbody",
		},
		{
			name:     "comments url",
			id:       "123",
			option:   StackOverflowLinkComment,
			expected: "https://api.stackexchange.com/2.3/questions/123/comments?site=stackoverflow&filter=withbody",
		},
		{
			name:     "unknown option",
			id:       "123",
			option:   StackOverflowLinkOption(999),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := StackOverflowLink{ID: tt.id}

			result := link.ConvertToURL(tt.option)

			assert.Equal(t, tt.expected, result)
		})
	}
}
