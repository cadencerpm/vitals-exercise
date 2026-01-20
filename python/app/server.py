from __future__ import annotations

import argparse
import threading

from .app_factory import create_app
from .alert_worker import AlertWorker
from .pubsub import PubSub
from .service import Service
from .store import InMemoryStore


def main() -> None:
    parser = argparse.ArgumentParser(description="Vitals HTTP server")
    parser.add_argument("--host", default="127.0.0.1", help="listen host")
    parser.add_argument("--port", type=int, default=5000, help="listen port")
    args = parser.parse_args()

    store = InMemoryStore()
    pubsub = PubSub()
    service = Service(store, pubsub)
    worker = AlertWorker(pubsub, store, buffer_size=16)

    stop_event = threading.Event()
    worker_thread = threading.Thread(target=worker.run, args=(stop_event,), daemon=True)
    worker_thread.start()

    app = create_app(service)
    try:
        app.run(host=args.host, port=args.port, use_reloader=False)
    finally:
        stop_event.set()
        pubsub.close()
        store.close()


if __name__ == "__main__":
    main()
