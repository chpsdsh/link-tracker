package repository

import (
	"reflect"
	"testing"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain/shared"
)

func TestLinkRepositoryChatExists(t *testing.T) {
	tests := []struct {
		name     string
		chatID   int64
		prepare  func(r *LinkRepository)
		expected bool
	}{
		{
			name:     "chat exists",
			chatID:   1,
			prepare:  func(r *LinkRepository) { r.AddChat(1) },
			expected: true,
		},
		{
			name:     "chat does not exist",
			chatID:   2,
			prepare:  func(r *LinkRepository) {},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()
			tt.prepare(r)

			result := r.ChatExists(tt.chatID)

			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLinkRepositoryAddChat(t *testing.T) {
	tests := []struct {
		name   string
		chatID int64
	}{
		{
			name:   "add first chat",
			chatID: 1,
		},
		{
			name:   "add another chat",
			chatID: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()

			r.AddChat(tt.chatID)

			if !r.ChatExists(tt.chatID) {
				t.Fatalf("chat %d should exist", tt.chatID)
			}

			links := r.GetLinks(tt.chatID)
			if len(links) != 0 {
				t.Fatalf("expected empty links slice, got %v", links)
			}
		})
	}
}

func TestLinkRepositoryGetLinks(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		chatID   int64
		prepare  func(r *LinkRepository)
		expected []shared.LinkInfo
	}{
		{
			name:   "get empty links",
			chatID: 1,
			prepare: func(r *LinkRepository) {
				r.AddChat(1)
			},
			expected: []shared.LinkInfo{},
		},
		{
			name:   "get links",
			chatID: 2,
			prepare: func(r *LinkRepository) {
				r.AddChat(2)
				r.AddLink(2, shared.LinkInfo{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: now})
				r.AddLink(2, shared.LinkInfo{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}, LastUpdateTime: now})
			},
			expected: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: now},
				{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}, LastUpdateTime: now},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()
			tt.prepare(r)

			result := r.GetLinks(tt.chatID)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLinkRepositoryAddLink(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		chatID   int64
		link     shared.LinkInfo
		expected []shared.LinkInfo
	}{
		{
			name:   "add one link",
			chatID: 1,
			link:   shared.LinkInfo{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: now},
			expected: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: now},
			},
		},
		{
			name:   "add second link",
			chatID: 2,
			link:   shared.LinkInfo{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}, LastUpdateTime: now},
			expected: []shared.LinkInfo{
				{Link: "https://github.com/golang/go", Tags: []string{"repo"}, LastUpdateTime: now},
				{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}, LastUpdateTime: now},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()
			r.AddChat(tt.chatID)

			if tt.name == "add second link" {
				r.AddLink(tt.chatID, shared.LinkInfo{Link: "https://github.com/golang/go", Tags: []string{"repo"}, LastUpdateTime: now})
			}

			r.AddLink(tt.chatID, tt.link)

			result := r.GetLinks(tt.chatID)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLinkRepositoryDeleteLink(t *testing.T) {
	now := time.Now()

	first := shared.LinkInfo{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: now}
	second := shared.LinkInfo{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}, LastUpdateTime: now}

	tests := []struct {
		name         string
		linkToDelete string
		expectedLink shared.LinkInfo
		expectedOK   bool
		expectedRest []shared.LinkInfo
	}{
		{
			name:         "delete existing link",
			linkToDelete: first.Link,
			expectedLink: first,
			expectedOK:   true,
			expectedRest: []shared.LinkInfo{second},
		},
		{
			name:         "delete non existing link",
			linkToDelete: "https://example.com",
			expectedLink: shared.LinkInfo{},
			expectedOK:   false,
			expectedRest: []shared.LinkInfo{first, second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()
			r.AddChat(1)
			r.AddLink(1, first)
			r.AddLink(1, second)

			link, ok := r.DeleteLink(1, tt.linkToDelete)

			if ok != tt.expectedOK {
				t.Fatalf("expected ok %v, got %v", tt.expectedOK, ok)
			}

			if !reflect.DeepEqual(link, tt.expectedLink) {
				t.Fatalf("expected link %v, got %v", tt.expectedLink, link)
			}

			rest := r.GetLinks(1)
			if !reflect.DeepEqual(rest, tt.expectedRest) {
				t.Fatalf("expected remaining links %v, got %v", tt.expectedRest, rest)
			}
		})
	}
}

func TestLinkRepositoryDeleteChat(t *testing.T) {
	tests := []struct {
		name   string
		chatID int64
	}{
		{
			name:   "delete existing chat",
			chatID: 1,
		},
		{
			name:   "delete another chat",
			chatID: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()
			r.AddChat(tt.chatID)

			r.DeleteChat(tt.chatID)

			if r.ChatExists(tt.chatID) {
				t.Fatalf("chat %d should not exist", tt.chatID)
			}
		})
	}
}

func TestLinkRepositoryGetAllLinks(t *testing.T) {
	now := time.Now()

	link1 := shared.LinkInfo{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: now}
	link2 := shared.LinkInfo{Link: "https://stackoverflow.com/questions/1/test", Tags: []string{"study"}, LastUpdateTime: now}

	tests := []struct {
		name     string
		prepare  func(r *LinkRepository)
		expected []shared.LinkInfo
	}{
		{
			name: "no links",
			prepare: func(r *LinkRepository) {
				r.AddChat(1)
			},
			expected: []shared.LinkInfo{},
		},
		{
			name: "links from multiple chats",
			prepare: func(r *LinkRepository) {
				r.AddChat(1)
				r.AddChat(2)
				r.AddLink(1, link1)
				r.AddLink(2, link2)
			},
			expected: []shared.LinkInfo{link1, link2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()
			tt.prepare(r)

			result := r.GetAllLinks()

			if len(result) != len(tt.expected) {
				t.Fatalf("expected len %d, got %d", len(tt.expected), len(result))
			}

			for _, expectedLink := range tt.expected {
				found := false
				for _, resultLink := range result {
					if reflect.DeepEqual(expectedLink, resultLink) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected link %v not found in result %v", expectedLink, result)
				}
			}
		})
	}
}

func TestLinkRepositoryGetChatIdsByLink(t *testing.T) {
	now := time.Now()

	target := "https://github.com/golang/go"

	tests := []struct {
		name     string
		prepare  func(r *LinkRepository)
		expected []int64
	}{
		{
			name: "link not found",
			prepare: func(r *LinkRepository) {
				r.AddChat(1)
				r.AddLink(1, shared.LinkInfo{Link: "https://example.com", LastUpdateTime: now})
			},
			expected: []int64{},
		},
		{
			name: "link found in several chats",
			prepare: func(r *LinkRepository) {
				r.AddChat(1)
				r.AddChat(2)
				r.AddChat(3)
				r.AddLink(1, shared.LinkInfo{Link: target, LastUpdateTime: now})
				r.AddLink(2, shared.LinkInfo{Link: "https://example.com", LastUpdateTime: now})
				r.AddLink(3, shared.LinkInfo{Link: target, LastUpdateTime: now})
			},
			expected: []int64{1, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()
			tt.prepare(r)

			result := r.GetChatIdsByLink(target)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLinkRepositoryUpdateLinksTime(t *testing.T) {
	oldTime := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(5 * time.Minute)

	tests := []struct {
		name        string
		linkToTrack string
		prepare     func(r *LinkRepository)
		expected    map[int64][]shared.LinkInfo
	}{
		{
			name:        "update matching links in multiple chats",
			linkToTrack: "https://github.com/golang/go",
			prepare: func(r *LinkRepository) {
				r.AddChat(1)
				r.AddChat(2)
				r.AddLink(1, shared.LinkInfo{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: oldTime})
				r.AddLink(1, shared.LinkInfo{Link: "https://example.com", Tags: []string{"other"}, LastUpdateTime: oldTime})
				r.AddLink(2, shared.LinkInfo{Link: "https://github.com/golang/go", Tags: []string{"repo"}, LastUpdateTime: oldTime})
			},
			expected: map[int64][]shared.LinkInfo{
				1: {
					{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: newTime},
					{Link: "https://example.com", Tags: []string{"other"}, LastUpdateTime: oldTime},
				},
				2: {
					{Link: "https://github.com/golang/go", Tags: []string{"repo"}, LastUpdateTime: newTime},
				},
			},
		},
		{
			name:        "do nothing for unknown link",
			linkToTrack: "https://unknown.com",
			prepare: func(r *LinkRepository) {
				r.AddChat(1)
				r.AddLink(1, shared.LinkInfo{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: oldTime})
			},
			expected: map[int64][]shared.LinkInfo{
				1: {
					{Link: "https://github.com/golang/go", Tags: []string{"work"}, LastUpdateTime: oldTime},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewLinkRepository()
			tt.prepare(r)

			r.UpdateLinksTime(newTime, shared.LinkInfo{Link: tt.linkToTrack})

			for chatID, expectedLinks := range tt.expected {
				result := r.GetLinks(chatID)
				if !reflect.DeepEqual(result, expectedLinks) {
					t.Fatalf("chat %d: expected %v, got %v", chatID, expectedLinks, result)
				}
			}
		})
	}
}
