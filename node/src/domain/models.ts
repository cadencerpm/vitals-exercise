export const SystolicUpperAlertThreshold = 180;
export const DiastolicUpperAlertThreshold = 120;

export type EventType = "VITAL_RECEIVED";

export const EventTypeVitalReceived: EventType = "VITAL_RECEIVED";

export type AlertStatus = "ACTIVE" | "RESOLVED" | "AUTO_RESOLVED";

export const AlertStatusActive: AlertStatus = "ACTIVE";
export const AlertStatusResolved: AlertStatus = "RESOLVED";
export const AlertStatusAutoResolved: AlertStatus = "AUTO_RESOLVED";

export interface Vital {
  id: string;
  patientId: string;
  systolic: number;
  diastolic: number;
  takenAt: Date;
  receivedAt: Date;
}

export interface Alert {
  id: string;
  vitalId: string;
  patientId: string;
  systolic: number;
  diastolic: number;
  takenAt: Date;
  receivedAt: Date;
  reason: string;
  status: AlertStatus;
  created: Date;
}

export interface Event {
  type: EventType;
  vital: Vital;
}

export type VitalInput = Partial<Vital> & { patientId: string };
export type AlertInput = Partial<Alert> & { patientId: string };

export function isAbnormal(vital: Vital): boolean {
  return (
    vital.systolic > SystolicUpperAlertThreshold ||
    vital.diastolic > DiastolicUpperAlertThreshold
  );
}

export function alertReason(vital: Vital): string {
  return `abnormal blood pressure ${vital.systolic}/${vital.diastolic}`;
}
