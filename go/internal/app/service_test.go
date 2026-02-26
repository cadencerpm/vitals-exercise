package app

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestServiceIngestPublishesEvent(t *testing.T) {
	store := NewInMemoryStore()
	pubsub := NewPubSub()
	service := NewService(store, pubsub)

	events, cancel := pubsub.Subscribe(1)
	defer cancel()

	ctx := context.Background()
	takenAt := time.Now().Add(-2 * time.Minute)
	stored, err := service.IngestVital(ctx, "patient-1", 120, 80, takenAt)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if stored.ID == 0 {
		t.Fatal("expected stored vital to have an ID")
	}

	vitals, err := store.ListVitals(ctx)
	if err != nil {
		t.Fatalf("list vitals failed: %v", err)
	}
	if len(vitals) != 1 {
		t.Fatalf("expected 1 vital, got %d", len(vitals))
	}
	if vitals[0].PatientID != "patient-1" {
		t.Fatalf("unexpected patient id: %s", vitals[0].PatientID)
	}

	select {
	case event := <-events:
		if event.Type != EventTypeVitalReceived {
			t.Fatalf("unexpected event type: %s", event.Type)
		}
		if event.Vital.ID != stored.ID {
			t.Fatalf("unexpected vital id: %d", event.Vital.ID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}

func TestServiceIngestValidatesInput(t *testing.T) {
	store := NewInMemoryStore()
	pubsub := NewPubSub()
	service := NewService(store, pubsub)

	ctx := context.Background()
	if _, err := service.IngestVital(ctx, "", 120, 80, time.Now()); err == nil {
		t.Fatal("expected error for missing patient id")
	}
	if _, err := service.IngestVital(ctx, "patient-1", -1, 80, time.Now()); err == nil {
		t.Fatal("expected error for invalid systolic")
	}
	if _, err := service.IngestVital(ctx, "patient-1", 120, 0, time.Now()); err == nil {
		t.Fatal("expected error for invalid diastolic")
	}
	if _, err := service.IngestVital(ctx, "patient-1", 120, 80, time.Time{}); err == nil {
		t.Fatal("expected error for missing taken_at")
	}
}

func TestServiceListAlertsFiltersByPatient(t *testing.T) {
	store := NewInMemoryStore()
	pubsub := NewPubSub()
	service := NewService(store, pubsub)

	ctx := context.Background()
	if _, err := store.AddAlert(ctx, Alert{PatientID: "patient-1", Status: AlertStatusActive}); err != nil {
		t.Fatalf("add alert: %v", err)
	}
	if _, err := store.AddAlert(ctx, Alert{PatientID: "patient-2", Status: AlertStatusActive}); err != nil {
		t.Fatalf("add alert: %v", err)
	}

	alerts, err := service.ListAlerts(ctx, "patient-1")
	if err != nil {
		t.Fatalf("list alerts failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].PatientID != "patient-1" {
		t.Fatalf("unexpected patient id: %s", alerts[0].PatientID)
	}
}

type flakyPublisher struct {
	calls int32
}

func (p *flakyPublisher) Publish(ctx context.Context, _ Event) error {
	if atomic.AddInt32(&p.calls, 1) == 1 {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

func TestServiceIngestRetryShouldBeIdempotentAfterPublishTimeout(t *testing.T) {
	store := NewInMemoryStore()
	pub := &flakyPublisher{}
	service := NewService(store, pub)

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer timeoutCancel()

	if _, err := service.IngestVital(timeoutCtx, "patient-1", 120, 80, time.Now().UTC()); err == nil {
		t.Fatal("expected ingest to fail due to publish timeout")
	} else if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded error, got %v", err)
	}

	if _, err := service.IngestVital(context.Background(), "patient-1", 120, 80, time.Now().UTC()); err != nil {
		t.Fatalf("retry ingest failed: %v", err)
	}

	vitals, err := store.ListVitals(context.Background())
	if err != nil {
		t.Fatalf("list vitals failed: %v", err)
	}
	if len(vitals) != 1 {
		t.Fatalf("expected 1 vital after retry, got %d", len(vitals))
	}
}
