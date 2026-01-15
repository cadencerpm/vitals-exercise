package api

import (
	"context"
	"errors"
	"time"

	"cadence-vitals-interview/internal/app"
	vitalsv1 "cadence-vitals-interview/proto/vitals/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	vitalsv1.UnimplementedVitalsServiceServer
	service *app.Service
}

func NewServer(service *app.Service) *Server {
	return &Server{service: service}
}

func (s *Server) IngestVital(ctx context.Context, req *vitalsv1.IngestVitalRequest) (*vitalsv1.IngestVitalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	takenAt := req.GetTakenAt()
	if takenAt <= 0 {
		return nil, status.Error(codes.InvalidArgument, "taken_at is required")
	}
	vital, err := s.service.IngestVital(ctx, req.GetPatientId(), req.GetSystolic(), req.GetDiastolic(), time.Unix(takenAt, 0).UTC())
	if err != nil {
		if errors.Is(err, app.ErrInvalidVital) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &vitalsv1.IngestVitalResponse{Vital: toProtoVital(vital)}, nil
}

func (s *Server) ListAlerts(ctx context.Context, req *vitalsv1.ListAlertsRequest) (*vitalsv1.ListAlertsResponse, error) {
	var patientID string
	if req != nil {
		patientID = req.GetPatientId()
	}
	alerts, err := s.service.ListAlerts(ctx, patientID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &vitalsv1.ListAlertsResponse{
		Alerts: make([]*vitalsv1.Alert, 0, len(alerts)),
	}
	for _, alert := range alerts {
		resp.Alerts = append(resp.Alerts, toProtoAlert(alert))
	}
	return resp, nil
}

func (s *Server) ListVitals(ctx context.Context, req *vitalsv1.ListVitalsRequest) (*vitalsv1.ListVitalsResponse, error) {
	var patientID string
	if req != nil {
		patientID = req.GetPatientId()
	}
	vitals, err := s.service.ListVitals(ctx, patientID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &vitalsv1.ListVitalsResponse{
		Vitals: make([]*vitalsv1.Vital, 0, len(vitals)),
	}
	for _, vital := range vitals {
		resp.Vitals = append(resp.Vitals, toProtoVital(vital))
	}
	return resp, nil
}

func toProtoVital(vital app.Vital) *vitalsv1.Vital {
	return &vitalsv1.Vital{
		Id:         vital.ID,
		PatientId:  vital.PatientID,
		Systolic:   vital.Systolic,
		Diastolic:  vital.Diastolic,
		TakenAt:    vital.TakenAt.Unix(),
		ReceivedAt: vital.ReceivedAt.Unix(),
	}
}

func toProtoAlert(alert app.Alert) *vitalsv1.Alert {
	return &vitalsv1.Alert{
		Id: alert.ID,
		Vital: &vitalsv1.Vital{
			Id:         alert.VitalID,
			PatientId:  alert.PatientID,
			Systolic:   alert.Systolic,
			Diastolic:  alert.Diastolic,
			TakenAt:    alert.TakenAt.Unix(),
			ReceivedAt: alert.ReceivedAt.Unix(),
		},
		Reason:    alert.Reason,
		CreatedAt: alert.Created.Unix(),
		Status:    toProtoAlertStatus(alert.Status),
	}
}

func toProtoAlertStatus(status app.AlertStatus) vitalsv1.AlertStatus {
	switch status {
	case app.AlertStatusResolved:
		return vitalsv1.AlertStatus_ALERT_STATUS_RESOLVED
	case app.AlertStatusAutoResolved:
		return vitalsv1.AlertStatus_ALERT_STATUS_AUTO_RESOLVED
	default:
		return vitalsv1.AlertStatus_ALERT_STATUS_ACTIVE
	}
}
