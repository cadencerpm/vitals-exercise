from __future__ import annotations

import threading
import webbrowser

from .app_factory import create_app
from .alert_worker import AlertWorker
from .message_queue import MessageQueue
from .message_worker import MessageWorker
from .pubsub import PubSub
from .service import Service
from .store import InMemoryStore

HOST = "127.0.0.1"
PORT = 5000


def main() -> None:
    store = InMemoryStore()
    pubsub = PubSub()
    service = Service(store, pubsub)

    # Message queue for patient notifications (5-20 second simulated delay)
    message_queue = MessageQueue(min_delay=5.0, max_delay=20.0)
    message_worker = MessageWorker(message_queue)

    # Alert worker
    alert_worker = AlertWorker(pubsub, store, buffer_size=16)

    stop_event = threading.Event()

    # Start background workers
    threading.Thread(target=alert_worker.run, args=(stop_event,), daemon=True).start()
    threading.Thread(target=message_worker.run, args=(stop_event,), daemon=True).start()

    app = create_app(service, message_queue)

    # Open browser to dashboard
    url = f"http://{HOST}:{PORT}/"
    threading.Timer(1.0, lambda: webbrowser.open(url)).start()

    print(f"Starting server at {url}")

    try:
        app.run(host=HOST, port=PORT, use_reloader=False)
    finally:
        stop_event.set()
        pubsub.close()
        store.close()


if __name__ == "__main__":
    main()
