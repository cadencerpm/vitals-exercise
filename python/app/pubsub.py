from __future__ import annotations

import queue
import threading
from typing import Callable

from .errors import PubSubClosedError
from .models import Event


class Subscription:
    def __init__(self, buffer_size: int) -> None:
        self.queue: queue.Queue[Event] = queue.Queue(maxsize=buffer_size)
        self.closed = threading.Event()


class PubSub:
    def __init__(self) -> None:
        self._lock = threading.RLock()
        self._closed = False
        self._subscription: Subscription | None = None

    def publish(self, event: Event) -> None:
        with self._lock:
            if self._closed:
                raise PubSubClosedError("pubsub is closed")
            subscription = self._subscription

        if subscription is None or subscription.closed.is_set():
            return

        while True:
            if subscription.closed.is_set():
                return
            try:
                subscription.queue.put(event, timeout=0.1)
                return
            except queue.Full:
                continue

    def subscribe(self, buffer_size: int) -> tuple[Subscription, Callable[[], None]]:
        if buffer_size <= 0:
            buffer_size = 1
        subscription = Subscription(buffer_size)

        with self._lock:
            if self._closed or self._subscription is not None:
                subscription.closed.set()
                return subscription, lambda: None
            self._subscription = subscription

        def cancel() -> None:
            with self._lock:
                if self._subscription is subscription:
                    self._subscription = None
                    subscription.closed.set()

        return subscription, cancel

    def close(self) -> None:
        with self._lock:
            if self._closed:
                return
            self._closed = True
            if self._subscription is not None:
                self._subscription.closed.set()
                self._subscription = None
