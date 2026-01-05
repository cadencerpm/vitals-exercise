package app

import (
	"context"
	"errors"
	"testing"
)

func TestInMemoryStoreHonorsContext(t *testing.T) {
	store := NewInMemoryStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := store.AddVital(ctx, Vital{PatientID: "patient-1"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error for add vital, got %v", err)
	}
	if _, err := store.AddAlert(ctx, Alert{PatientID: "patient-1"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error for add alert, got %v", err)
	}
	if _, err := store.ListVitals(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error for list vitals, got %v", err)
	}
	if _, err := store.ListAlerts(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error for list alerts, got %v", err)
	}

	vitals, err := store.ListVitals(context.Background())
	if err != nil {
		t.Fatalf("list vitals failed: %v", err)
	}
	if len(vitals) != 0 {
		t.Fatalf("expected no vitals stored, got %d", len(vitals))
	}

	alerts, err := store.ListAlerts(context.Background())
	if err != nil {
		t.Fatalf("list alerts failed: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts stored, got %d", len(alerts))
	}
}
