from __future__ import annotations

from datetime import datetime, timezone
import logging
import queue
import threading

from .models import Alert, AlertStatus, Event, EventType, alert_reason, is_abnormal
from .pubsub import PubSub
from .store import Store


class AlertWorker:
    def __init__(self, pubsub: PubSub, store: Store, buffer_size: int) -> None:
        self._subscription, self._cancel = pubsub.subscribe(buffer_size)
        self._store = store

    def run(self, stop_event: threading.Event) -> None:
        try:
            while not stop_event.is_set():
                if self._subscription.closed.is_set() and self._subscription.queue.empty():
                    return
                try:
                    event = self._subscription.queue.get(timeout=0.1)
                except queue.Empty:
                    continue
                self._handle_event(event)
        finally:
            self._cancel()

    def _handle_event(self, event: Event) -> None:
        if event.type != EventType.VITAL_RECEIVED:
            return
        if not is_abnormal(event.vital):
            return

        alert = Alert(
            vital_id=event.vital.id,
            patient_id=event.vital.patient_id,
            systolic=event.vital.systolic,
            diastolic=event.vital.diastolic,
            taken_at=event.vital.taken_at,
            received_at=event.vital.received_at,
            reason=alert_reason(event.vital),
            status=AlertStatus.ACTIVE,
            created=datetime.now(timezone.utc),
        )

        try:
            self._store.add_alert(alert)
        except Exception as exc:
            logging.warning("alert worker failed to store alert: %s", exc)
