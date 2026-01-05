package app

import (
	"fmt"
	"time"
)

const (
	MaxSystolic  = 180
	MaxDiastolic = 120
)

type EventType string

const EventTypeVitalReceived EventType = "VITAL_RECEIVED"

type Vital struct {
	ID         int64
	PatientID  string
	Systolic   int32
	Diastolic  int32
	TakenAt    time.Time
	ReceivedAt time.Time
}

type AlertStatus int32

const (
	AlertStatusActive       AlertStatus = 0
	AlertStatusResolved     AlertStatus = 1
	AlertStatusAutoResolved AlertStatus = 2
)

type Alert struct {
	ID         int64
	VitalID    int64
	PatientID  string
	Systolic   int32
	Diastolic  int32
	TakenAt    time.Time
	ReceivedAt time.Time
	Reason     string
	Status     AlertStatus
	Created    time.Time
}

type Event struct {
	Type  EventType
	Vital Vital
}

func IsAbnormal(vital Vital) bool {
	return vital.Systolic > MaxSystolic || vital.Diastolic > MaxDiastolic
}

func AlertReason(vital Vital) string {
	return fmt.Sprintf("abnormal blood pressure %d/%d", vital.Systolic, vital.Diastolic)
}
