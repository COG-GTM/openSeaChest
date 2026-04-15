import { useEffect, useState, useCallback } from "react";
import type { Device } from "../api/types";
import { getDevices, getDevice } from "../api/client";

interface UseDevicesResult {
  devices: Device[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export function useDevices(): UseDevicesResult {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDevices = useCallback(() => {
    setLoading(true);
    setError(null);
    getDevices()
      .then(setDevices)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchDevices();
  }, [fetchDevices]);

  return { devices, loading, error, refetch: fetchDevices };
}

interface UseDeviceResult {
  device: Device | null;
  loading: boolean;
  error: string | null;
}

export function useDevice(serial: string): UseDeviceResult {
  const [device, setDevice] = useState<Device | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getDevice(serial)
      .then(setDevice)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [serial]);

  return { device, loading, error };
}
