package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	vitalsv1 "cadence-vitals-interview/proto/vitals/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "insert-vital":
		insertVitalCmd(os.Args[2:])
	case "list-alerts":
		listAlertsCmd(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func insertVitalCmd(args []string) {
	fs := flag.NewFlagSet("insert-vital", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:50051", "gRPC server address")
	patientID := fs.String("patient", "", "patient identifier")
	systolic := fs.Int("systolic", 0, "systolic blood pressure")
	diastolic := fs.Int("diastolic", 0, "diastolic blood pressure")
	takenAt := fs.Int64("taken-at", 0, "unix timestamp when blood pressure was taken")
	fs.Parse(args)

	client, cleanup := newClient(*addr)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if *takenAt == 0 {
		*takenAt = time.Now().Unix()
	}

	resp, err := client.IngestVital(ctx, &vitalsv1.IngestVitalRequest{
		PatientId: *patientID,
		Systolic:  int32(*systolic),
		Diastolic: int32(*diastolic),
		TakenAt:   *takenAt,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert vital failed: %v\n", err)
		os.Exit(1)
	}

	vital := resp.GetVital()
	fmt.Printf("stored vital id=%d patient=%s bp=%d/%d taken_at=%d received_at=%d\n", vital.GetId(), vital.GetPatientId(), vital.GetSystolic(), vital.GetDiastolic(), vital.GetTakenAt(), vital.GetReceivedAt())
}

func listAlertsCmd(args []string) {
	fs := flag.NewFlagSet("list-alerts", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:50051", "gRPC server address")
	patientID := fs.String("patient", "", "patient identifier")
	fs.Parse(args)

	client, cleanup := newClient(*addr)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ListAlerts(ctx, &vitalsv1.ListAlertsRequest{PatientId: *patientID})
	if err != nil {
		fmt.Fprintf(os.Stderr, "list alerts failed: %v\n", err)
		os.Exit(1)
	}

	if len(resp.GetAlerts()) == 0 {
		fmt.Println("no alerts")
		return
	}

	for _, alert := range resp.GetAlerts() {
		vital := alert.GetVital()
		fmt.Printf("alert id=%d patient=%s bp=%d/%d status=%s reason=%s created_at=%d\n", alert.GetId(), vital.GetPatientId(), vital.GetSystolic(), vital.GetDiastolic(), alert.GetStatus().String(), alert.GetReason(), alert.GetCreatedAt())
	}
}

func newClient(addr string) (vitalsv1.VitalsServiceClient, func()) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial %s: %v\n", addr, err)
		os.Exit(1)
	}
	cleanup := func() { _ = conn.Close() }
	return vitalsv1.NewVitalsServiceClient(conn), cleanup
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  cli insert-vital --patient <id> --systolic <value> --diastolic <value> [--taken-at <unix>] [--addr host:port]")
	fmt.Fprintln(os.Stderr, "  cli list-alerts [--patient <id>] [--addr host:port]")
}
