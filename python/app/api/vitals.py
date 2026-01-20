from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

from flask import Blueprint, jsonify, request

from ..errors import RequestError
from ..service import Service
from .serializers import vital_to_dict


def create_blueprint(service: Service) -> Blueprint:
    bp = Blueprint("vitals", __name__)

    @bp.post("/vitals")
    def ingest_vital():
        payload = request.get_json(silent=True)
        if payload is None:
            raise RequestError("request body is required")

        taken_at = _parse_taken_at(payload.get("taken_at"))
        patient_id = payload.get("patient_id", "")
        systolic = _parse_int(payload.get("systolic"), default=0)
        diastolic = _parse_int(payload.get("diastolic"), default=0)

        vital = service.ingest_vital(patient_id, systolic, diastolic, taken_at)
        return jsonify({"vital": vital_to_dict(vital)})

    @bp.get("/vitals")
    def list_vitals():
        patient_id = request.args.get("patient_id", "")
        vitals = service.list_vitals(patient_id)
        return jsonify({"vitals": [vital_to_dict(vital) for vital in vitals]})

    return bp


def _parse_int(value: Any, default: int = 0) -> int:
    try:
        return int(value)
    except (TypeError, ValueError):
        return default


def _parse_taken_at(value: Any) -> datetime:
    timestamp = _parse_int(value, default=0)
    if timestamp <= 0:
        raise RequestError("taken_at is required")
    try:
        return datetime.fromtimestamp(timestamp, tz=timezone.utc)
    except (OverflowError, OSError, ValueError):
        raise RequestError("taken_at is required")
