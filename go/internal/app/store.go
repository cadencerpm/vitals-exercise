package app

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrStoreClosed = errors.New("store is closed")

type Store interface {
	AddVital(ctx context.Context, vital Vital) (Vital, error)
	AddAlert(ctx context.Context, alert Alert) (Alert, error)
	ListAlerts(ctx context.Context) ([]Alert, error)
	ListVitals(ctx context.Context) ([]Vital, error)
	Close()
}

type InMemoryStore struct {
	mu       sync.Mutex
	closed   bool
	vitalSeq int64
	alertSeq int64
	vitals   []Vital
	alerts   []Alert
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

func (s *InMemoryStore) AddVital(ctx context.Context, vital Vital) (Vital, error) {
	if err := ctx.Err(); err != nil {
		return Vital{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Vital{}, err
	}
	if s.closed {
		return Vital{}, ErrStoreClosed
	}
	if vital.ID == 0 {
		s.vitalSeq++
		vital.ID = s.vitalSeq
	}
	if vital.ReceivedAt.IsZero() {
		vital.ReceivedAt = time.Now().UTC()
	}
	s.vitals = append(s.vitals, vital)
	return vital, nil
}

func (s *InMemoryStore) AddAlert(ctx context.Context, alert Alert) (Alert, error) {
	if err := ctx.Err(); err != nil {
		return Alert{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Alert{}, err
	}
	if s.closed {
		return Alert{}, ErrStoreClosed
	}
	if alert.ID == 0 {
		s.alertSeq++
		alert.ID = s.alertSeq
	}
	if alert.Created.IsZero() {
		alert.Created = time.Now().UTC()
	}
	s.alerts = append(s.alerts, alert)
	return alert, nil
}

func (s *InMemoryStore) ListAlerts(ctx context.Context) ([]Alert, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.closed {
		return nil, ErrStoreClosed
	}
	alerts := make([]Alert, len(s.alerts))
	copy(alerts, s.alerts)
	return alerts, nil
}

func (s *InMemoryStore) ListVitals(ctx context.Context) ([]Vital, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.closed {
		return nil, ErrStoreClosed
	}
	vitals := make([]Vital, len(s.vitals))
	copy(vitals, s.vitals)
	return vitals, nil
}

func (s *InMemoryStore) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.vitals = nil
	s.alerts = nil
}
