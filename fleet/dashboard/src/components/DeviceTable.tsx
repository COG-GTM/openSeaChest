import { useNavigate } from 'react-router-dom';
import type { DeviceInfo } from '../api/types';
import StatusBadge from './StatusBadge';

interface Column {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (device: DeviceInfo) => React.ReactNode;
}

interface DeviceTableProps {
  devices: DeviceInfo[];
  sortBy: string;
  sortOrder: 'asc' | 'desc';
  onSort: (key: string) => void;
  hostMap: Record<string, string>;
}

function formatCapacity(bytes: number): string {
  const tb = bytes / 1e12;
  if (tb >= 1) return `${tb.toFixed(1)} TB`;
  const gb = bytes / 1e9;
  return `${gb.toFixed(0)} GB`;
}

export default function DeviceTable({
  devices,
  sortBy,
  sortOrder,
  onSort,
  hostMap,
}: DeviceTableProps) {
  const navigate = useNavigate();

  const columns: Column[] = [
    {
      key: 'host_id',
      label: 'Host',
      sortable: true,
      render: (d) => (
        <span className="text-gray-700 dark:text-gray-300">
          {hostMap[d.host_id] ?? d.host_id}
        </span>
      ),
    },
    { key: 'model', label: 'Model', sortable: true },
    { key: 'serial_number', label: 'Serial Number', sortable: true },
    { key: 'firmware_revision', label: 'FW Rev', sortable: true },
    { key: 'interface_type', label: 'Interface', sortable: true },
    {
      key: 'capacity_bytes',
      label: 'Capacity',
      sortable: true,
      render: (d) => formatCapacity(d.capacity_bytes),
    },
    {
      key: 'smart_status',
      label: 'SMART Status',
      sortable: true,
      render: (d) => <StatusBadge status={d.smart_status} size="sm" />,
    },
    {
      key: 'temperature_current',
      label: 'Temp (°C)',
      sortable: true,
      render: (d) => (
        <span className={d.temperature_current > 50 ? 'text-danger-600 font-medium' : ''}>
          {d.temperature_current}°C
        </span>
      ),
    },
    {
      key: 'power_on_hours',
      label: 'Power-On Hours',
      sortable: true,
      render: (d) => d.power_on_hours.toLocaleString(),
    },
    {
      key: 'compliance_status',
      label: 'Compliance',
      sortable: true,
      render: (d) => <StatusBadge status={d.compliance_status} size="sm" />,
    },
  ];

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-800">
          <tr>
            {columns.map((col) => (
              <th
                key={col.key}
                onClick={() => col.sortable && onSort(col.key)}
                className={`px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider ${
                  col.sortable ? 'cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none' : ''
                }`}
              >
                <span className="flex items-center gap-1">
                  {col.label}
                  {col.sortable && sortBy === col.key && (
                    <span>{sortOrder === 'asc' ? '↑' : '↓'}</span>
                  )}
                </span>
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
          {devices.map((device) => (
            <tr
              key={device.serial_number}
              onClick={() => navigate(`/devices/${device.serial_number}`)}
              className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer transition-colors"
            >
              {columns.map((col) => (
                <td
                  key={col.key}
                  className="px-4 py-3 text-sm text-gray-700 dark:text-gray-300 whitespace-nowrap"
                >
                  {col.render
                    ? col.render(device)
                    : String(device[col.key as keyof DeviceInfo])}
                </td>
              ))}
            </tr>
          ))}
          {devices.length === 0 && (
            <tr>
              <td
                colSpan={columns.length}
                className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400"
              >
                No devices found matching the current filters.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
