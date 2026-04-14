import type {
  DeviceInfo,
  Host,
  HealthSnapshot,
  FirmwarePolicy,
  ComplianceReport,
  ComplianceSummary,
  Alert,
  AlertRule,
  FarmSnapshot,
  FleetHealthSummary,
  FleetReliability,
  SmartStatus,
  ComplianceStatus,
  InterfaceType,
  AlertSeverity,
  AlertStatus,
} from '../api/types';

const MODELS = [
  { model: 'ST4000DX001-1CE168', fw: 'CC44', iface: 'SATA' as InterfaceType, capacity: 4e12, rpm: 5900, ff: '3.5' },
  { model: 'ST16000NM013J', fw: 'E82E', iface: 'SAS' as InterfaceType, capacity: 16e12, rpm: 7200, ff: '3.5' },
  { model: 'ST8000VN004-2M2101', fw: 'SC60', iface: 'SATA' as InterfaceType, capacity: 8e12, rpm: 7200, ff: '3.5' },
  { model: 'ST2000DM008-2UB102', fw: 'SN06', iface: 'SATA' as InterfaceType, capacity: 2e12, rpm: 7200, ff: '3.5' },
  { model: 'ST12000NM001J', fw: 'E82E', iface: 'SAS' as InterfaceType, capacity: 12e12, rpm: 7200, ff: '3.5' },
  { model: 'Samsung 990 PRO 2TB', fw: '4B2QJXD7', iface: 'NVMe' as InterfaceType, capacity: 2e12, rpm: 0, ff: 'M.2' },
  { model: 'WDC WD4003FFBX-68MU', fw: '83.00A83', iface: 'SATA' as InterfaceType, capacity: 4e12, rpm: 7200, ff: '3.5' },
  { model: 'INTEL SSDPE2KX040T8', fw: 'VDV10184', iface: 'NVMe' as InterfaceType, capacity: 4e12, rpm: 0, ff: '2.5' },
];

const HOSTNAMES = [
  'storage-node-01', 'storage-node-02', 'storage-node-03',
  'nas-primary', 'nas-backup',
  'compute-gpu-01', 'compute-gpu-02',
  'archive-cold-01', 'archive-cold-02', 'archive-cold-03',
];

function seededRandom(seed: number): () => number {
  let s = seed;
  return () => {
    s = (s * 16807 + 0) % 2147483647;
    return (s - 1) / 2147483646;
  };
}

const rand = seededRandom(42);

function randomElement<T>(arr: T[]): T {
  return arr[Math.floor(rand() * arr.length)];
}

function randomSerial(): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
  let s = '';
  for (let i = 0; i < 8; i++) {
    s += chars[Math.floor(rand() * chars.length)];
  }
  return s;
}

function randomSmartStatus(): SmartStatus {
  const r = rand();
  if (r < 0.7) return 'good';
  if (r < 0.85) return 'warning';
  if (r < 0.93) return 'tripped';
  return 'unknown';
}

function randomDate(daysAgo: number): string {
  const now = Date.now();
  const offset = Math.floor(rand() * daysAgo * 86400000);
  return new Date(now - offset).toISOString();
}

export const mockHosts: Host[] = HOSTNAMES.map((hostname, i) => ({
  id: `host-${i + 1}`,
  hostname,
  agent_id: `agent-${randomSerial()}`,
  device_count: 0,
  first_seen: randomDate(365),
  last_seen: randomDate(1),
  status: rand() > 0.1 ? 'online' : 'offline',
}));

export const mockDevices: DeviceInfo[] = Array.from({ length: 75 }, (_, i) => {
  const modelDef = randomElement(MODELS);
  const host = randomElement(mockHosts);
  const smartStatus = randomSmartStatus();
  const complianceR = rand();
  const complianceStatus: ComplianceStatus =
    complianceR < 0.65 ? 'compliant' : complianceR < 0.9 ? 'non_compliant' : 'no_policy';
  const serial = i === 0 ? 'ZQ3034X7R' : i === 1 ? 'ZRS004C2' : randomSerial();
  const fw = i === 0 ? 'CC44' : i === 1 ? 'E82E' : modelDef.fw;
  const temp = 20 + Math.floor(rand() * 30);
  const poh = Math.floor(rand() * 50000);

  host.device_count++;

  return {
    model: modelDef.model,
    serial_number: serial,
    firmware_revision: fw,
    world_wide_name: `500${randomSerial()}${randomSerial()}`,
    interface_type: modelDef.iface,
    capacity_bytes: modelDef.capacity,
    rotation_rate: modelDef.rpm,
    form_factor: modelDef.ff,
    device_handle: `/dev/sg${i}`,
    smart_status: smartStatus,
    compliance_status: complianceStatus,
    host_id: host.id,
    first_seen: randomDate(365),
    last_seen: randomDate(1),
    drive_capacity_tb: modelDef.capacity / 1e12,
    native_capacity_tb: modelDef.capacity / 1e12,
    temperature_current: temp,
    temperature_highest: temp + Math.floor(rand() * 15),
    temperature_lowest: temp - Math.floor(rand() * 10),
    power_on_hours: poh,
    logical_sector_size: 512,
    physical_sector_size: 4096,
    max_interface_speed: modelDef.iface === 'NVMe' ? '32.0 GT/s' : '6.0 Gb/s',
    negotiated_interface_speed: modelDef.iface === 'NVMe' ? '32.0 GT/s' : '6.0 Gb/s',
    encryption_support: rand() > 0.5 ? 'Supported' : 'Not Supported',
    firmware_download_support: 'Immediate, Segmented',
    annualized_workload_rate_tb_yr: Math.round(rand() * 100 * 100) / 100,
    total_bytes_read_gb: Math.round(rand() * 50000 * 100) / 100,
    total_bytes_written_gb: Math.round(rand() * 50000 * 100) / 100,
    cache_size_mib: modelDef.iface === 'NVMe' ? 0 : 256,
  };
});

function generateHealthTimeSeries(serial: string, days: number): HealthSnapshot[] {
  const snapshots: HealthSnapshot[] = [];
  const now = Date.now();
  const interval = 15 * 60 * 1000;
  const count = (days * 24 * 60) / 15;
  let baseTemp = 25 + Math.floor(rand() * 10);
  let poh = Math.floor(rand() * 10000);

  for (let i = 0; i < count; i++) {
    const ts = now - (count - i) * interval;
    const tempVariation = Math.floor(rand() * 6) - 3;
    baseTemp = Math.max(18, Math.min(55, baseTemp + tempVariation));
    poh += 0.25;

    snapshots.push({
      device_serial: serial,
      timestamp: new Date(ts).toISOString(),
      smart_status: 'good',
      temperature_c: baseTemp,
      power_on_hours: Math.floor(poh),
      power_cycle_count: Math.floor(rand() * 200),
      bytes_read: Math.floor(rand() * 1e12),
      bytes_written: Math.floor(rand() * 1e12),
      workload_rate_tb_yr: Math.round(rand() * 50 * 100) / 100,
      ssd_percentage_used: null,
      nvme_available_spare: null,
      nvme_media_errors: null,
    });
  }
  return snapshots;
}

const healthCache = new Map<string, HealthSnapshot[]>();

export function getHealthTimeSeriesForDevice(serial: string, days: number): HealthSnapshot[] {
  const cacheKey = `${serial}-${days}`;
  if (!healthCache.has(cacheKey)) {
    healthCache.set(cacheKey, generateHealthTimeSeries(serial, days));
  }
  return healthCache.get(cacheKey)!;
}

export const mockFirmwarePolicies: FirmwarePolicy[] = [
  { id: 'pol-1', model_pattern: 'ST4000DX001*', interface_type: 'SATA', required_firmware_revision: 'CC44', description: 'ST4000DX001 production firmware' },
  { id: 'pol-2', model_pattern: 'ST16000NM013J*', interface_type: 'SAS', required_firmware_revision: 'E82E', description: 'ST16000NM013J enterprise firmware' },
  { id: 'pol-3', model_pattern: 'ST8000VN004*', interface_type: 'SATA', required_firmware_revision: 'SC60', description: 'ST8000VN004 NAS firmware' },
  { id: 'pol-4', model_pattern: 'ST2000DM008*', interface_type: 'SATA', required_firmware_revision: 'SN06', description: 'ST2000DM008 desktop firmware' },
  { id: 'pol-5', model_pattern: 'ST12000NM001J*', interface_type: 'SAS', required_firmware_revision: 'E82E', description: 'ST12000NM001J enterprise firmware' },
];

export const mockComplianceReports: ComplianceReport[] = mockDevices.map((d) => ({
  device_serial: d.serial_number,
  model: d.model,
  current_firmware: d.firmware_revision,
  required_firmware: mockFirmwarePolicies.find((p) => d.model.startsWith(p.model_pattern.replace('*', '')))?.required_firmware_revision ?? 'N/A',
  compliance_status: d.compliance_status,
  policy_id: mockFirmwarePolicies.find((p) => d.model.startsWith(p.model_pattern.replace('*', '')))?.id ?? '',
}));

export const mockComplianceSummary: ComplianceSummary[] = (() => {
  const map = new Map<string, ComplianceSummary>();
  for (const d of mockDevices) {
    if (!map.has(d.model)) {
      map.set(d.model, { model: d.model, total_devices: 0, compliant: 0, non_compliant: 0, no_policy: 0 });
    }
    const s = map.get(d.model)!;
    s.total_devices++;
    if (d.compliance_status === 'compliant') s.compliant++;
    else if (d.compliance_status === 'non_compliant') s.non_compliant++;
    else s.no_policy++;
  }
  return Array.from(map.values());
})();

export const mockAlerts: Alert[] = Array.from({ length: 40 }, (_, i) => {
  const device = randomElement(mockDevices);
  const host = mockHosts.find((h) => h.id === device.host_id)!;
  const severities: AlertSeverity[] = ['critical', 'warning', 'info'];
  const severity = randomElement(severities);
  const statuses: AlertStatus[] = ['firing', 'acknowledged', 'resolved'];
  const status = randomElement(statuses);
  const metrics = ['temperature_c', 'power_on_hours', 'workload_rate_tb_yr', 'smart_status', 'unrecoverable_read_errors'];
  const metric = randomElement(metrics);

  return {
    id: `alert-${i + 1}`,
    rule_id: `rule-${Math.floor(rand() * 5) + 1}`,
    device_serial: device.serial_number,
    host: host.hostname,
    timestamp: randomDate(30),
    severity,
    metric,
    current_value: Math.round(rand() * 100),
    threshold_value: 50,
    message: `${metric} on ${device.serial_number} (${host.hostname}) exceeded threshold`,
    status,
  };
}).sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

export const mockAlertRules: AlertRule[] = [
  { id: 'rule-1', name: 'High Temperature', metric: 'temperature_c', condition_type: 'gt', threshold_value: 55, severity: 'warning', enabled: true },
  { id: 'rule-2', name: 'Critical Temperature', metric: 'temperature_c', condition_type: 'gt', threshold_value: 65, severity: 'critical', enabled: true },
  { id: 'rule-3', name: 'High Power-On Hours', metric: 'power_on_hours', condition_type: 'gt', threshold_value: 40000, severity: 'warning', enabled: true },
  { id: 'rule-4', name: 'Unrecoverable Read Errors', metric: 'unrecoverable_read_errors', condition_type: 'gt', threshold_value: 0, severity: 'critical', enabled: true },
  { id: 'rule-5', name: 'High Workload Rate', metric: 'workload_rate_tb_yr', condition_type: 'gt', threshold_value: 80, severity: 'info', enabled: true },
];

export function getMockFarmSnapshot(serial: string): FarmSnapshot {
  const device = mockDevices.find((d) => d.serial_number === serial);
  return {
    device_serial: serial,
    timestamp: new Date().toISOString(),
    power_on_hours: device?.power_on_hours ?? 1409,
    power_cycle_count: 77,
    unrecoverable_read_errors: serial === 'ZRS004C2' ? 30 : Math.floor(rand() * 5),
    unrecoverable_write_errors: 0,
    current_temperature_c: device?.temperature_current ?? 37,
    total_read_commands: 454267222,
    total_write_commands: 258095933,
  };
}

export const mockFleetHealthSummary: FleetHealthSummary = (() => {
  const staleThreshold = Date.now() - 24 * 60 * 60 * 1000;
  const staleCount = mockDevices.filter(
    (d) => new Date(d.last_seen).getTime() < staleThreshold,
  ).length;
  const summary: FleetHealthSummary = {
    good: 0,
    warning: 0,
    tripped: 0,
    unknown: 0,
    total: mockDevices.length,
    total_hosts: mockHosts.length,
    stale_devices: staleCount,
  };
  for (const d of mockDevices) {
    summary[d.smart_status]++;
  }
  return summary;
})();

export const mockFleetReliability: FleetReliability = {
  total_devices: mockDevices.length,
  total_unrecoverable_read_errors: 142,
  total_unrecoverable_write_errors: 3,
  total_mechanical_start_failures: 0,
  average_power_on_hours: Math.round(mockDevices.reduce((a, d) => a + d.power_on_hours, 0) / mockDevices.length),
  devices_by_interface: {
    SATA: mockDevices.filter((d) => d.interface_type === 'SATA').length,
    SAS: mockDevices.filter((d) => d.interface_type === 'SAS').length,
    NVMe: mockDevices.filter((d) => d.interface_type === 'NVMe').length,
  },
};
