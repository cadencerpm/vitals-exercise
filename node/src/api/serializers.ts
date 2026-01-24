import { Message } from "../domain/messageQueue";
import { Alert, Vital } from "../domain/models";

export function vitalToResponse(vital: Vital) {
  return {
    id: vital.id,
    patientId: vital.patientId,
    systolic: vital.systolic,
    diastolic: vital.diastolic,
    takenAt: toUnixSeconds(vital.takenAt),
    receivedAt: toUnixSeconds(vital.receivedAt),
  };
}

export function alertToResponse(alert: Alert) {
  return {
    id: alert.id,
    vital: {
      id: alert.vitalId,
      patientId: alert.patientId,
      systolic: alert.systolic,
      diastolic: alert.diastolic,
      takenAt: toUnixSeconds(alert.takenAt),
      receivedAt: toUnixSeconds(alert.receivedAt),
    },
    reason: alert.reason,
    createdAt: toUnixSeconds(alert.created),
    status: alert.status,
  };
}

export function messageToResponse(message: Message) {
  return {
    id: message.id,
    patientId: message.patientId,
    content: message.content,
    status: message.status,
    queuedAt: toUnixSeconds(message.queuedAt),
    sentAt: message.sentAt ? toUnixSeconds(message.sentAt) : 0,
  };
}

function toUnixSeconds(value: Date): number {
  return Math.floor(value.getTime() / 1000);
}
