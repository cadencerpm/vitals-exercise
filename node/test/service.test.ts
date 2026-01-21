import test from "node:test";
import assert from "node:assert/strict";
import { InMemoryStore } from "../src/domain/store";
import { PubSub } from "../src/domain/pubsub";
import { Service } from "../src/domain/service";
import { AlertStatusActive, EventTypeVitalReceived } from "../src/domain/models";
import { InvalidVitalError, RequestError } from "../src/domain/errors";

const uuidPattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

test("service ingest publishes event", async () => {
  const store = new InMemoryStore();
  const pubsub = new PubSub();
  const service = new Service(store, pubsub);

  const { subscription, cancel } = pubsub.subscribe(1);
  const iterator = subscription[Symbol.asyncIterator]();
  const eventPromise = iterator.next();

  const takenAt = new Date(Date.now() - 2 * 60 * 1000);
  const stored = await service.ingestVital(
    "patient-1",
    120,
    80,
    takenAt
  );

  assert.match(stored.id, uuidPattern);
  const vitals = await store.listVitals({ patientId: "patient-1" });
  assert.equal(vitals.items.length, 1);
  assert.equal(vitals.items[0].patientId, "patient-1");

  const result = await eventPromise;
  assert.equal(result.done, false);
  assert.equal(result.value.type, EventTypeVitalReceived);
  assert.equal(result.value.vital.id, stored.id);

  cancel();
});

test("service ingest validates input", async () => {
  const store = new InMemoryStore();
  const pubsub = new PubSub();
  const service = new Service(store, pubsub);

  await assert.rejects(
    service.ingestVital("", 120, 80, new Date()),
    InvalidVitalError
  );
  await assert.rejects(
    service.ingestVital("patient-1", -1, 80, new Date()),
    InvalidVitalError
  );
  await assert.rejects(
    service.ingestVital("patient-1", 120, 0, new Date()),
    InvalidVitalError
  );
  await assert.rejects(
    service.ingestVital("patient-1", 120, 80, new Date("invalid")),
    InvalidVitalError
  );
});

test("service list alerts filters by patient", async () => {
  const store = new InMemoryStore();
  const pubsub = new PubSub();
  const service = new Service(store, pubsub);

  await store.addAlert({ patientId: "patient-1", status: AlertStatusActive });
  await store.addAlert({ patientId: "patient-2", status: AlertStatusActive });

  const alerts = await service.listAlerts({ patientId: "patient-1" });
  assert.equal(alerts.items.length, 1);
  assert.equal(alerts.items[0].patientId, "patient-1");
});

test("service list alerts requires patient id", async () => {
  const store = new InMemoryStore();
  const pubsub = new PubSub();
  const service = new Service(store, pubsub);

  await assert.rejects(service.listAlerts({ patientId: "" }), RequestError);
});

test("service list vitals filters by patient", async () => {
  const store = new InMemoryStore();
  const pubsub = new PubSub();
  const service = new Service(store, pubsub);

  await store.addVital({ patientId: "patient-1" });
  await store.addVital({ patientId: "patient-2" });

  const vitals = await service.listVitals({ patientId: "patient-1" });
  assert.equal(vitals.items.length, 1);
  assert.equal(vitals.items[0].patientId, "patient-1");
});

test("service list vitals requires patient id", async () => {
  const store = new InMemoryStore();
  const pubsub = new PubSub();
  const service = new Service(store, pubsub);

  await assert.rejects(
    service.listVitals({ patientId: "   " }),
    RequestError
  );
});
