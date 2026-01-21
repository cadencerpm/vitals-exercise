import { InvalidVitalError, RequestError } from "./errors";
import {
  Alert,
  Event,
  EventTypeVitalReceived,
  Vital,
  VitalInput,
} from "./models";
import { ListOptions, ListResult, Store } from "./store";

export interface Publisher {
  publish(event: Event): Promise<void>;
}

export class Service {
  constructor(private store: Store, private publisher: Publisher) {}

  async ingestVital(
    patientId: string,
    systolic: number,
    diastolic: number,
    takenAt: Date,
    signal?: AbortSignal
  ): Promise<Vital> {
    const trimmed = (patientId || "").trim();
    if (!trimmed) {
      throw new InvalidVitalError("invalid vital: patientId is required");
    }
    if (systolic <= 0 || diastolic <= 0) {
      throw new InvalidVitalError(
        "invalid vital: systolic and diastolic must be positive"
      );
    }
    if (!takenAt || Number.isNaN(takenAt.getTime())) {
      throw new InvalidVitalError("invalid vital: takenAt is required");
    }

    const vital: VitalInput = {
      patientId: trimmed,
      systolic,
      diastolic,
      takenAt: new Date(takenAt.getTime()),
      receivedAt: new Date(),
    };

    const stored = await this.store.addVital(vital, signal);
    const event: Event = { type: EventTypeVitalReceived, vital: stored };
    await this.publisher.publish(event);
    return stored;
  }

  async listAlerts(
    options: ListOptions,
    signal?: AbortSignal
  ): Promise<ListResult<Alert>> {
    const patientId = (options.patientId || "").trim();
    if (!patientId) {
      throw new RequestError("patientId is required");
    }
    return this.store.listAlerts({ ...options, patientId }, signal);
  }

  async listVitals(
    options: ListOptions,
    signal?: AbortSignal
  ): Promise<ListResult<Vital>> {
    const patientId = (options.patientId || "").trim();
    if (!patientId) {
      throw new RequestError("patientId is required");
    }
    return this.store.listVitals({ ...options, patientId }, signal);
  }
}
