from __future__ import annotations

from flask import Blueprint, jsonify, request

from ..service import Service
from .serializers import alert_to_dict


def create_blueprint(service: Service) -> Blueprint:
    bp = Blueprint("alerts", __name__)

    @bp.get("/alerts")
    def list_alerts():
        patient_id = request.args.get("patient_id", "")
        alerts = service.list_alerts(patient_id)
        return jsonify({"alerts": [alert_to_dict(alert) for alert in alerts]})

    return bp
