import { http, HttpResponse } from 'msw';
import type {
  DeviceQueryParams,
  AlertQueryParams,
  ComplianceQueryParams,
  DeviceInfo,
  Alert,
  ComplianceReport,
} from '../api/types';
import {
  mockDevices,
  mockHosts,
  mockFirmwarePolicies,
  mockComplianceReports,
  mockComplianceSummary,
  mockAlerts,
  mockAlertRules,
  mockFleetHealthSummary,
  mockFleetReliability,
  getMockFarmSnapshot,
  getHealthTimeSeriesForDevice,
} from './data';

function paginate<T>(items: T[], page: number, pageSize: number) {
  const start = (page - 1) * pageSize;
  return {
    items: items.slice(start, start + pageSize),
    total: items.length,
    page,
    page_size: pageSize,
  };
}

function filterDevices(devices: DeviceInfo[], params: URLSearchParams): DeviceInfo[] {
  let result = [...devices];

  const model = params.get('model');
  if (model) result = result.filter((d) => d.model.toLowerCase().includes(model.toLowerCase()));

  const firmware = params.get('firmware');
  if (firmware) result = result.filter((d) => d.firmware_revision === firmware);

  const ifaceType = params.get('interface_type') as DeviceQueryParams['interface_type'];
  if (ifaceType) result = result.filter((d) => d.interface_type === ifaceType);

  const host = params.get('host');
  if (host) {
    const matchHost = mockHosts.find((h) => h.hostname.toLowerCase().includes(host.toLowerCase()));
    if (matchHost) result = result.filter((d) => d.host_id === matchHost.id);
    else result = [];
  }

  const healthStatus = params.get('health_status') as DeviceQueryParams['health_status'];
  if (healthStatus) result = result.filter((d) => d.smart_status === healthStatus);

  const complianceStatus = params.get('compliance_status') as DeviceQueryParams['compliance_status'];
  if (complianceStatus) result = result.filter((d) => d.compliance_status === complianceStatus);

  const sortBy = params.get('sort_by');
  const sortOrder = params.get('sort_order') || 'asc';
  if (sortBy) {
    result.sort((a, b) => {
      const aVal = a[sortBy as keyof DeviceInfo];
      const bVal = b[sortBy as keyof DeviceInfo];
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortOrder === 'asc' ? aVal - bVal : bVal - aVal;
      }
      const aStr = String(aVal);
      const bStr = String(bVal);
      return sortOrder === 'asc' ? aStr.localeCompare(bStr) : bStr.localeCompare(aStr);
    });
  }

  return result;
}

function filterAlerts(alerts: Alert[], params: URLSearchParams): Alert[] {
  let result = [...alerts];

  const severity = params.get('severity') as AlertQueryParams['severity'];
  if (severity) result = result.filter((a) => a.severity === severity);

  const device = params.get('device');
  if (device) result = result.filter((a) => a.device_serial.toLowerCase().includes(device.toLowerCase()));

  const host = params.get('host');
  if (host) result = result.filter((a) => a.host.toLowerCase().includes(host.toLowerCase()));

  const status = params.get('status') as AlertQueryParams['status'];
  if (status) result = result.filter((a) => a.status === status);

  return result;
}

function filterCompliance(reports: ComplianceReport[], params: URLSearchParams): ComplianceReport[] {
  let result = [...reports];

  const model = params.get('model');
  if (model) result = result.filter((r) => r.model.toLowerCase().includes(model.toLowerCase()));

  const complianceStatus = params.get('compliance_status') as ComplianceQueryParams['compliance_status'];
  if (complianceStatus) result = result.filter((r) => r.compliance_status === complianceStatus);

  return result;
}

export const handlers = [
  http.get('/api/v1/devices', ({ request }) => {
    const url = new URL(request.url);
    const page = parseInt(url.searchParams.get('page') || '1');
    const pageSize = parseInt(url.searchParams.get('page_size') || '25');
    const filtered = filterDevices(mockDevices, url.searchParams);
    return HttpResponse.json(paginate(filtered, page, pageSize));
  }),

  http.get('/api/v1/devices/:serial', ({ params }) => {
    const device = mockDevices.find((d) => d.serial_number === params.serial);
    if (!device) return new HttpResponse(null, { status: 404 });
    return HttpResponse.json(device);
  }),

  http.get('/api/v1/hosts', () => {
    return HttpResponse.json(mockHosts);
  }),

  http.get('/api/v1/hosts/:hostId/devices', ({ params }) => {
    const devices = mockDevices.filter((d) => d.host_id === params.hostId);
    return HttpResponse.json(devices);
  }),

  http.get('/api/v1/devices/:serial/health/current', ({ params }) => {
    const series = getHealthTimeSeriesForDevice(params.serial as string, 1);
    const latest = series[series.length - 1];
    if (!latest) return new HttpResponse(null, { status: 404 });
    return HttpResponse.json(latest);
  }),

  http.get('/api/v1/devices/:serial/health/history', ({ request, params }) => {
    const url = new URL(request.url);
    const from = url.searchParams.get('from');
    const to = url.searchParams.get('to');
    const fromDate = from ? new Date(from).getTime() : Date.now() - 30 * 86400000;
    const toDate = to ? new Date(to).getTime() : Date.now();
    const days = Math.ceil((toDate - fromDate) / 86400000);
    const series = getHealthTimeSeriesForDevice(params.serial as string, days);
    const filtered = series.filter((s) => {
      const ts = new Date(s.timestamp).getTime();
      return ts >= fromDate && ts <= toDate;
    });
    // Downsample for large ranges
    const maxPoints = 500;
    if (filtered.length > maxPoints) {
      const step = Math.ceil(filtered.length / maxPoints);
      return HttpResponse.json(filtered.filter((_, i) => i % step === 0));
    }
    return HttpResponse.json(filtered);
  }),

  http.get('/api/v1/fleet/health/summary', () => {
    return HttpResponse.json(mockFleetHealthSummary);
  }),

  http.get('/api/v1/firmware-policies', () => {
    return HttpResponse.json(mockFirmwarePolicies);
  }),

  http.get('/api/v1/firmware-compliance', ({ request }) => {
    const url = new URL(request.url);
    const page = parseInt(url.searchParams.get('page') || '1');
    const pageSize = parseInt(url.searchParams.get('page_size') || '25');
    const filtered = filterCompliance(mockComplianceReports, url.searchParams);
    return HttpResponse.json(paginate(filtered, page, pageSize));
  }),

  http.get('/api/v1/firmware-compliance/summary', () => {
    return HttpResponse.json(mockComplianceSummary);
  }),

  http.get('/api/v1/alerts', ({ request }) => {
    const url = new URL(request.url);
    const page = parseInt(url.searchParams.get('page') || '1');
    const pageSize = parseInt(url.searchParams.get('page_size') || '25');
    const filtered = filterAlerts(mockAlerts, url.searchParams);
    return HttpResponse.json(paginate(filtered, page, pageSize));
  }),

  http.put('/api/v1/alerts/:id/acknowledge', async ({ params, request }) => {
    const body = (await request.json()) as { annotation: string };
    const alert = mockAlerts.find((a) => a.id === params.id);
    if (!alert) return new HttpResponse(null, { status: 404 });
    alert.status = 'acknowledged';
    return HttpResponse.json({ ...alert, annotation: body.annotation });
  }),

  http.get('/api/v1/alert-rules', () => {
    return HttpResponse.json(mockAlertRules);
  }),

  http.get('/api/v1/devices/:serial/farm/latest', ({ params }) => {
    return HttpResponse.json(getMockFarmSnapshot(params.serial as string));
  }),

  http.get('/api/v1/devices/:serial/farm/history', ({ params }) => {
    const snap = getMockFarmSnapshot(params.serial as string);
    const history = Array.from({ length: 30 }, (_, i) => ({
      ...snap,
      timestamp: new Date(Date.now() - (30 - i) * 86400000).toISOString(),
      power_on_hours: snap.power_on_hours - (30 - i) * 24,
      current_temperature_c: snap.current_temperature_c + Math.floor(Math.random() * 4) - 2,
    }));
    return HttpResponse.json(history);
  }),

  http.get('/api/v1/fleet/reliability', () => {
    return HttpResponse.json(mockFleetReliability);
  }),
];
