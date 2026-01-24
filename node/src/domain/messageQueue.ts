export type MessageStatus = "QUEUED" | "PROCESSING" | "SENT";

export const MessageStatusQueued: MessageStatus = "QUEUED";
export const MessageStatusProcessing: MessageStatus = "PROCESSING";
export const MessageStatusSent: MessageStatus = "SENT";

export interface Message {
  id: number;
  patientId: string;
  content: string;
  status: MessageStatus;
  queuedAt: Date;
  sentAt: Date | null;
}

export type MessageListener = (message: Message) => void;

export class MessageQueue {
  private queue: Message[] = [];
  private messages: Message[] = [];
  private listeners: MessageListener[] = [];
  private seq = 0;
  private minDelay: number;
  private maxDelay: number;

  constructor(minDelay = 5.0, maxDelay = 20.0) {
    this.minDelay = minDelay;
    this.maxDelay = maxDelay;
  }

  enqueue(patientId: string, content: string): Message {
    this.seq += 1;
    const message: Message = {
      id: this.seq,
      patientId,
      content,
      status: MessageStatusQueued,
      queuedAt: new Date(),
      sentAt: null,
    };
    this.messages.push(message);
    this.queue.push(message);
    this.notify(message);
    return message;
  }

  async processNext(signal?: AbortSignal): Promise<Message | null> {
    const message = this.queue.shift();
    if (!message) {
      return null;
    }

    // Mark as processing
    const processing = this.updateMessage(message, MessageStatusProcessing);
    this.notify(processing);

    // Simulate delay
    const delay =
      this.minDelay + Math.random() * (this.maxDelay - this.minDelay);
    const delayMs = delay * 1000;
    const startTime = Date.now();

    while (Date.now() - startTime < delayMs) {
      if (signal?.aborted) {
        return null;
      }
      await sleep(100);
    }

    // Mark as sent
    const sent = this.updateMessage(processing, MessageStatusSent, true);
    this.notify(sent);
    return sent;
  }

  private updateMessage(
    message: Message,
    status: MessageStatus,
    sent = false
  ): Message {
    const sentAt = sent ? new Date() : message.sentAt;
    const updated: Message = { ...message, status, sentAt };

    const index = this.messages.findIndex((m) => m.id === message.id);
    if (index !== -1) {
      this.messages[index] = updated;
    }
    return updated;
  }

  listMessages(): Message[] {
    return [...this.messages];
  }

  addListener(callback: MessageListener): void {
    this.listeners.push(callback);
  }

  private notify(message: Message): void {
    for (const listener of this.listeners) {
      try {
        listener(message);
      } catch {
        // Ignore listener errors
      }
    }
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
