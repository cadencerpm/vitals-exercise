import {
  AlertStatusActive,
  Event,
  EventTypeVitalReceived,
  alertReason,
  isAbnormal,
} from "./models";
import { PubSub, Subscription } from "./pubsub";
import { Store } from "./store";

export class AlertWorker {
  private subscription: Subscription;
  private cancel: () => void;

  constructor(pubsub: PubSub, private store: Store, bufferSize: number) {
    const { subscription, cancel } = pubsub.subscribe(bufferSize);
    this.subscription = subscription;
    this.cancel = cancel;
  }

  async run(signal?: AbortSignal): Promise<void> {
    const abortHandler = () => this.cancel();
    signal?.addEventListener("abort", abortHandler);

    try {
      for await (const event of this.subscription) {
        if (signal?.aborted) {
          break;
        }
        await this.handleEvent(event);
      }
    } finally {
      signal?.removeEventListener("abort", abortHandler);
      this.cancel();
    }
  }

  private async handleEvent(event: Event): Promise<void> {
    if (event.type !== EventTypeVitalReceived) {
      return;
    }
    if (!isAbnormal(event.vital)) {
      return;
    }

    try {
      await this.store.addAlert({
        vitalId: event.vital.id,
        patientId: event.vital.patientId,
        systolic: event.vital.systolic,
        diastolic: event.vital.diastolic,
        takenAt: event.vital.takenAt,
        receivedAt: event.vital.receivedAt,
        reason: alertReason(event.vital),
        status: AlertStatusActive,
        created: new Date(),
      });
    } catch (error) {
      console.warn("alert worker failed to store alert", error);
    }
  }
}
