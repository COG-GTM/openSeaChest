import { useQuery } from '@tanstack/react-query';
import {
  getHealthCurrent,
  getHealthHistory,
  getFleetHealthSummary,
  getFarmLatest,
  getFarmHistory,
  getFleetReliability,
} from '../api/client';

export function useHealthCurrent(serial: string) {
  return useQuery({
    queryKey: ['health', 'current', serial],
    queryFn: () => getHealthCurrent(serial),
    enabled: !!serial,
  });
}

export function useHealthHistory(serial: string, from: string, to: string) {
  return useQuery({
    queryKey: ['health', 'history', serial, from, to],
    queryFn: () => getHealthHistory(serial, from, to),
    enabled: !!serial && !!from && !!to,
  });
}

export function useFleetHealthSummary() {
  return useQuery({
    queryKey: ['fleet', 'health', 'summary'],
    queryFn: getFleetHealthSummary,
  });
}

export function useFarmLatest(serial: string) {
  return useQuery({
    queryKey: ['farm', 'latest', serial],
    queryFn: () => getFarmLatest(serial),
    enabled: !!serial,
  });
}

export function useFarmHistory(serial: string, from: string, to: string) {
  return useQuery({
    queryKey: ['farm', 'history', serial, from, to],
    queryFn: () => getFarmHistory(serial, from, to),
    enabled: !!serial && !!from && !!to,
  });
}

export function useFleetReliability() {
  return useQuery({
    queryKey: ['fleet', 'reliability'],
    queryFn: getFleetReliability,
  });
}
