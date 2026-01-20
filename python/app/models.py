from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from datetime import datetime

SYSTOLIC_UPPER_ALERT_THRESHOLD = 180
DIASTOLIC_UPPER_ALERT_THRESHOLD = 120


class EventType(str, Enum):
    VITAL_RECEIVED = "VITAL_RECEIVED"


class AlertStatus(Enum):
    ACTIVE = "ACTIVE"
    RESOLVED = "RESOLVED"
    AUTO_RESOLVED = "AUTO_RESOLVED"


@dataclass
class Vital:
    id: int = 0
    patient_id: str = ""
    systolic: int = 0
    diastolic: int = 0
    taken_at: datetime | None = None
    received_at: datetime | None = None


@dataclass
class Alert:
    id: int = 0
    vital_id: int = 0
    patient_id: str = ""
    systolic: int = 0
    diastolic: int = 0
    taken_at: datetime | None = None
    received_at: datetime | None = None
    reason: str = ""
    status: AlertStatus = AlertStatus.ACTIVE
    created: datetime | None = None


@dataclass
class Event:
    type: EventType
    vital: Vital


def is_abnormal(vital: Vital) -> bool:
    return (
        vital.systolic > SYSTOLIC_UPPER_ALERT_THRESHOLD
        or vital.diastolic > DIASTOLIC_UPPER_ALERT_THRESHOLD
    )


def alert_reason(vital: Vital) -> str:
    return f"abnormal blood pressure {vital.systolic}/{vital.diastolic}"
