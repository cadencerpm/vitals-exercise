class RequestError(ValueError):
    """Raised for malformed HTTP requests."""


class InvalidVitalError(ValueError):
    """Raised when a vital fails validation."""


class StoreClosedError(RuntimeError):
    """Raised when the store has been closed."""


class PubSubClosedError(RuntimeError):
    """Raised when the pubsub is closed."""
