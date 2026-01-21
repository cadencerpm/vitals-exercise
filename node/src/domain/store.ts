import { randomUUID } from "node:crypto";
import {
  Alert,
  AlertInput,
  AlertStatusActive,
  Vital,
  VitalInput,
} from "./models";
import { RequestError, StoreClosedError } from "./errors";

export interface Store {
  addVital(vital: VitalInput, signal?: AbortSignal): Promise<Vital>;
  addAlert(alert: AlertInput, signal?: AbortSignal): Promise<Alert>;
  listAlerts(options: ListOptions, signal?: AbortSignal): Promise<ListResult<Alert>>;
  listVitals(options: ListOptions, signal?: AbortSignal): Promise<ListResult<Vital>>;
  close(): void;
}

export type ListOptions = {
  patientId: string;
  limit?: number;
  cursor?: string;
};

export type ListResult<T> = {
  items: T[];
  nextCursor?: string;
};

const DefaultLimit = 100;
const MaxLimit = 100;

export class InMemoryStore implements Store {
  private closed = false;
  private vitals: Vital[] = [];
  private alerts: Alert[] = [];

  async addVital(vital: VitalInput, signal?: AbortSignal): Promise<Vital> {
    ensureActive(signal);
    if (this.closed) {
      throw new StoreClosedError("store is closed");
    }

    const id = randomUUID();
    const receivedAt = vital.receivedAt ?? new Date();
    const takenAt = vital.takenAt ?? receivedAt;

    const stored: Vital = {
      id,
      patientId: vital.patientId,
      systolic: vital.systolic ?? 0,
      diastolic: vital.diastolic ?? 0,
      takenAt,
      receivedAt,
    };

    this.vitals.push(stored);
    return stored;
  }

  async addAlert(alert: AlertInput, signal?: AbortSignal): Promise<Alert> {
    ensureActive(signal);
    if (this.closed) {
      throw new StoreClosedError("store is closed");
    }

    const id = randomUUID();
    const created = alert.created ?? new Date();
    const receivedAt = alert.receivedAt ?? created;
    const takenAt = alert.takenAt ?? receivedAt;

    const stored: Alert = {
      id,
      vitalId: alert.vitalId ?? "",
      patientId: alert.patientId,
      systolic: alert.systolic ?? 0,
      diastolic: alert.diastolic ?? 0,
      takenAt,
      receivedAt,
      reason: alert.reason ?? "",
      status: alert.status ?? AlertStatusActive,
      created,
    };

    this.alerts.push(stored);
    return stored;
  }

  async listAlerts(
    options: ListOptions,
    signal?: AbortSignal
  ): Promise<ListResult<Alert>> {
    ensureActive(signal);
    if (this.closed) {
      throw new StoreClosedError("store is closed");
    }
    const patientId = normalizePatientId(options.patientId);
    const limit = normalizeLimit(options.limit);
    const cursor = parseCursor(options.cursor);

    const filtered = this.alerts
      .filter((alert) => alert.patientId === patientId)
      .sort(compareByTakenAtDesc);
    const paged = applyCursor(filtered, cursor);
    const items = paged.slice(0, limit);
    const nextCursor =
      paged.length > limit && items.length > 0
        ? makeCursor(items[items.length - 1])
        : undefined;

    return { items, nextCursor };
  }

  async listVitals(
    options: ListOptions,
    signal?: AbortSignal
  ): Promise<ListResult<Vital>> {
    ensureActive(signal);
    if (this.closed) {
      throw new StoreClosedError("store is closed");
    }
    const patientId = normalizePatientId(options.patientId);
    const limit = normalizeLimit(options.limit);
    const cursor = parseCursor(options.cursor);

    const filtered = this.vitals
      .filter((vital) => vital.patientId === patientId)
      .sort(compareByTakenAtDesc);
    const paged = applyCursor(filtered, cursor);
    const items = paged.slice(0, limit);
    const nextCursor =
      paged.length > limit && items.length > 0
        ? makeCursor(items[items.length - 1])
        : undefined;

    return { items, nextCursor };
  }

  close(): void {
    this.closed = true;
    this.vitals = [];
    this.alerts = [];
  }
}

type Cursor = {
  timeMs: number;
  id: string;
};

function normalizePatientId(value: string): string {
  const trimmed = (value || "").trim();
  if (!trimmed) {
    throw new RequestError("patientId is required");
  }
  return trimmed;
}

function normalizeLimit(limit?: number): number {
  if (limit === undefined) {
    return DefaultLimit;
  }
  if (!Number.isFinite(limit)) {
    throw new RequestError("limit must be a positive integer");
  }
  const normalized = Math.floor(limit);
  if (normalized <= 0) {
    throw new RequestError("limit must be a positive integer");
  }
  return Math.min(normalized, MaxLimit);
}

function parseCursor(cursor?: string): Cursor | null {
  if (!cursor) {
    return null;
  }
  const [timePart, idPart] = cursor.split(":");
  if (!timePart || !idPart) {
    throw new RequestError("invalid cursor");
  }
  const timeMs = Number(timePart);
  if (!Number.isFinite(timeMs) || timeMs < 0) {
    throw new RequestError("invalid cursor");
  }
  return { timeMs, id: idPart };
}

function makeCursor(item: { takenAt: Date; id: string }): string {
  return `${item.takenAt.getTime()}:${item.id}`;
}

function compareByTakenAtDesc(
  a: { takenAt: Date; id: string },
  b: { takenAt: Date; id: string }
): number {
  const aTime = a.takenAt.getTime();
  const bTime = b.takenAt.getTime();
  if (aTime !== bTime) {
    return bTime - aTime;
  }
  if (a.id === b.id) {
    return 0;
  }
  return a.id < b.id ? 1 : -1;
}

function applyCursor<T extends { takenAt: Date; id: string }>(
  items: T[],
  cursor: Cursor | null
): T[] {
  if (!cursor) {
    return items;
  }
  return items.filter((item) => {
    const timeMs = item.takenAt.getTime();
    if (timeMs < cursor.timeMs) {
      return true;
    }
    if (timeMs > cursor.timeMs) {
      return false;
    }
    return item.id < cursor.id;
  });
}

function ensureActive(signal?: AbortSignal): void {
  if (!signal) {
    return;
  }

  if (typeof signal.throwIfAborted === "function") {
    signal.throwIfAborted();
    return;
  }

  if (signal.aborted) {
    if (signal.reason instanceof Error) {
      throw signal.reason;
    }

    const error = new Error("operation aborted");
    error.name = "AbortError";
    throw error;
  }
}
