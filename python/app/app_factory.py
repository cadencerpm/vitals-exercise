from __future__ import annotations

from flask import Flask

from .api import register_routes
from .service import Service


def create_app(service: Service) -> Flask:
    app = Flask(__name__)
    register_routes(app, service)
    return app
