from __future__ import annotations

import json
import queue
import threading
from flask import Blueprint, Response, render_template_string

from ..message_queue import MessageQueue
from ..models import Message
from ..service import Service
from .serializers import alert_to_dict, message_to_dict, vital_to_dict

DASHBOARD_HTML = """
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
        function formatTime(unix) {
            if (!unix) return '';
            return new Date(unix * 1000).toLocaleTimeString();
        }

        function sendVital(patientId, systolic, diastolic) {
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
                return `<div class="item ${isAbnormal ? 'abnormal' : 'normal'}">
                    <strong>${v.patient_id}</strong>: ${v.systolic}/${v.diastolic}
                    ${isAbnormal ? '⚠️' : '✓'}
                    <span class="time">${formatTime(v.received_at)}</span>
                </div>`;
            }).join('');
        }

        function renderAlerts(alerts) {
            const list = document.getElementById('alerts-list');
            if (!alerts.length) {
                list.innerHTML = '<div class="empty">No alerts</div>';
                return;
            }
            list.innerHTML = alerts.slice().reverse().map(a =>
                `<div class="item abnormal">
                    <strong>${a.vital.patient_id}</strong>: ${a.vital.systolic}/${a.vital.diastolic}
                    <span class="time">${formatTime(a.created_at)}</span>
                </div>`
            ).join('');
        }

        function renderMessages(messages) {
            const list = document.getElementById('messages-list');
            if (!messages.length) {
                list.innerHTML = '<div class="empty">No messages queued</div>';
                return;
            }
            list.innerHTML = messages.slice().reverse().map(m =>
                `<div class="item">
                    <span class="status ${m.status}">${m.status}</span>
                    <strong>${m.patient_id}</strong>: ${m.content}
                    <span class="time">${formatTime(m.status === 'SENT' ? m.sent_at : m.queued_at)}</span>
                </div>`
            ).join('');
        }

        function refreshData() {
            fetch('/vitals').then(r => r.json()).then(data => renderVitals(data.vitals));
            fetch('/alerts').then(r => r.json()).then(data => renderAlerts(data.alerts));
            fetch('/messages').then(r => r.json()).then(data => renderMessages(data.messages));
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
"""


def create_blueprint(service: Service, message_queue: MessageQueue) -> Blueprint:
    bp = Blueprint("dashboard", __name__)

    # Store for SSE clients
    sse_clients: list[queue.Queue] = []
    sse_lock = threading.Lock()

    def broadcast_message_update(message: Message) -> None:
        """Called when a message status changes."""
        data = json.dumps({"type": "message_update", "message": message_to_dict(message)})
        with sse_lock:
            for client_queue in sse_clients:
                try:
                    client_queue.put_nowait(data)
                except queue.Full:
                    pass

    # Register listener for message updates
    message_queue.add_listener(broadcast_message_update)

    @bp.get("/")
    def dashboard():
        return render_template_string(DASHBOARD_HTML)

    @bp.get("/messages")
    def list_messages():
        messages = message_queue.list_messages()
        return {"messages": [message_to_dict(m) for m in messages]}

    @bp.get("/events")
    def sse_events():
        """Server-Sent Events endpoint for real-time updates."""
        def generate():
            client_queue: queue.Queue = queue.Queue(maxsize=100)
            with sse_lock:
                sse_clients.append(client_queue)
            try:
                yield "data: {\"type\": \"connected\"}\n\n"
                while True:
                    try:
                        data = client_queue.get(timeout=30)
                        yield f"data: {data}\n\n"
                    except queue.Empty:
                        yield ": keepalive\n\n"
            finally:
                with sse_lock:
                    if client_queue in sse_clients:
                        sse_clients.remove(client_queue)

        return Response(generate(), mimetype="text/event-stream")

    return bp
