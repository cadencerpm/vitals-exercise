from __future__ import annotations

from datetime import datetime, timezone
from typing import Protocol

from .errors import InvalidVitalError
from .models import Alert, Event, EventType, Vital
from .store import Store


class Publisher(Protocol):
    def publish(self, event: Event) -> None:
        ...


class Service:
    def __init__(self, store: Store, publisher: Publisher) -> None:
        self._store = store
        self._publisher = publisher

    def ingest_vital(
        self,
        patient_id: str,
        systolic: int,
        diastolic: int,
        taken_at: datetime | None,
    ) -> Vital:
        patient_id = (patient_id or "").strip()
        if not patient_id:
            raise InvalidVitalError("invalid vital: patient_id is required")
        if systolic <= 0 or diastolic <= 0:
            raise InvalidVitalError("invalid vital: systolic and diastolic must be positive")
        if taken_at is None:
            raise InvalidVitalError("invalid vital: taken_at is required")

        if taken_at.tzinfo is None:
            taken_at = taken_at.replace(tzinfo=timezone.utc)
        taken_at = taken_at.astimezone(timezone.utc)

        vital = Vital(
            patient_id=patient_id,
            systolic=systolic,
            diastolic=diastolic,
            taken_at=taken_at,
            received_at=datetime.now(timezone.utc),
        )

        stored = self._store.add_vital(vital)
        event = Event(type=EventType.VITAL_RECEIVED, vital=stored)
        self._publisher.publish(event)
        return stored

    def list_alerts(self, patient_id: str) -> list[Alert]:
        alerts = self._store.list_alerts()
        patient_id = (patient_id or "").strip()
        if not patient_id:
            return alerts
        return [alert for alert in alerts if alert.patient_id == patient_id]

    def list_vitals(self, patient_id: str) -> list[Vital]:
        vitals = self._store.list_vitals()
        patient_id = (patient_id or "").strip()
        if not patient_id:
            return vitals
        return [vital for vital in vitals if vital.patient_id == patient_id]
