from __future__ import annotations

from flask import Flask

from .api import register_routes
from .message_queue import MessageQueue
from .service import Service


def create_app(
    service: Service, message_queue: MessageQueue | None = None
) -> Flask:
    app = Flask(__name__)
    register_routes(app, service, message_queue)
    return app
