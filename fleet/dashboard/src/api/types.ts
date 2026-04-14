export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export type InterfaceType = 'SATA' | 'SAS' | 'NVMe';
export type SmartStatus = 'good' | 'warning' | 'tripped' | 'unknown';
export type ComplianceStatus = 'compliant' | 'non_compliant' | 'no_policy';
export type AlertSeverity = 'critical' | 'warning' | 'info';
export type AlertStatus = 'firing' | 'acknowledged' | 'resolved';
export type ConditionType = 'gt' | 'lt' | 'gte' | 'lte' | 'eq';

export interface DeviceInfo {
  model: string;
  serial_number: string;
  firmware_revision: string;
  world_wide_name: string;
  interface_type: InterfaceType;
  capacity_bytes: number;
  rotation_rate: number;
  form_factor: string;
  device_handle: string;
  smart_status: SmartStatus;
  compliance_status: ComplianceStatus;
  host_id: string;
  first_seen: string;
  last_seen: string;
  drive_capacity_tb: number;
  native_capacity_tb: number;
  temperature_current: number;
  temperature_highest: number;
  temperature_lowest: number;
  power_on_hours: number;
  logical_sector_size: number;
  physical_sector_size: number;
  max_interface_speed: string;
  negotiated_interface_speed: string;
  encryption_support: string;
  firmware_download_support: string;
  annualized_workload_rate_tb_yr: number;
  total_bytes_read_gb: number;
  total_bytes_written_gb: number;
  cache_size_mib: number;
}

export interface Host {
  id: string;
  hostname: string;
  agent_id: string;
  device_count: number;
  first_seen: string;
  last_seen: string;
  status: string;
}

export interface HealthSnapshot {
  device_serial: string;
  timestamp: string;
  smart_status: SmartStatus;
  temperature_c: number;
  power_on_hours: number;
  power_cycle_count: number;
  bytes_read: number;
  bytes_written: number;
  workload_rate_tb_yr: number;
  ssd_percentage_used: number | null;
  nvme_available_spare: number | null;
  nvme_media_errors: number | null;
}

export interface FirmwarePolicy {
  id: string;
  model_pattern: string;
  interface_type: InterfaceType;
  required_firmware_revision: string;
  description: string;
}

export interface ComplianceReport {
  device_serial: string;
  model: string;
  current_firmware: string;
  required_firmware: string;
  compliance_status: ComplianceStatus;
  policy_id: string;
}

export interface ComplianceSummary {
  model: string;
  total_devices: number;
  compliant: number;
  non_compliant: number;
  no_policy: number;
}

export interface Alert {
  id: string;
  rule_id: string;
  device_serial: string;
  host: string;
  timestamp: string;
  severity: AlertSeverity;
  metric: string;
  current_value: number;
  threshold_value: number;
  message: string;
  status: AlertStatus;
}

export interface AlertRule {
  id: string;
  name: string;
  metric: string;
  condition_type: ConditionType;
  threshold_value: number;
  severity: AlertSeverity;
  enabled: boolean;
}

export interface FarmSnapshot {
  device_serial: string;
  timestamp: string;
  power_on_hours: number;
  power_cycle_count: number;
  unrecoverable_read_errors: number;
  unrecoverable_write_errors: number;
  current_temperature_c: number;
  total_read_commands: number;
  total_write_commands: number;
}

export interface FleetHealthSummary {
  good: number;
  warning: number;
  tripped: number;
  unknown: number;
  total: number;
}

export interface FleetReliability {
  total_devices: number;
  total_unrecoverable_read_errors: number;
  total_unrecoverable_write_errors: number;
  total_mechanical_start_failures: number;
  average_power_on_hours: number;
  devices_by_interface: Record<InterfaceType, number>;
}

export interface DeviceQueryParams {
  model?: string;
  firmware?: string;
  interface_type?: InterfaceType;
  host?: string;
  health_status?: SmartStatus;
  compliance_status?: ComplianceStatus;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface AlertQueryParams {
  severity?: AlertSeverity;
  device?: string;
  host?: string;
  status?: AlertStatus;
  from?: string;
  to?: string;
  page?: number;
  page_size?: number;
}

export interface ComplianceQueryParams {
  model?: string;
  compliance_status?: ComplianceStatus;
  page?: number;
  page_size?: number;
}
