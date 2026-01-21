import test from "node:test";
import assert from "node:assert/strict";
import { InMemoryStore } from "../src/domain/store";

test("in-memory store honors aborted signals", async () => {
  const store = new InMemoryStore();
  const controller = new AbortController();
  controller.abort();

  await assert.rejects(
    store.addVital({ patientId: "patient-1" }, controller.signal),
    { name: "AbortError" }
  );
  await assert.rejects(
    store.addAlert({ patientId: "patient-1" }, controller.signal),
    { name: "AbortError" }
  );
  await assert.rejects(
    store.listVitals({ patientId: "patient-1" }, controller.signal),
    {
      name: "AbortError",
    }
  );
  await assert.rejects(
    store.listAlerts({ patientId: "patient-1" }, controller.signal),
    {
      name: "AbortError",
    }
  );

  const vitals = await store.listVitals({ patientId: "patient-1" });
  assert.equal(vitals.items.length, 0);

  const alerts = await store.listAlerts({ patientId: "patient-1" });
  assert.equal(alerts.items.length, 0);
});

test("in-memory store paginates vitals by takenAt", async () => {
  const store = new InMemoryStore();
  const patientId = "patient-1";

  const first = await store.addVital({
    patientId,
    systolic: 120,
    diastolic: 80,
    takenAt: new Date("2024-01-01T00:00:00Z"),
    receivedAt: new Date("2024-01-01T00:00:01Z"),
  });
  const second = await store.addVital({
    patientId,
    systolic: 121,
    diastolic: 81,
    takenAt: new Date("2024-01-02T00:00:00Z"),
    receivedAt: new Date("2024-01-02T00:00:01Z"),
  });
  const third = await store.addVital({
    patientId,
    systolic: 122,
    diastolic: 82,
    takenAt: new Date("2024-01-03T00:00:00Z"),
    receivedAt: new Date("2024-01-03T00:00:01Z"),
  });

  const pageOne = await store.listVitals({ patientId, limit: 2 });
  assert.deepEqual(
    pageOne.items.map((vital) => vital.id),
    [third.id, second.id]
  );
  assert.ok(pageOne.nextCursor);

  const pageTwo = await store.listVitals({
    patientId,
    limit: 2,
    cursor: pageOne.nextCursor,
  });
  assert.deepEqual(
    pageTwo.items.map((vital) => vital.id),
    [first.id]
  );
  assert.equal(pageTwo.nextCursor, undefined);
});
