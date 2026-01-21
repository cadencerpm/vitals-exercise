import type { FastifyInstance } from "fastify";
import { RequestError } from "../domain/errors";
import { Service } from "../domain/service";
import { alertToResponse, vitalToResponse } from "./serializers";

type IngestVitalBody = {
  patientId: string;
  systolic: number;
  diastolic: number;
  takenAt: number;
};

type ListQuery = {
  patientId: string;
  limit?: number;
  cursor?: string;
};

export function registerRoutes(app: FastifyInstance, service: Service): void {
  app.post<{ Body: IngestVitalBody }>(
    "/vitals",
    {
      schema: {
        body: {
          type: "object",
          required: ["patientId", "systolic", "diastolic", "takenAt"],
          additionalProperties: false,
          properties: {
            patientId: { type: "string", minLength: 1 },
            systolic: { type: "integer", minimum: 1 },
            diastolic: { type: "integer", minimum: 1 },
            takenAt: { type: "integer", minimum: 1 },
          },
        },
      },
    },
    async (request) => {
      const { patientId, systolic, diastolic, takenAt } = request.body;
      const takenAtDate = parseTakenAt(takenAt);
      const vital = await service.ingestVital(
        patientId,
        systolic,
        diastolic,
        takenAtDate
      );
      return { vital: vitalToResponse(vital) };
    }
  );

  app.get<{ Querystring: ListQuery }>(
    "/vitals",
    {
      schema: {
        querystring: {
          type: "object",
          required: ["patientId"],
          additionalProperties: false,
          properties: {
            patientId: { type: "string", minLength: 1 },
            limit: { type: "integer", minimum: 1, maximum: 100 },
            cursor: { type: "string" },
          },
        },
      },
    },
    async (request) => {
      const { patientId, limit, cursor } = request.query;
      const result = await service.listVitals({ patientId, limit, cursor });
      return {
        vitals: result.items.map(vitalToResponse),
        nextCursor: result.nextCursor,
      };
    }
  );

  app.get<{ Querystring: ListQuery }>(
    "/alerts",
    {
      schema: {
        querystring: {
          type: "object",
          required: ["patientId"],
          additionalProperties: false,
          properties: {
            patientId: { type: "string", minLength: 1 },
            limit: { type: "integer", minimum: 1, maximum: 100 },
            cursor: { type: "string" },
          },
        },
      },
    },
    async (request) => {
      const { patientId, limit, cursor } = request.query;
      const result = await service.listAlerts({ patientId, limit, cursor });
      return {
        alerts: result.items.map(alertToResponse),
        nextCursor: result.nextCursor,
      };
    }
  );
}

function parseTakenAt(value: number): Date {
  if (!Number.isFinite(value) || value <= 0) {
    throw new RequestError("takenAt is required");
  }
  const date = new Date(value * 1000);
  if (Number.isNaN(date.getTime())) {
    throw new RequestError("takenAt is required");
  }
  return date;
}
