import test from "node:test";
import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { setTimeout as delay } from "node:timers/promises";
import { AlertWorker } from "../src/domain/alertWorker";
import { PubSub } from "../src/domain/pubsub";
import { InMemoryStore } from "../src/domain/store";
import {
  AlertStatusActive,
  EventTypeVitalReceived,
  Vital,
} from "../src/domain/models";

test("alert worker creates alert for abnormal vitals", async () => {
  const store = new InMemoryStore();
  const pubsub = new PubSub();
  const worker = new AlertWorker(pubsub, store, 8);

  const controller = new AbortController();
  const workerPromise = worker.run(controller.signal);

  const normal: Vital = {
    id: randomUUID(),
    patientId: "patient-normal",
    systolic: 120,
    diastolic: 80,
    takenAt: new Date(),
    receivedAt: new Date(),
  };
  await pubsub.publish({ type: EventTypeVitalReceived, vital: normal });

  const abnormal: Vital = {
    id: randomUUID(),
    patientId: "patient-abnormal",
    systolic: 200,
    diastolic: 130,
    takenAt: new Date(),
    receivedAt: new Date(),
  };
  await pubsub.publish({ type: EventTypeVitalReceived, vital: abnormal });

  await waitFor(
    500,
    async () =>
      (await store.listAlerts({ patientId: "patient-abnormal" })).items
        .length === 1
  );

  const alerts = await store.listAlerts({ patientId: "patient-abnormal" });
  assert.equal(alerts.items.length, 1);
  assert.equal(alerts.items[0].patientId, "patient-abnormal");
  assert.equal(alerts.items[0].vitalId, abnormal.id);
  assert.equal(alerts.items[0].status, AlertStatusActive);

  controller.abort();
  await workerPromise;
});

async function waitFor(
  timeoutMs: number,
  condition: () => Promise<boolean>
): Promise<void> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    if (await condition()) {
      return;
    }
    await delay(10);
  }
  throw new Error("timeout waiting for condition");
}
