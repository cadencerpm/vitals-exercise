from __future__ import annotations

import queue
import random
import threading
import time
from dataclasses import replace
from datetime import datetime, timezone
from typing import Callable

from .models import Message, MessageStatus


class MessageQueue:
    """Simple message queue with simulated processing delays."""

    def __init__(self, min_delay: float = 5.0, max_delay: float = 20.0) -> None:
        self._queue: queue.Queue[Message] = queue.Queue()
        self._min_delay = min_delay
        self._max_delay = max_delay
        self._lock = threading.Lock()
        self._seq = 0
        self._messages: list[Message] = []
        self._listeners: list[Callable[[Message], None]] = []

    def enqueue(self, patient_id: str, content: str) -> Message:
        """Queue a message for delivery."""
        with self._lock:
            self._seq += 1
            message = Message(
                id=self._seq,
                patient_id=patient_id,
                content=content,
                status=MessageStatus.QUEUED,
                queued_at=datetime.now(timezone.utc),
            )
            self._messages.append(message)

        self._queue.put(message)
        self._notify(message)
        return message

    def process_next(self, stop_event: threading.Event) -> Message | None:
        """Process the next message with a simulated delay."""
        try:
            message = self._queue.get(timeout=0.1)
        except queue.Empty:
            return None

        # Mark as processing
        message = self._update(message, MessageStatus.PROCESSING)
        self._notify(message)

        # Simulate delay
        delay = random.uniform(self._min_delay, self._max_delay)
        for _ in range(int(delay * 10)):
            if stop_event.is_set():
                return None
            time.sleep(0.1)

        # Mark as sent
        message = self._update(message, MessageStatus.SENT, sent=True)
        self._notify(message)
        return message

    def _update(self, message: Message, status: MessageStatus, sent: bool = False) -> Message:
        with self._lock:
            sent_at = datetime.now(timezone.utc) if sent else message.sent_at
            updated = replace(message, status=status, sent_at=sent_at)
            for i, m in enumerate(self._messages):
                if m.id == message.id:
                    self._messages[i] = updated
                    break
            return updated

    def list_messages(self) -> list[Message]:
        with self._lock:
            return list(self._messages)

    def add_listener(self, callback: Callable[[Message], None]) -> None:
        with self._lock:
            self._listeners.append(callback)

    def _notify(self, message: Message) -> None:
        with self._lock:
            listeners = list(self._listeners)
        for listener in listeners:
            try:
                listener(message)
            except Exception:
                pass
