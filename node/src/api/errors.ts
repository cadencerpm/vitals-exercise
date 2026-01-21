import type { FastifyInstance } from "fastify";
import { InvalidVitalError, RequestError } from "../domain/errors";

export function registerErrorHandler(app: FastifyInstance): void {
  app.setErrorHandler((error, request, reply) => {
    if (error instanceof RequestError || error instanceof InvalidVitalError) {
      reply.status(400).send({ error: error.message });
      return;
    }

    if (isValidationError(error)) {
      reply.status(400).send({ error: error.message });
      return;
    }

    request.log.error({ err: error }, "request failed");
    reply.status(500).send({ error: error.message });
  });
}

function isValidationError(error: unknown): boolean {
  if (!error || typeof error !== "object") {
    return false;
  }
  return "validation" in error || "validationContext" in error;
}
