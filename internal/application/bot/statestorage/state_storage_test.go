package statestorage

import "testing"

func TestStateStorageGetSetState(t *testing.T) {
	tests := []struct {
		name     string
		chatID   int64
		state    State
		expected State
	}{
		{
			name:     "set waiting track state",
			chatID:   1,
			state:    WaitingForTrackURLState,
			expected: WaitingForTrackURLState,
		},
		{
			name:     "set waiting untrack state",
			chatID:   2,
			state:    WaitingForUntrackURLState,
			expected: WaitingForUntrackURLState,
		},
		{
			name:     "set initial state",
			chatID:   3,
			state:    InitialState,
			expected: InitialState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStateStorage()

			s.SetState(tt.chatID, tt.state)
			result := s.GetState(tt.chatID)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestStateStorageSetLinkAndGetLink(t *testing.T) {
	tests := []struct {
		name     string
		chatID   int64
		link     string
		state    State
		expLink  string
		expState State
	}{
		{
			name:     "set link and waiting tags state",
			chatID:   1,
			link:     "https://github.com/golang/go",
			state:    WaitingForTagsState,
			expLink:  "https://github.com/golang/go",
			expState: WaitingForTagsState,
		},
		{
			name:     "set link and waiting track state",
			chatID:   2,
			link:     "https://stackoverflow.com/questions/123",
			state:    WaitingForTrackURLState,
			expLink:  "https://stackoverflow.com/questions/123",
			expState: WaitingForTrackURLState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStateStorage()

			s.SetLinkAndUpdateState(tt.chatID, tt.link, tt.state)

			link := s.GetLink(tt.chatID)
			state := s.GetState(tt.chatID)

			if link != tt.expLink {
				t.Errorf("expected link %s, got %s", tt.expLink, link)
			}

			if state != tt.expState {
				t.Errorf("expected state %v, got %v", tt.expState, state)
			}
		})
	}
}

func TestStateStorageClearLinkAndUpdateState(t *testing.T) {
	tests := []struct {
		name         string
		chatID       int64
		initialLink  string
		initialState State
		newState     State
	}{
		{
			name:         "clear link and set initial state",
			chatID:       1,
			initialLink:  "https://github.com/golang/go",
			initialState: WaitingForTagsState,
			newState:     InitialState,
		},
		{
			name:         "clear link and set waiting state",
			chatID:       2,
			initialLink:  "https://stackoverflow.com/questions/123",
			initialState: WaitingForTrackURLState,
			newState:     WaitingForUntrackURLState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStateStorage()

			s.SetLinkAndUpdateState(tt.chatID, tt.initialLink, tt.initialState)
			s.ClearLinkAndUpdateState(tt.chatID, tt.newState)

			link := s.GetLink(tt.chatID)
			state := s.GetState(tt.chatID)

			if link != "" {
				t.Errorf("expected empty link, got %s", link)
			}

			if state != tt.newState {
				t.Errorf("expected state %v, got %v", tt.newState, state)
			}
		})
	}
}

func TestStateStorageDefaultState(t *testing.T) {
	tests := []struct {
		name     string
		chatID   int64
		expected State
	}{
		{
			name:     "unknown chat returns no state",
			chatID:   1,
			expected: NoState,
		},
		{
			name:     "another unknown chat returns no state",
			chatID:   999,
			expected: NoState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStateStorage()

			state := s.GetState(tt.chatID)

			if state != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, state)
			}
		})
	}
}
