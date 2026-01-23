from __future__ import annotations

import logging
import threading

from .message_queue import MessageQueue


class MessageWorker:
    """Background worker that processes queued messages."""

    def __init__(self, message_queue: MessageQueue) -> None:
        self._queue = message_queue

    def run(self, stop_event: threading.Event) -> None:
        while not stop_event.is_set():
            try:
                self._queue.process_next(stop_event)
            except Exception as exc:
                logging.warning("message worker error: %s", exc)
