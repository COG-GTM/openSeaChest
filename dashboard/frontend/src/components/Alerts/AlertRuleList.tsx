import { useState } from "react";
import type { AlertRule, AlertRuleType } from "../../api/types";
import AlertRuleForm from "./AlertRuleForm";

interface AlertRuleListProps {
  rules: AlertRule[];
  onEdit: (id: string, rule: Partial<AlertRule>) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onAdd: (rule: {
    name: string;
    rule_type: AlertRuleType;
    threshold_value: number | null;
    enabled: boolean;
  }) => Promise<void>;
}

function formatRuleType(type: AlertRuleType): string {
  const map: Record<AlertRuleType, string> = {
    smart_tripped: "SMART Tripped",
    temperature_threshold: "Temperature Threshold",
    firmware_not_approved: "Firmware Not Approved",
    poh_threshold: "POH Threshold",
  };
  return map[type];
}

export default function AlertRuleList({
  rules,
  onEdit,
  onDelete,
  onAdd,
}: AlertRuleListProps) {
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  const handleAdd = async (rule: {
    name: string;
    rule_type: AlertRuleType;
    threshold_value: number | null;
    enabled: boolean;
  }) => {
    await onAdd(rule);
    setShowForm(false);
  };

  const handleEdit = async (rule: {
    name: string;
    rule_type: AlertRuleType;
    threshold_value: number | null;
    enabled: boolean;
  }) => {
    if (editingId) {
      await onEdit(editingId, rule);
      setEditingId(null);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold text-gray-900">Alert Rules</h3>
        {!showForm && (
          <button
            onClick={() => setShowForm(true)}
            className="px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors"
          >
            + New Rule
          </button>
        )}
      </div>

      {showForm && (
        <AlertRuleForm
          onSubmit={handleAdd}
          onCancel={() => setShowForm(false)}
        />
      )}

      <div className="space-y-2">
        {rules.map((rule) =>
          editingId === rule.id ? (
            <AlertRuleForm
              key={rule.id}
              onSubmit={handleEdit}
              onCancel={() => setEditingId(null)}
              initialValues={{
                name: rule.name,
                rule_type: rule.rule_type,
                threshold_value: rule.threshold_value,
                enabled: rule.enabled,
              }}
            />
          ) : (
            <div
              key={rule.id}
              className="flex items-center justify-between bg-white border border-gray-200 rounded-lg p-4"
            >
              <div className="flex-1">
                <div className="flex items-center gap-3">
                  <span className="font-medium text-gray-900">{rule.name}</span>
                  <span
                    className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                      rule.enabled
                        ? "bg-green-100 text-green-800"
                        : "bg-gray-100 text-gray-600"
                    }`}
                  >
                    {rule.enabled ? "Enabled" : "Disabled"}
                  </span>
                </div>
                <div className="text-sm text-gray-500 mt-1">
                  {formatRuleType(rule.rule_type)}
                  {rule.threshold_value !== null &&
                    ` — threshold: ${rule.threshold_value}`}
                </div>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setEditingId(rule.id)}
                  className="px-3 py-1 text-sm text-blue-600 hover:text-blue-800 font-medium"
                >
                  Edit
                </button>
                <button
                  onClick={() => onDelete(rule.id)}
                  className="px-3 py-1 text-sm text-red-600 hover:text-red-800 font-medium"
                >
                  Delete
                </button>
              </div>
            </div>
          )
        )}
        {rules.length === 0 && (
          <p className="text-center py-4 text-gray-500 text-sm">
            No alert rules configured.
          </p>
        )}
      </div>
    </div>
  );
}
