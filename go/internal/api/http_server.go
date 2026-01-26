package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"cadence-vitals-interview/internal/app"
)

type HTTPServer struct {
	service      *app.Service
	messageQueue *app.MessageQueue

	mu         sync.RWMutex
	sseClients map[chan []byte]struct{}
}

func NewHTTPServer(service *app.Service, messageQueue *app.MessageQueue) *HTTPServer {
	s := &HTTPServer{
		service:      service,
		messageQueue: messageQueue,
		sseClients:   make(map[chan []byte]struct{}),
	}

	if messageQueue != nil {
		messageQueue.AddListener(s.onMessageUpdate)
	}

	return s
}

func (s *HTTPServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/vitals", s.handleVitals)
	mux.HandleFunc("/alerts", s.handleAlerts)
	mux.HandleFunc("/messages", s.handleMessages)
	mux.HandleFunc("/events", s.handleSSE)
	return mux
}

func (s *HTTPServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

func (s *HTTPServer) handleVitals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		patientID := r.URL.Query().Get("patient_id")
		vitals, err := s.service.ListVitals(r.Context(), patientID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := map[string]any{"vitals": vitalsToJSON(vitals)}
		json.NewEncoder(w).Encode(resp)

	case http.MethodPost:
		var req struct {
			PatientID string `json:"patient_id"`
			Systolic  int32  `json:"systolic"`
			Diastolic int32  `json:"diastolic"`
			TakenAt   int64  `json:"taken_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.TakenAt <= 0 {
			writeError(w, http.StatusBadRequest, "taken_at is required")
			return
		}
		vital, err := s.service.IngestVital(r.Context(), req.PatientID, req.Systolic, req.Diastolic, time.Unix(req.TakenAt, 0).UTC())
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		resp := map[string]any{"vital": vitalToJSON(vital)}
		json.NewEncoder(w).Encode(resp)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *HTTPServer) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	patientID := r.URL.Query().Get("patient_id")
	alerts, err := s.service.ListAlerts(r.Context(), patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := map[string]any{"alerts": alertsToJSON(alerts)}
	json.NewEncoder(w).Encode(resp)
}

func (s *HTTPServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if s.messageQueue == nil {
		json.NewEncoder(w).Encode(map[string]any{"messages": []any{}})
		return
	}
	messages := s.messageQueue.ListMessages()
	resp := map[string]any{"messages": messagesToJSON(messages)}
	json.NewEncoder(w).Encode(resp)
}

func (s *HTTPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan []byte, 100)
	s.mu.Lock()
	s.sseClients[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.sseClients, ch)
		s.mu.Unlock()
		close(ch)
	}()

	// Send connected message
	fmt.Fprintf(w, "data: {\"type\": \"connected\"}\n\n")
	flusher.Flush()

	// Keepalive ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (s *HTTPServer) onMessageUpdate(msg app.Message) {
	data, _ := json.Marshal(map[string]any{
		"type":    "message_update",
		"message": messageToJSON(msg),
	})

	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.sseClients {
		select {
		case ch <- data:
		default:
			// Channel full, skip
		}
	}
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func vitalToJSON(v app.Vital) map[string]any {
	return map[string]any{
		"id":          v.ID,
		"patient_id":  v.PatientID,
		"systolic":    v.Systolic,
		"diastolic":   v.Diastolic,
		"taken_at":    v.TakenAt.Unix(),
		"received_at": v.ReceivedAt.Unix(),
	}
}

func vitalsToJSON(vitals []app.Vital) []map[string]any {
	result := make([]map[string]any, len(vitals))
	for i, v := range vitals {
		result[i] = vitalToJSON(v)
	}
	return result
}

func alertToJSON(a app.Alert) map[string]any {
	return map[string]any{
		"id": a.ID,
		"vital": map[string]any{
			"id":          a.VitalID,
			"patient_id":  a.PatientID,
			"systolic":    a.Systolic,
			"diastolic":   a.Diastolic,
			"taken_at":    a.TakenAt.Unix(),
			"received_at": a.ReceivedAt.Unix(),
		},
		"reason":     a.Reason,
		"created_at": a.Created.Unix(),
		"status":     alertStatusString(a.Status),
	}
}

func alertsToJSON(alerts []app.Alert) []map[string]any {
	result := make([]map[string]any, len(alerts))
	for i, a := range alerts {
		result[i] = alertToJSON(a)
	}
	return result
}

func alertStatusString(s app.AlertStatus) string {
	switch s {
	case app.AlertStatusActive:
		return "ACTIVE"
	case app.AlertStatusAutoResolved:
		return "AUTO_RESOLVED"
	case app.AlertStatusResolvedByRetake:
		return "RESOLVED_BY_RETAKE"
	case app.AlertStatusConfirmedAbnormal:
		return "CONFIRMED_ABNORMAL"
	default:
		return "ACTIVE"
	}
}

func messageToJSON(m app.Message) map[string]any {
	result := map[string]any{
		"id":         m.ID,
		"patient_id": m.PatientID,
		"content":    m.Content,
		"status":     m.Status.String(),
		"queued_at":  m.QueuedAt.Unix(),
	}
	if !m.SentAt.IsZero() {
		result["sent_at"] = m.SentAt.Unix()
	} else {
		result["sent_at"] = 0
	}
	return result
}

func messagesToJSON(messages []app.Message) []map[string]any {
	result := make([]map[string]any, len(messages))
	for i, m := range messages {
		result[i] = messageToJSON(m)
	}
	return result
}

const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Vitals Monitor</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; padding: 20px; }
        h1 { margin-bottom: 20px; color: #333; }
        .container { max-width: 1200px; margin: 0 auto; }
        .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px; }
        .card { background: white; border-radius: 8px; padding: 16px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .card h2 { font-size: 16px; color: #666; margin-bottom: 12px; border-bottom: 1px solid #eee; padding-bottom: 8px; }
        .list { max-height: 300px; overflow-y: auto; }
        .item { padding: 8px; border-bottom: 1px solid #f0f0f0; font-size: 14px; }
        .item:last-child { border-bottom: none; }
        .item.normal { color: #2e7d32; }
        .item.abnormal { color: #c62828; }
        .status { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 500; }
        .status.QUEUED { background: #fff3e0; color: #e65100; }
        .status.PROCESSING { background: #e3f2fd; color: #1565c0; }
        .status.SENT { background: #e8f5e9; color: #2e7d32; }
        .full-width { grid-column: 1 / -1; }
        .quick-buttons { display: flex; gap: 10px; margin-bottom: 15px; }
        .btn { padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 500; }
        .btn-normal { background: #e8f5e9; color: #2e7d32; }
        .btn-normal:hover { background: #c8e6c9; }
        .btn-abnormal { background: #ffebee; color: #c62828; }
        .btn-abnormal:hover { background: #ffcdd2; }
        .btn-submit { background: #1976d2; color: white; }
        .btn-submit:hover { background: #1565c0; }
        .form-row { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
        .form-row label { font-size: 14px; color: #666; }
        .form-row input { padding: 8px; border: 1px solid #ddd; border-radius: 4px; width: 100px; }
        .form-row input[name="patient_id"] { width: 150px; }
        .empty { color: #999; font-style: italic; padding: 20px; text-align: center; }
        .time { color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Vitals Monitor Dashboard</h1>

        <div class="card full-width" style="margin-bottom: 20px;">
            <h2>Insert Vital</h2>
            <div class="quick-buttons">
                <button class="btn btn-normal" onclick="sendNormalVital()">Send Normal Vital (120/80)</button>
                <button class="btn btn-abnormal" onclick="sendAbnormalVital()">Send Abnormal Vital (190/130)</button>
            </div>
            <div class="form-row">
                <label>Patient:</label>
                <input type="text" name="patient_id" id="patient_id" value="patient-1" placeholder="patient-1">
                <label>Systolic:</label>
                <input type="number" name="systolic" id="systolic" value="120">
                <label>Diastolic:</label>
                <input type="number" name="diastolic" id="diastolic" value="80">
                <button class="btn btn-submit" onclick="sendCustomVital()">Submit</button>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h2>Recent Vitals</h2>
                <div class="list" id="vitals-list">
                    <div class="empty">No vitals yet</div>
                </div>
            </div>
            <div class="card">
                <h2>Active Alerts</h2>
                <div class="list" id="alerts-list">
                    <div class="empty">No alerts</div>
                </div>
            </div>
        </div>

        <div class="card full-width">
            <h2>Message Queue</h2>
            <div class="list" id="messages-list">
                <div class="empty">No messages queued</div>
            </div>
        </div>
    </div>

    <script>
        let currentPatientId = 'patient-1';

        function formatTime(unix) {
            if (!unix) return '';
            return new Date(unix * 1000).toLocaleTimeString();
        }

        function sendVital(patientId, systolic, diastolic) {
            currentPatientId = patientId;
            fetch('/vitals', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    patient_id: patientId,
                    systolic: systolic,
                    diastolic: diastolic,
                    taken_at: Math.floor(Date.now() / 1000)
                })
            }).then(() => refreshData());
        }

        function sendNormalVital() {
            const patientId = document.getElementById('patient_id').value || 'patient-1';
            sendVital(patientId, 120, 80);
        }

        function sendAbnormalVital() {
            const patientId = document.getElementById('patient_id').value || 'patient-1';
            sendVital(patientId, 190, 130);
        }

        function sendCustomVital() {
            const patientId = document.getElementById('patient_id').value || 'patient-1';
            const systolic = parseInt(document.getElementById('systolic').value) || 120;
            const diastolic = parseInt(document.getElementById('diastolic').value) || 80;
            sendVital(patientId, systolic, diastolic);
        }

        function renderVitals(vitals) {
            const list = document.getElementById('vitals-list');
            if (!vitals.length) {
                list.innerHTML = '<div class="empty">No vitals yet</div>';
                return;
            }
            list.innerHTML = vitals.slice().reverse().map(v => {
                const isAbnormal = v.systolic > 180 || v.diastolic > 120;
                return '<div class="item ' + (isAbnormal ? 'abnormal' : 'normal') + '">' +
                    '<strong>' + v.patient_id + '</strong>: ' + v.systolic + '/' + v.diastolic +
                    (isAbnormal ? ' ⚠️' : ' ✓') +
                    ' <span class="time">' + formatTime(v.received_at) + '</span>' +
                '</div>';
            }).join('');
        }

        function renderAlerts(alerts) {
            const list = document.getElementById('alerts-list');
            if (!alerts.length) {
                list.innerHTML = '<div class="empty">No alerts</div>';
                return;
            }
            list.innerHTML = alerts.slice().reverse().map(a =>
                '<div class="item abnormal">' +
                    '<strong>' + a.vital.patient_id + '</strong>: ' + a.vital.systolic + '/' + a.vital.diastolic +
                    ' <span class="time">' + formatTime(a.created_at) + '</span>' +
                '</div>'
            ).join('');
        }

        function renderMessages(messages) {
            const list = document.getElementById('messages-list');
            if (!messages.length) {
                list.innerHTML = '<div class="empty">No messages queued</div>';
                return;
            }
            list.innerHTML = messages.slice().reverse().map(m =>
                '<div class="item">' +
                    '<span class="status ' + m.status + '">' + m.status + '</span> ' +
                    '<strong>' + m.patient_id + '</strong>: ' + m.content +
                    ' <span class="time">' + formatTime(m.status === 'SENT' ? m.sent_at : m.queued_at) + '</span>' +
                '</div>'
            ).join('');
        }

        function refreshData() {
            fetch('/vitals').then(r => r.json()).then(data => renderVitals(data.vitals || []));
            fetch('/alerts').then(r => r.json()).then(data => renderAlerts(data.alerts || []));
            fetch('/messages').then(r => r.json()).then(data => renderMessages(data.messages || []));
        }

        // Server-Sent Events for real-time updates
        const eventSource = new EventSource('/events');
        eventSource.onmessage = function(event) {
            const data = JSON.parse(event.data);
            if (data.type === 'message_update') {
                refreshData();
            }
        };

        // Initial load
        refreshData();
    </script>
</body>
</html>
`
