import Fastify from "fastify";
import { registerDashboardRoutes } from "./api/dashboard";
import { registerErrorHandler } from "./api/errors";
import { registerRoutes } from "./api/routes";
import { AlertWorker } from "./domain/alertWorker";
import { MessageQueue } from "./domain/messageQueue";
import { MessageWorker } from "./domain/messageWorker";
import { InMemoryStore, Store } from "./domain/store";
import { PubSub } from "./domain/pubsub";
import { Service } from "./domain/service";

export interface AppContext {
  store: Store;
  pubsub: PubSub;
  service: Service;
  worker: AlertWorker;
  workerController: AbortController;
  messageQueue: MessageQueue;
  messageWorker: MessageWorker;
}

export function buildServer(overrides: Partial<AppContext> = {}) {
  const store = overrides.store ?? new InMemoryStore();
  const pubsub = overrides.pubsub ?? new PubSub();
  const service = overrides.service ?? new Service(store, pubsub);
  const messageQueue = overrides.messageQueue ?? new MessageQueue(5.0, 20.0);
  const messageWorker = overrides.messageWorker ?? new MessageWorker(messageQueue);
  const worker =
    overrides.worker ?? new AlertWorker(pubsub, store, 16, messageQueue);
  const workerController = overrides.workerController ?? new AbortController();

  const app = Fastify({ logger: true });

  registerErrorHandler(app);
  registerRoutes(app, service);
  registerDashboardRoutes(app, service, messageQueue);

  app.addHook("onReady", async () => {
    void worker.run(workerController.signal);
    void messageWorker.run(workerController.signal);
  });

  app.addHook("onClose", async () => {
    workerController.abort();
    pubsub.close();
    store.close();
  });

  return {
    app,
    context: {
      store,
      pubsub,
      service,
      worker,
      workerController,
      messageQueue,
      messageWorker,
    },
  };
}
