from __future__ import annotations

from dataclasses import replace
from datetime import datetime, timezone
import threading
from typing import Protocol

from .errors import StoreClosedError
from .models import Alert, Vital


class Store(Protocol):
    def add_vital(self, vital: Vital) -> Vital:
        ...

    def add_alert(self, alert: Alert) -> Alert:
        ...

    def list_alerts(self) -> list[Alert]:
        ...

    def list_vitals(self) -> list[Vital]:
        ...

    def close(self) -> None:
        ...


class InMemoryStore:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._closed = False
        self._vital_seq = 0
        self._alert_seq = 0
        self._vitals: list[Vital] = []
        self._alerts: list[Alert] = []

    def add_vital(self, vital: Vital) -> Vital:
        with self._lock:
            if self._closed:
                raise StoreClosedError("store is closed")

            vital_id = vital.id
            if vital_id == 0:
                self._vital_seq += 1
                vital_id = self._vital_seq

            received_at = vital.received_at
            if received_at is None:
                received_at = datetime.now(timezone.utc)

            stored = replace(vital, id=vital_id, received_at=received_at)
            self._vitals.append(stored)
            return stored

    def add_alert(self, alert: Alert) -> Alert:
        with self._lock:
            if self._closed:
                raise StoreClosedError("store is closed")

            alert_id = alert.id
            if alert_id == 0:
                self._alert_seq += 1
                alert_id = self._alert_seq

            created_at = alert.created
            if created_at is None:
                created_at = datetime.now(timezone.utc)

            stored = replace(alert, id=alert_id, created=created_at)
            self._alerts.append(stored)
            return stored

    def list_alerts(self) -> list[Alert]:
        with self._lock:
            if self._closed:
                raise StoreClosedError("store is closed")
            return list(self._alerts)

    def list_vitals(self) -> list[Vital]:
        with self._lock:
            if self._closed:
                raise StoreClosedError("store is closed")
            return list(self._vitals)

    def close(self) -> None:
        with self._lock:
            self._closed = True
            self._vitals = []
            self._alerts = []
