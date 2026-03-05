package statestorage

import "sync"

type State int

const (
	NoState State = iota
	WaitingForTrackUrlState
	WaitingForUnTrackUrlState
	WaitingForTagsState
	InitialState
)

type StateStorage struct {
	stateMap   map[int64]State
	stateMutex sync.RWMutex
}

func NewStateStorage() *StateStorage {
	return &StateStorage{stateMap: make(map[int64]State), stateMutex: sync.RWMutex{}}
}

func (s *StateStorage) GetState(chatId int64) State {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.stateMap[chatId]
}

func (s *StateStorage) SetState(chatId int64, state State) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	s.stateMap[chatId] = state
}
