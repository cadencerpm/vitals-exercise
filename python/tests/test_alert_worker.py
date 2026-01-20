from __future__ import annotations

from datetime import datetime, timezone
import threading
import time

from app.alert_worker import AlertWorker
from app.models import AlertStatus, Event, EventType, Vital
from app.pubsub import PubSub
from app.store import InMemoryStore


def test_alert_worker_creates_alert_for_abnormal_vitals() -> None:
    store = InMemoryStore()
    pubsub = PubSub()
    worker = AlertWorker(pubsub, store, buffer_size=8)

    stop_event = threading.Event()
    worker_thread = threading.Thread(target=worker.run, args=(stop_event,), daemon=True)
    worker_thread.start()

    try:
        normal = Vital(
            id=1,
            patient_id="patient-normal",
            systolic=120,
            diastolic=80,
            taken_at=datetime.now(timezone.utc),
            received_at=datetime.now(timezone.utc),
        )
        pubsub.publish(Event(type=EventType.VITAL_RECEIVED, vital=normal))

        abnormal = Vital(
            id=2,
            patient_id="patient-abnormal",
            systolic=200,
            diastolic=130,
            taken_at=datetime.now(timezone.utc),
            received_at=datetime.now(timezone.utc),
        )
        pubsub.publish(Event(type=EventType.VITAL_RECEIVED, vital=abnormal))

        _wait_for(0.5, lambda: len(store.list_alerts()) == 1)

        alerts = store.list_alerts()
        assert len(alerts) == 1
        assert alerts[0].patient_id == "patient-abnormal"
        assert alerts[0].vital_id == abnormal.id
        assert alerts[0].status == AlertStatus.ACTIVE
    finally:
        stop_event.set()


def _wait_for(timeout: float, condition) -> None:
    deadline = time.time() + timeout
    while time.time() < deadline:
        if condition():
            return
        time.sleep(0.01)
    raise AssertionError("timeout waiting for condition")
