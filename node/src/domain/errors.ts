export class RequestError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "RequestError";
  }
}

export class InvalidVitalError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "InvalidVitalError";
  }
}

export class StoreClosedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "StoreClosedError";
  }
}

export class PubSubClosedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "PubSubClosedError";
  }
}
