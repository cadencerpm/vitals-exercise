import { PubSubClosedError } from "./errors";
import { Event } from "./models";

class AsyncQueue<T> {
  private items: T[] = [];
  private waiters: Array<(result: IteratorResult<T>) => void> = [];
  private spaceWaiters: Array<() => void> = [];
  private closed = false;

  constructor(private maxSize: number) {}

  async push(item: T): Promise<void> {
    if (this.closed) {
      throw new PubSubClosedError("subscription is closed");
    }

    while (this.items.length >= this.maxSize && !this.closed) {
      await new Promise<void>((resolve) => this.spaceWaiters.push(resolve));
    }

    if (this.closed) {
      throw new PubSubClosedError("subscription is closed");
    }

    const waiter = this.waiters.shift();
    if (waiter) {
      waiter({ value: item, done: false });
      return;
    }

    this.items.push(item);
  }

  async shift(): Promise<IteratorResult<T>> {
    if (this.items.length > 0) {
      const item = this.items.shift() as T;
      this.notifySpace();
      return { value: item, done: false };
    }

    if (this.closed) {
      return { value: undefined as T, done: true };
    }

    return new Promise<IteratorResult<T>>((resolve) => {
      this.waiters.push(resolve);
    });
  }

  close(): void {
    if (this.closed) {
      return;
    }
    this.closed = true;
    while (this.waiters.length > 0) {
      const resolve = this.waiters.shift();
      if (resolve) {
        resolve({ value: undefined as T, done: true });
      }
    }
    this.notifyAllSpace();
  }

  private notifySpace(): void {
    const resolve = this.spaceWaiters.shift();
    if (resolve) {
      resolve();
    }
  }

  private notifyAllSpace(): void {
    while (this.spaceWaiters.length > 0) {
      const resolve = this.spaceWaiters.shift();
      if (resolve) {
        resolve();
      }
    }
  }
}

export class Subscription implements AsyncIterable<Event> {
  constructor(private queue: AsyncQueue<Event>) {}

  [Symbol.asyncIterator](): AsyncIterator<Event> {
    return {
      next: () => this.queue.shift(),
    };
  }
}

export class PubSub {
  private closed = false;
  private queue: AsyncQueue<Event> | null = null;

  async publish(event: Event): Promise<void> {
    if (this.closed) {
      throw new PubSubClosedError("pubsub is closed");
    }
    if (!this.queue) {
      return;
    }
    await this.queue.push(event);
  }

  subscribe(bufferSize: number): { subscription: Subscription; cancel: () => void } {
    const size = bufferSize > 0 ? bufferSize : 1;
    const queue = new AsyncQueue<Event>(size);

    if (this.closed || this.queue) {
      queue.close();
      return { subscription: new Subscription(queue), cancel: () => {} };
    }

    this.queue = queue;

    const cancel = () => {
      if (this.queue === queue) {
        this.queue = null;
      }
      queue.close();
    };

    return { subscription: new Subscription(queue), cancel };
  }

  close(): void {
    if (this.closed) {
      return;
    }
    this.closed = true;
    if (this.queue) {
      this.queue.close();
      this.queue = null;
    }
  }
}
