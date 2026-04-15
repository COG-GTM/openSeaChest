import { useEffect, useState, useCallback } from "react";
import type { FiredAlert, AlertRule, AlertRuleType } from "../api/types";
import {
  getFiredAlerts,
  getAlertRules,
  createAlertRule,
  updateAlertRule,
  deleteAlertRule,
} from "../api/client";

interface UseAlertsResult {
  firedAlerts: FiredAlert[];
  alertRules: AlertRule[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
  addRule: (rule: {
    name: string;
    rule_type: AlertRuleType;
    threshold_value: number | null;
    enabled: boolean;
  }) => Promise<void>;
  editRule: (id: string, rule: Partial<AlertRule>) => Promise<void>;
  removeRule: (id: string) => Promise<void>;
}

export function useAlerts(): UseAlertsResult {
  const [firedAlerts, setFiredAlerts] = useState<FiredAlert[]>([]);
  const [alertRules, setAlertRules] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAll = useCallback(() => {
    setLoading(true);
    setError(null);
    Promise.all([getFiredAlerts(), getAlertRules()])
      .then(([alerts, rules]) => {
        setFiredAlerts(alerts);
        setAlertRules(rules);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const addRule = useCallback(
    async (rule: {
      name: string;
      rule_type: AlertRuleType;
      threshold_value: number | null;
      enabled: boolean;
    }) => {
      const newRule = await createAlertRule(rule);
      setAlertRules((prev) => [...prev, newRule]);
    },
    []
  );

  const editRule = useCallback(
    async (id: string, updates: Partial<AlertRule>) => {
      const updated = await updateAlertRule(id, updates);
      setAlertRules((prev) => prev.map((r) => (r.id === id ? updated : r)));
    },
    []
  );

  const removeRule = useCallback(async (id: string) => {
    await deleteAlertRule(id);
    setAlertRules((prev) => prev.filter((r) => r.id !== id));
  }, []);

  return {
    firedAlerts,
    alertRules,
    loading,
    error,
    refetch: fetchAll,
    addRule,
    editRule,
    removeRule,
  };
}
