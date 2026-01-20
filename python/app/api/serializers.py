from __future__ import annotations

from datetime import datetime, timezone

from ..models import Alert, AlertStatus, Vital


def _to_unix(dt: datetime | None) -> int:
    if dt is None:
        return 0
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return int(dt.timestamp())


def vital_to_dict(vital: Vital) -> dict:
    return {
        "id": vital.id,
        "patient_id": vital.patient_id,
        "systolic": vital.systolic,
        "diastolic": vital.diastolic,
        "taken_at": _to_unix(vital.taken_at),
        "received_at": _to_unix(vital.received_at),
    }


def alert_to_dict(alert: Alert) -> dict:
    return {
        "id": alert.id,
        "vital": {
            "id": alert.vital_id,
            "patient_id": alert.patient_id,
            "systolic": alert.systolic,
            "diastolic": alert.diastolic,
            "taken_at": _to_unix(alert.taken_at),
            "received_at": _to_unix(alert.received_at),
        },
        "reason": alert.reason,
        "created_at": _to_unix(alert.created),
        "status": _status_name(alert.status),
    }


def _status_name(status: AlertStatus | None) -> str:
    if status is None:
        return AlertStatus.ACTIVE.name
    return status.name
