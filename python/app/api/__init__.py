from __future__ import annotations

from flask import Flask

from .alerts import create_blueprint as create_alerts_blueprint
from .dashboard import create_blueprint as create_dashboard_blueprint
from .errors import register_error_handlers
from .vitals import create_blueprint as create_vitals_blueprint
from ..message_queue import MessageQueue
from ..service import Service


def register_routes(
    app: Flask, service: Service, message_queue: MessageQueue | None = None
) -> None:
    app.register_blueprint(create_vitals_blueprint(service))
    app.register_blueprint(create_alerts_blueprint(service))
    if message_queue is not None:
        app.register_blueprint(create_dashboard_blueprint(service, message_queue))
    register_error_handlers(app)
