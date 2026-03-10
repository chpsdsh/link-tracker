package statestorage

import "sync"

type State int

type CurrentState struct {
	State       State
	LinkToTrack string
}

const (
	NoState State = iota
	WaitingForTrackUrlState
	WaitingForUnTrackUrlState
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

func (s *StateStorage) GetState(chatId int64) State {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.stateMap[chatId].State
}

func (s *StateStorage) SetState(chatId int64, state State) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	currentState := s.stateMap[chatId]
	currentState.State = state
	s.stateMap[chatId] = currentState
}

func (s *StateStorage) SetLinkAndUpdateState(chatId int64, link string, state State) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	currentState := s.stateMap[chatId]
	currentState.LinkToTrack = link
	currentState.State = state
	s.stateMap[chatId] = currentState
}

func (s *StateStorage) GetLink(chatId int64) string {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.stateMap[chatId].LinkToTrack
}

func (s *StateStorage) ClearLinkAndUpdateState(chatId int64, state State) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	currentState := s.stateMap[chatId]
	currentState.LinkToTrack = ""
	currentState.State = state
	s.stateMap[chatId] = currentState
}
