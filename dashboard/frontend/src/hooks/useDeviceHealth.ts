import { useEffect, useState } from "react";
import type { HealthSnapshot } from "../api/types";
import { getDeviceHealthHistory } from "../api/client";

interface UseDeviceHealthResult {
  history: HealthSnapshot[];
  loading: boolean;
  error: string | null;
}

export function useDeviceHealth(serial: string): UseDeviceHealthResult {
  const [history, setHistory] = useState<HealthSnapshot[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getDeviceHealthHistory(serial)
      .then(setHistory)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [serial]);

  return { history, loading, error };
}
