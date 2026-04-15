import { useState, useCallback } from "react";
import type { FirmwareComplianceResult } from "../api/types";
import { getFirmwareCompliance } from "../api/client";

interface UseFirmwareComplianceResult {
  result: FirmwareComplianceResult | null;
  loading: boolean;
  error: string | null;
  checkCompliance: (targetFirmware: string) => void;
}

export function useFirmwareCompliance(): UseFirmwareComplianceResult {
  const [result, setResult] = useState<FirmwareComplianceResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const checkCompliance = useCallback((targetFirmware: string) => {
    setLoading(true);
    setError(null);
    setResult(null);
    getFirmwareCompliance(targetFirmware)
      .then(setResult)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return { result, loading, error, checkCompliance };
}
