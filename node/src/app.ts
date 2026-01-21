import Fastify from "fastify";
import { registerErrorHandler } from "./api/errors";
import { registerRoutes } from "./api/routes";
import { AlertWorker } from "./domain/alertWorker";
import { InMemoryStore, Store } from "./domain/store";
import { PubSub } from "./domain/pubsub";
import { Service } from "./domain/service";

export interface AppContext {
  store: Store;
  pubsub: PubSub;
  service: Service;
  worker: AlertWorker;
  workerController: AbortController;
}

export function buildServer(overrides: Partial<AppContext> = {}) {
  const store = overrides.store ?? new InMemoryStore();
  const pubsub = overrides.pubsub ?? new PubSub();
  const service = overrides.service ?? new Service(store, pubsub);
  const worker = overrides.worker ?? new AlertWorker(pubsub, store, 16);
  const workerController = overrides.workerController ?? new AbortController();

  const app = Fastify({ logger: true });

  registerErrorHandler(app);
  registerRoutes(app, service);

  app.addHook("onReady", async () => {
    void worker.run(workerController.signal);
  });

  app.addHook("onClose", async () => {
    workerController.abort();
    pubsub.close();
    store.close();
  });

  return { app, context: { store, pubsub, service, worker, workerController } };
}
