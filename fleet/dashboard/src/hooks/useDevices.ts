import { useQuery } from '@tanstack/react-query';
import { getDevices, getDevice } from '../api/client';
import type { DeviceQueryParams } from '../api/types';

export function useDevices(params?: DeviceQueryParams) {
  return useQuery({
    queryKey: ['devices', params],
    queryFn: () => getDevices(params),
    placeholderData: (prev) => prev,
  });
}

export function useDevice(serial: string) {
  return useQuery({
    queryKey: ['device', serial],
    queryFn: () => getDevice(serial),
    enabled: !!serial,
  });
}
