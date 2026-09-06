package ingestion

import (
	"context"
	"sync"
)

type FakeAdapter struct {
	AdapterName string
	Records     []Record
	Failures    int

	mu     sync.Mutex
	next   int
	opened bool
}

func (a *FakeAdapter) Name() string {
	if a.AdapterName == "" {
		return "fake"
	}
	return a.AdapterName
}

func (a *FakeAdapter) Open(_ context.Context, checkpoint Checkpoint) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.opened = true
	a.next = 0
	if checkpoint.Value == "" {
		return nil
	}
	for i, record := range a.Records {
		if record.SourceSequence == checkpoint.Value {
			a.next = i + 1
			return nil
		}
	}
	return nil
}

func (a *FakeAdapter) Next(context.Context) (Record, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Failures > 0 {
		a.Failures--
		return Record{}, ErrInvalidRecord
	}
	if !a.opened || a.next >= len(a.Records) {
		return Record{}, EndOfSource
	}
	record := a.Records[a.next]
	a.next++
	return record, nil
}

func (a *FakeAdapter) Close(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.opened = false
	return nil
}

type MemoryCheckpointStore struct {
	mu    sync.RWMutex
	items map[Checkpoint]Checkpoint
}

func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{items: make(map[Checkpoint]Checkpoint)}
}

func (s *MemoryCheckpointStore) Load(_ context.Context, adapter string, mode Mode) (Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := Checkpoint{Adapter: adapter, Mode: mode}
	cp, ok := s.items[key]
	if !ok {
		return Checkpoint{}, ErrCheckpointNotFound
	}
	return cp, nil
}

func (s *MemoryCheckpointStore) Save(_ context.Context, checkpoint Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[Checkpoint{Adapter: checkpoint.Adapter, Mode: checkpoint.Mode}] = checkpoint
	return nil
}
