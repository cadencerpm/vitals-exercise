import { MessageQueue } from "./messageQueue";

export class MessageWorker {
  constructor(private queue: MessageQueue) {}

  async run(signal?: AbortSignal): Promise<void> {
    while (!signal?.aborted) {
      try {
        const message = await this.queue.processNext(signal);
        if (!message && !signal?.aborted) {
          // No message in queue, wait a bit before checking again
          await sleep(100);
        }
      } catch (error) {
        console.warn("message worker error:", error);
      }
    }
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
