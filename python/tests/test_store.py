from __future__ import annotations

import pytest

from app.errors import StoreClosedError
from app.models import Alert, Vital
from app.store import InMemoryStore


def test_in_memory_store_rejects_calls_after_close() -> None:
    store = InMemoryStore()
    store.close()

    with pytest.raises(StoreClosedError):
        store.add_vital(Vital(patient_id="patient-1"))
    with pytest.raises(StoreClosedError):
        store.add_alert(Alert(patient_id="patient-1"))
    with pytest.raises(StoreClosedError):
        store.list_vitals()
    with pytest.raises(StoreClosedError):
        store.list_alerts()
