import { MessageQueue } from "./messageQueue";
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
  private messageQueue: MessageQueue | null;

  constructor(
    pubsub: PubSub,
    private store: Store,
    bufferSize: number,
    messageQueue?: MessageQueue
  ) {
    const { subscription, cancel } = pubsub.subscribe(bufferSize);
    this.subscription = subscription;
    this.cancel = cancel;
    this.messageQueue = messageQueue ?? null;
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

    const reason = alertReason(event.vital);

    try {
      await this.store.addAlert({
        vitalId: event.vital.id,
        patientId: event.vital.patientId,
        systolic: event.vital.systolic,
        diastolic: event.vital.diastolic,
        takenAt: event.vital.takenAt,
        receivedAt: event.vital.receivedAt,
        reason,
        status: AlertStatusActive,
        created: new Date(),
      });

      console.log(`[Alert] ${event.vital.patientId}: ${reason}`);

      if (this.messageQueue) {
        this.messageQueue.enqueue(
          event.vital.patientId,
          `Alert: ${reason}. Please retake your vitals.`
        );
      }
    } catch (error) {
      console.warn("alert worker failed to store alert", error);
    }
  }
}
