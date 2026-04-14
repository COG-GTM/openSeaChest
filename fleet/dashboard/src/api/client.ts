import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import type {
  PaginatedResponse,
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
  DeviceQueryParams,
  AlertQueryParams,
  ComplianceQueryParams,
} from './types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

const client = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('fleet_auth_token');
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

client.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('fleet_auth_token');
    }
    return Promise.reject(error);
  },
);

export async function getDevices(
  params?: DeviceQueryParams,
): Promise<PaginatedResponse<DeviceInfo>> {
  const { data } = await client.get('/v1/devices', { params });
  return data;
}

export async function getDevice(serial: string): Promise<DeviceInfo> {
  const { data } = await client.get(`/v1/devices/${serial}`);
  return data;
}

export async function getHosts(): Promise<Host[]> {
  const { data } = await client.get('/v1/hosts');
  return data;
}

export async function getHostDevices(hostId: string): Promise<DeviceInfo[]> {
  const { data } = await client.get(`/v1/hosts/${hostId}/devices`);
  return data;
}

export async function getHealthCurrent(
  serial: string,
): Promise<HealthSnapshot> {
  const { data } = await client.get(`/v1/devices/${serial}/health/current`);
  return data;
}

export async function getHealthHistory(
  serial: string,
  from: string,
  to: string,
): Promise<HealthSnapshot[]> {
  const { data } = await client.get(`/v1/devices/${serial}/health/history`, {
    params: { from, to },
  });
  return data;
}

export async function getFleetHealthSummary(): Promise<FleetHealthSummary> {
  const { data } = await client.get('/v1/fleet/health/summary');
  return data;
}

export async function getFirmwarePolicies(): Promise<FirmwarePolicy[]> {
  const { data } = await client.get('/v1/firmware-policies');
  return data;
}

export async function getComplianceReport(
  params?: ComplianceQueryParams,
): Promise<PaginatedResponse<ComplianceReport>> {
  const { data } = await client.get('/v1/firmware-compliance', { params });
  return data;
}

export async function getComplianceSummary(): Promise<ComplianceSummary[]> {
  const { data } = await client.get('/v1/firmware-compliance/summary');
  return data;
}

export async function getAlerts(
  params?: AlertQueryParams,
): Promise<PaginatedResponse<Alert>> {
  const { data } = await client.get('/v1/alerts', { params });
  return data;
}

export async function acknowledgeAlert(
  id: string,
  annotation: string,
): Promise<Alert> {
  const { data } = await client.put(`/v1/alerts/${id}/acknowledge`, {
    annotation,
  });
  return data;
}

export async function getAlertRules(): Promise<AlertRule[]> {
  const { data } = await client.get('/v1/alert-rules');
  return data;
}

export async function getFarmLatest(serial: string): Promise<FarmSnapshot> {
  const { data } = await client.get(`/v1/devices/${serial}/farm/latest`);
  return data;
}

export async function getFarmHistory(
  serial: string,
  from: string,
  to: string,
): Promise<FarmSnapshot[]> {
  const { data } = await client.get(`/v1/devices/${serial}/farm/history`, {
    params: { from, to },
  });
  return data;
}

export async function getFleetReliability(): Promise<FleetReliability> {
  const { data } = await client.get('/v1/fleet/reliability');
  return data;
}
