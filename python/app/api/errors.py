from __future__ import annotations

from flask import Flask, jsonify
from werkzeug.exceptions import HTTPException

from ..errors import InvalidVitalError, RequestError


def register_error_handlers(app: Flask) -> None:
    @app.errorhandler(RequestError)
    def handle_request_error(error: RequestError):
        return jsonify({"error": str(error)}), 400

    @app.errorhandler(InvalidVitalError)
    def handle_invalid_vital(error: InvalidVitalError):
        return jsonify({"error": str(error)}), 400

    @app.errorhandler(HTTPException)
    def handle_http_exception(error: HTTPException):
        return jsonify({"error": error.description}), error.code or 500

    @app.errorhandler(Exception)
    def handle_unexpected(error: Exception):
        return jsonify({"error": str(error)}), 500
