import { Link } from 'react-router-dom';
import { useFleetHealthSummary } from '../hooks/useHealth';
import { useAlerts } from '../hooks/useAlerts';
import { getComplianceSummary } from '../api/client';
import { useQuery } from '@tanstack/react-query';
import FleetPieChart from '../components/PieChart';
import StatusBadge from '../components/StatusBadge';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Legend,
} from 'recharts';

function StatCard({
  label,
  value,
  color,
  to,
}: {
  label: string;
  value: number | string;
  color: string;
  to: string;
}) {
  return (
    <Link
      to={to}
      className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5 hover:shadow-md transition-shadow"
    >
      <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">{label}</p>
      <p className={`text-2xl font-bold ${color}`}>{value}</p>
    </Link>
  );
}

export default function FleetOverview() {
  const { data: health, isLoading: healthLoading } = useFleetHealthSummary();
  const { data: alertsData } = useAlerts({ page: 1, page_size: 10 });
  const { data: complianceSummary } = useQuery({
    queryKey: ['compliance', 'summary'],
    queryFn: getComplianceSummary,
  });

  const totalDevices = health ? health.total : 0;
  const warningDevices = health?.warning ?? 0;
  const trippedDevices = health?.tripped ?? 0;

  const nonCompliant = complianceSummary
    ? complianceSummary.reduce((acc, s) => acc + s.non_compliant, 0)
    : 0;

  const totalHosts = health?.total_hosts ?? 0;
  const staleDevices = health?.stale_devices ?? 0;

  const pieData = health
    ? [
        { name: 'Good', value: health.good },
        { name: 'Warning', value: health.warning },
        { name: 'Tripped', value: health.tripped },
        { name: 'Unknown', value: health.unknown },
      ].filter((d) => d.value > 0)
    : [];

  const complianceBarData = complianceSummary
    ? complianceSummary.map((s) => ({
        model: s.model.length > 20 ? s.model.substring(0, 20) + '...' : s.model,
        Compliant: s.compliant,
        'Non-Compliant': s.non_compliant,
        'No Policy': s.no_policy,
      }))
    : [];

  const recentAlerts = alertsData?.items
    .filter((a) => a.severity === 'critical' || a.severity === 'warning')
    .slice(0, 10) ?? [];

  if (healthLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
        Fleet Overview
      </h2>

      {/* Stats cards */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        <StatCard
          label="Total Devices"
          value={totalDevices}
          color="text-gray-900 dark:text-white"
          to="/devices"
        />
        <StatCard
          label="Total Hosts"
          value={totalHosts}
          color="text-gray-900 dark:text-white"
          to="/devices"
        />
        <StatCard
          label="Warnings"
          value={warningDevices}
          color="text-warning-600"
          to="/devices?health_status=warning"
        />
        <StatCard
          label="SMART Tripped"
          value={trippedDevices}
          color="text-danger-600"
          to="/devices?health_status=tripped"
        />
        <StatCard
          label="Non-Compliant FW"
          value={nonCompliant}
          color="text-danger-600"
          to="/firmware"
        />
        <StatCard
          label="Stale Devices"
          value={staleDevices}
          color="text-gray-500"
          to="/devices"
        />
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Health Pie Chart */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5">
          <FleetPieChart data={pieData} title="Fleet Health Status" />
        </div>

        {/* Compliance Bar Chart */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5">
          <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">
            Firmware Compliance by Model
          </h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={complianceBarData} layout="vertical" margin={{ left: 20 }}>
                <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                <XAxis type="number" tick={{ fontSize: 11 }} />
                <YAxis
                  type="category"
                  dataKey="model"
                  tick={{ fontSize: 10 }}
                  width={140}
                />
                <Tooltip />
                <Legend />
                <Bar dataKey="Compliant" stackId="a" fill="#22c55e" />
                <Bar dataKey="Non-Compliant" stackId="a" fill="#ef4444" />
                <Bar dataKey="No Policy" stackId="a" fill="#9ca3af" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Recent alerts */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        <div className="px-5 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h3 className="text-sm font-medium text-gray-900 dark:text-white">
            Recent Critical/Warning Alerts
          </h3>
          <Link
            to="/alerts"
            className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400"
          >
            View all
          </Link>
        </div>
        <div className="divide-y divide-gray-200 dark:divide-gray-700">
          {recentAlerts.map((alert) => (
            <div
              key={alert.id}
              className="px-5 py-3 flex items-center gap-4"
            >
              <StatusBadge status={alert.severity} size="sm" />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-gray-900 dark:text-white truncate">
                  {alert.message}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  {alert.device_serial} · {alert.host}
                </p>
              </div>
              <span className="text-xs text-gray-400 whitespace-nowrap">
                {new Date(alert.timestamp).toLocaleString()}
              </span>
            </div>
          ))}
          {recentAlerts.length === 0 && (
            <div className="px-5 py-8 text-center text-sm text-gray-500 dark:text-gray-400">
              No recent critical or warning alerts.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
