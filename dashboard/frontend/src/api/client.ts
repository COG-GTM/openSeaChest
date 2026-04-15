import type {
  Device,
  HealthSnapshot,
  FirmwareComplianceResult,
  FiredAlert,
  AlertRule,
} from "./types";

const BASE_URL = "/api/v1";

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!response.ok) {
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }
  if (response.status === 204) {
    return undefined as unknown as T;
  }
  return response.json() as Promise<T>;
}

/** Fetch all devices in the fleet. */
export function getDevices(): Promise<Device[]> {
  return fetchJSON<Device[]>(`${BASE_URL}/devices`);
}

/** Fetch a single device by serial number. */
export function getDevice(serial: string): Promise<Device> {
  return fetchJSON<Device>(`${BASE_URL}/devices/${serial}`);
}

/** Fetch health history for a device. */
export function getDeviceHealthHistory(
  serial: string
): Promise<HealthSnapshot[]> {
  return fetchJSON<HealthSnapshot[]>(
    `${BASE_URL}/devices/${serial}/health/history`
  );
}

/** Check firmware compliance against a target version. */
export function getFirmwareCompliance(
  targetFirmware: string
): Promise<FirmwareComplianceResult> {
  return fetchJSON<FirmwareComplianceResult>(
    `${BASE_URL}/firmware/compliance?target_firmware=${encodeURIComponent(targetFirmware)}`
  );
}

/** Fetch all fired alerts. */
export function getFiredAlerts(): Promise<FiredAlert[]> {
  return fetchJSON<FiredAlert[]>(`${BASE_URL}/alerts`);
}

/** Fetch all alert rules. */
export function getAlertRules(): Promise<AlertRule[]> {
  return fetchJSON<AlertRule[]>(`${BASE_URL}/alerts/rules`);
}

/** Create a new alert rule. */
export function createAlertRule(
  rule: Omit<AlertRule, "id" | "created_at">
): Promise<AlertRule> {
  return fetchJSON<AlertRule>(`${BASE_URL}/alerts/rules`, {
    method: "POST",
    body: JSON.stringify(rule),
  });
}

/** Update an existing alert rule. */
export function updateAlertRule(
  id: string,
  rule: Partial<AlertRule>
): Promise<AlertRule> {
  return fetchJSON<AlertRule>(`${BASE_URL}/alerts/rules/${id}`, {
    method: "PUT",
    body: JSON.stringify(rule),
  });
}

/** Delete an alert rule. */
export function deleteAlertRule(id: string): Promise<void> {
  return fetchJSON<void>(`${BASE_URL}/alerts/rules/${id}`, {
    method: "DELETE",
  });
}
