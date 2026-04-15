import type { HealthSnapshot } from "../api/types";
import { mockDevices } from "./devices";

function generateHealthHistory(
  serial: string,
  count: number
): HealthSnapshot[] {
  const now = Date.now();
  const oneHour = 3600_000;
  const snapshots: HealthSnapshot[] = [];
  let basePoh = Math.floor(Math.random() * 40000) + 5000;
  let baseTemp = Math.floor(Math.random() * 10) + 30;

  for (let i = count - 1; i >= 0; i--) {
    const timestamp = new Date(now - i * oneHour).toISOString();
    const tempVariation = Math.floor(Math.random() * 6) - 3;
    snapshots.push({
      id: `hs-${serial}-${i}`,
      device_serial: serial,
      timestamp,
      temperature_c: baseTemp + tempVariation,
      power_on_hours: basePoh + (count - i),
      smart_status: Math.random() > 0.95 ? "FAILED" : "PASSED",
      reallocated_sectors: Math.floor(Math.random() * 5),
      pending_sectors: Math.floor(Math.random() * 3),
      crc_errors: Math.floor(Math.random() * 2),
    });
    baseTemp += Math.random() > 0.5 ? 1 : -1;
    baseTemp = Math.max(25, Math.min(55, baseTemp));
    basePoh += 1;
  }

  return snapshots;
}

/** Pre-generated health data keyed by serial number. */
export const mockHealthData: Record<string, HealthSnapshot[]> = {};

// Generate 48 hours of data for each device
for (const device of mockDevices) {
  mockHealthData[device.serial_number] = generateHealthHistory(
    device.serial_number,
    48
  );
}
