from __future__ import annotations

from datetime import datetime, timedelta, timezone

import pytest

from app.errors import InvalidVitalError
from app.models import Alert, AlertStatus, EventType, Vital
from app.pubsub import PubSub
from app.service import Service
from app.store import InMemoryStore


def test_service_ingest_publishes_event() -> None:
    store = InMemoryStore()
    pubsub = PubSub()
    service = Service(store, pubsub)

    subscription, cancel = pubsub.subscribe(1)
    try:
        taken_at = datetime.now(timezone.utc) - timedelta(minutes=2)
        stored = service.ingest_vital("patient-1", 120, 80, taken_at)
        assert stored.id != 0

        vitals = store.list_vitals()
        assert len(vitals) == 1
        assert vitals[0].patient_id == "patient-1"

        event = subscription.queue.get(timeout=0.5)
        assert event.type == EventType.VITAL_RECEIVED
        assert event.vital.id == stored.id
    finally:
        cancel()


def test_service_ingest_validates_input() -> None:
    store = InMemoryStore()
    pubsub = PubSub()
    service = Service(store, pubsub)

    with pytest.raises(InvalidVitalError):
        service.ingest_vital("", 120, 80, datetime.now(timezone.utc))
    with pytest.raises(InvalidVitalError):
        service.ingest_vital("patient-1", -1, 80, datetime.now(timezone.utc))
    with pytest.raises(InvalidVitalError):
        service.ingest_vital("patient-1", 120, 0, datetime.now(timezone.utc))
    with pytest.raises(InvalidVitalError):
        service.ingest_vital("patient-1", 120, 80, None)


def test_service_list_alerts_filters_by_patient() -> None:
    store = InMemoryStore()
    pubsub = PubSub()
    service = Service(store, pubsub)

    store.add_alert(Alert(patient_id="patient-1", status=AlertStatus.ACTIVE))
    store.add_alert(Alert(patient_id="patient-2", status=AlertStatus.ACTIVE))

    alerts = service.list_alerts("patient-1")
    assert len(alerts) == 1
    assert alerts[0].patient_id == "patient-1"


def test_service_list_vitals_filters_by_patient() -> None:
    store = InMemoryStore()
    pubsub = PubSub()
    service = Service(store, pubsub)

    store.add_vital(Vital(patient_id="patient-1"))
    store.add_vital(Vital(patient_id="patient-2"))

    vitals = service.list_vitals("patient-1")
    assert len(vitals) == 1
    assert vitals[0].patient_id == "patient-1"
