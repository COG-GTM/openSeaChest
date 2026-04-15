/** Core device representation from the fleet inventory. */
export interface Device {
  serial_number: string;
  model: string;
  device_type: "HDD" | "SSD" | "NVMe";
  host: string;
  firmware_version: string;
  smart_status: "PASSED" | "FAILED" | "UNKNOWN";
  temperature_c: number;
  power_on_hours: number;
  wwn: string;
  capacity_gb: number;
  interface_type: string;
  features: string[];
}

/** A point-in-time health snapshot for a device. */
export interface HealthSnapshot {
  id: string;
  device_serial: string;
  timestamp: string;
  temperature_c: number;
  power_on_hours: number;
  smart_status: "PASSED" | "FAILED" | "UNKNOWN";
  reallocated_sectors: number;
  pending_sectors: number;
  crc_errors: number;
}

/** Result of a firmware compliance check. */
export interface FirmwareComplianceResult {
  compliant: Device[];
  non_compliant: Device[];
}

/** Alert rule types supported by the backend. */
export type AlertRuleType =
  | "smart_tripped"
  | "temperature_threshold"
  | "firmware_not_approved"
  | "poh_threshold";

/** A user-defined alert rule. */
export interface AlertRule {
  id: string;
  name: string;
  rule_type: AlertRuleType;
  threshold_value: number | null;
  enabled: boolean;
  created_at: string;
}

/** A fired alert instance. */
export interface FiredAlert {
  id: string;
  device_serial: string;
  rule_id: string;
  rule_name: string;
  fired_at: string;
  message: string;
  resolved: boolean;
}
