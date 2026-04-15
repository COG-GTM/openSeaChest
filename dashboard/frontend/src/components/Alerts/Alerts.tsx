import { useState } from "react";
import { useAlerts } from "../../hooks/useAlerts";
import FiredAlertsList from "./FiredAlertsList";
import AlertRuleList from "./AlertRuleList";

type Tab = "fired" | "rules";

export default function Alerts() {
  const [activeTab, setActiveTab] = useState<Tab>("fired");
  const {
    firedAlerts,
    alertRules,
    loading,
    error,
    addRule,
    editRule,
    removeRule,
  } = useAlerts();

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4 text-red-700">
        Error loading alerts: {error}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">Alerts</h2>

      <div className="border-b border-gray-200">
        <nav className="flex space-x-8">
          <button
            onClick={() => setActiveTab("fired")}
            className={`py-3 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === "fired"
                ? "border-blue-500 text-blue-600"
                : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
            }`}
          >
            Fired Alerts
            {firedAlerts.length > 0 && (
              <span className="ml-2 bg-red-100 text-red-600 text-xs rounded-full px-2 py-0.5">
                {firedAlerts.filter((a) => !a.resolved).length}
              </span>
            )}
          </button>
          <button
            onClick={() => setActiveTab("rules")}
            className={`py-3 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === "rules"
                ? "border-blue-500 text-blue-600"
                : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
            }`}
          >
            Alert Rules
            <span className="ml-2 bg-gray-100 text-gray-600 text-xs rounded-full px-2 py-0.5">
              {alertRules.length}
            </span>
          </button>
        </nav>
      </div>

      <div className="bg-white rounded-lg shadow border border-gray-200 overflow-hidden">
        {activeTab === "fired" ? (
          <FiredAlertsList alerts={firedAlerts} />
        ) : (
          <div className="p-6">
            <AlertRuleList
              rules={alertRules}
              onEdit={editRule}
              onDelete={removeRule}
              onAdd={addRule}
            />
          </div>
        )}
      </div>
    </div>
  );
}
