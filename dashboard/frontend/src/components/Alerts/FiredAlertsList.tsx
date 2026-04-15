import { useNavigate } from "react-router-dom";
import type { FiredAlert } from "../../api/types";

interface FiredAlertsListProps {
  alerts: FiredAlert[];
}

export default function FiredAlertsList({ alerts }: FiredAlertsListProps) {
  const navigate = useNavigate();

  if (alerts.length === 0) {
    return (
      <p className="text-center py-8 text-gray-500 text-sm">
        No fired alerts.
      </p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {["Device", "Rule", "Fired At", "Message", "Status"].map((h) => (
              <th
                key={h}
                className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {alerts.map((alert) => (
            <tr key={alert.id} className="hover:bg-gray-50">
              <td
                className="px-4 py-3 text-sm font-mono text-blue-600 hover:text-blue-800 cursor-pointer"
                onClick={() =>
                  navigate(`/devices/${alert.device_serial}`)
                }
              >
                {alert.device_serial}
              </td>
              <td className="px-4 py-3 text-sm text-gray-900">
                {alert.rule_name}
              </td>
              <td className="px-4 py-3 text-sm text-gray-600">
                {new Date(alert.fired_at).toLocaleString()}
              </td>
              <td className="px-4 py-3 text-sm text-gray-700 max-w-xs truncate">
                {alert.message}
              </td>
              <td className="px-4 py-3 text-sm">
                <span
                  className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                    alert.resolved
                      ? "bg-green-100 text-green-800"
                      : "bg-red-100 text-red-800"
                  }`}
                >
                  {alert.resolved ? "Resolved" : "Active"}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
