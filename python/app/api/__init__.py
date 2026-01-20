from __future__ import annotations

from flask import Flask

from .alerts import create_blueprint as create_alerts_blueprint
from .errors import register_error_handlers
from .vitals import create_blueprint as create_vitals_blueprint
from ..service import Service


def register_routes(app: Flask, service: Service) -> None:
    app.register_blueprint(create_vitals_blueprint(service))
    app.register_blueprint(create_alerts_blueprint(service))
    register_error_handlers(app)
