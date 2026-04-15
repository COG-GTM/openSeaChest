import { useState } from "react";
import type { AlertRuleType } from "../../api/types";

interface AlertRuleFormProps {
  onSubmit: (rule: {
    name: string;
    rule_type: AlertRuleType;
    threshold_value: number | null;
    enabled: boolean;
  }) => Promise<void>;
  onCancel: () => void;
  initialValues?: {
    name: string;
    rule_type: AlertRuleType;
    threshold_value: number | null;
    enabled: boolean;
  };
}

const RULE_TYPES: { value: AlertRuleType; label: string }[] = [
  { value: "smart_tripped", label: "SMART Tripped" },
  { value: "temperature_threshold", label: "Temperature Threshold" },
  { value: "firmware_not_approved", label: "Firmware Not Approved" },
  { value: "poh_threshold", label: "POH Threshold" },
];

const THRESHOLD_RULE_TYPES: AlertRuleType[] = [
  "temperature_threshold",
  "poh_threshold",
];

export default function AlertRuleForm({
  onSubmit,
  onCancel,
  initialValues,
}: AlertRuleFormProps) {
  const [name, setName] = useState(initialValues?.name ?? "");
  const [ruleType, setRuleType] = useState<AlertRuleType>(
    initialValues?.rule_type ?? "smart_tripped"
  );
  const [thresholdValue, setThresholdValue] = useState<string>(
    initialValues?.threshold_value?.toString() ?? ""
  );
  const [enabled, setEnabled] = useState(initialValues?.enabled ?? true);
  const [submitting, setSubmitting] = useState(false);

  const needsThreshold = THRESHOLD_RULE_TYPES.includes(ruleType);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({
        name,
        rule_type: ruleType,
        threshold_value: needsThreshold ? Number(thresholdValue) : null,
        enabled,
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="bg-gray-50 border border-gray-200 rounded-lg p-4 space-y-4"
    >
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Rule Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
            placeholder="e.g. High Temperature Alert"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Rule Type
          </label>
          <select
            value={ruleType}
            onChange={(e) => setRuleType(e.target.value as AlertRuleType)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
          >
            {RULE_TYPES.map((rt) => (
              <option key={rt.value} value={rt.value}>
                {rt.label}
              </option>
            ))}
          </select>
        </div>
        {needsThreshold && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Threshold Value
            </label>
            <input
              type="number"
              value={thresholdValue}
              onChange={(e) => setThresholdValue(e.target.value)}
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
              placeholder={
                ruleType === "temperature_threshold"
                  ? "e.g. 50 (°C)"
                  : "e.g. 40000 (hours)"
              }
            />
          </div>
        )}
        <div className="flex items-center pt-6">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={enabled}
              onChange={(e) => setEnabled(e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
            />
            <span className="text-sm font-medium text-gray-700">Enabled</span>
          </label>
        </div>
      </div>
      <div className="flex gap-2">
        <button
          type="submit"
          disabled={submitting || !name.trim()}
          className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {submitting ? "Saving..." : initialValues ? "Update Rule" : "Create Rule"}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 bg-white text-gray-700 text-sm font-medium rounded-md border border-gray-300 hover:bg-gray-50 transition-colors"
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
