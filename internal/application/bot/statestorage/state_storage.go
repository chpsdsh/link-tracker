package statestorage

import "sync"

type State int

type CurrentState struct {
	State       State
	LinkToTrack string
}

const (
	NoState State = iota
	WaitingForTrackURLState
	WaitingForUntrackURLState
	WaitingForTagsState
	InitialState
)

type StateStorage struct {
	stateMap   map[int64]CurrentState
	stateMutex sync.RWMutex
}

func NewStateStorage() *StateStorage {
	return &StateStorage{stateMap: make(map[int64]CurrentState), stateMutex: sync.RWMutex{}}
}

func (s *StateStorage) GetState(chatID int64) State {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.stateMap[chatID].State
}

func (s *StateStorage) SetState(chatID int64, state State) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	currentState := s.stateMap[chatID]
	currentState.State = state
	s.stateMap[chatID] = currentState
}

func (s *StateStorage) SetLinkAndUpdateState(chatID int64, link string, state State) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	currentState := s.stateMap[chatID]
	currentState.LinkToTrack = link
	currentState.State = state
	s.stateMap[chatID] = currentState
}

func (s *StateStorage) GetLink(chatID int64) string {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.stateMap[chatID].LinkToTrack
}

func (s *StateStorage) ClearLinkAndUpdateState(chatID int64, state State) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	currentState := s.stateMap[chatID]
	currentState.LinkToTrack = ""
	currentState.State = state
	s.stateMap[chatID] = currentState
}
