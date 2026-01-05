package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidVital = errors.New("invalid vital")

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

type Service struct {
	store Store
	pub   Publisher
}

func NewService(store Store, pub Publisher) *Service {
	return &Service{store: store, pub: pub}
}

func (s *Service) IngestVital(ctx context.Context, patientID string, systolic, diastolic int32, takenAt time.Time) (Vital, error) {
	patientID = strings.TrimSpace(patientID)
	if patientID == "" {
		return Vital{}, fmt.Errorf("%w: patient_id is required", ErrInvalidVital)
	}
	if systolic <= 0 || diastolic <= 0 {
		return Vital{}, fmt.Errorf("%w: systolic and diastolic must be positive", ErrInvalidVital)
	}
	if takenAt.IsZero() {
		return Vital{}, fmt.Errorf("%w: taken_at is required", ErrInvalidVital)
	}

	vital := Vital{
		PatientID:  patientID,
		Systolic:   systolic,
		Diastolic:  diastolic,
		TakenAt:    takenAt.UTC(),
		ReceivedAt: time.Now().UTC(),
	}

	stored, err := s.store.AddVital(ctx, vital)
	if err != nil {
		return Vital{}, err
	}

	event := Event{
		Type:  EventTypeVitalReceived,
		Vital: stored,
	}

	if err := s.pub.Publish(ctx, event); err != nil {
		return stored, err
	}

	return stored, nil
}

func (s *Service) ListAlerts(ctx context.Context, patientID string) ([]Alert, error) {
	alerts, err := s.store.ListAlerts(ctx)
	if err != nil {
		return nil, err
	}
	patientID = strings.TrimSpace(patientID)
	if patientID == "" {
		return alerts, nil
	}
	filtered := make([]Alert, 0, len(alerts))
	for _, alert := range alerts {
		if alert.PatientID == patientID {
			filtered = append(filtered, alert)
		}
	}
	return filtered, nil
}
