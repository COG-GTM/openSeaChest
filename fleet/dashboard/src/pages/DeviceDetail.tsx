import { useState, useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useDevice } from '../hooks/useDevices';
import { useHealthHistory, useFarmLatest, useHealthCurrent } from '../hooks/useHealth';
import { useAlerts } from '../hooks/useAlerts';
import StatusBadge from '../components/StatusBadge';
import HealthChart from '../components/HealthChart';

type TimeRange = '30d' | '90d' | '365d';

const RANGE_DAYS: Record<TimeRange, number> = {
  '30d': 30,
  '90d': 90,
  '365d': 365,
};

function InfoRow({ label, value }: { label: string; value: string | number | React.ReactNode }) {
  return (
    <div className="flex justify-between py-2 border-b border-gray-100 dark:border-gray-700 last:border-0">
      <span className="text-sm text-gray-500 dark:text-gray-400">{label}</span>
      <span className="text-sm font-medium text-gray-900 dark:text-white text-right">
        {value}
      </span>
    </div>
  );
}

function formatCapacity(bytes: number): string {
  const tb = bytes / 1e12;
  if (tb >= 1) return `${tb.toFixed(2)} TB`;
  const gb = bytes / 1e9;
  return `${gb.toFixed(0)} GB`;
}

export default function DeviceDetail() {
  const { serial } = useParams<{ serial: string }>();
  const [timeRange, setTimeRange] = useState<TimeRange>('30d');

  const { data: device, isLoading, isError } = useDevice(serial ?? '');
  const { data: currentHealth } = useHealthCurrent(serial ?? '');
  const { data: farm } = useFarmLatest(serial ?? '');

  const now = useMemo(() => new Date().toISOString(), []);
  const from = useMemo(() => {
    const d = new Date();
    d.setDate(d.getDate() - RANGE_DAYS[timeRange]);
    return d.toISOString();
  }, [timeRange]);

  const { data: healthHistory } = useHealthHistory(serial ?? '', from, now);

  const { data: deviceAlerts } = useAlerts({
    device: serial,
    page: 1,
    page_size: 10,
  });

  const tempData = useMemo(
    () =>
      healthHistory?.map((h) => ({
        timestamp: h.timestamp,
        value: h.temperature_c,
      })) ?? [],
    [healthHistory],
  );

  const pohData = useMemo(
    () =>
      healthHistory?.map((h) => ({
        timestamp: h.timestamp,
        value: h.power_on_hours,
      })) ?? [],
    [healthHistory],
  );

  const workloadData = useMemo(
    () =>
      healthHistory?.map((h) => ({
        timestamp: h.timestamp,
        value: h.workload_rate_tb_yr,
      })) ?? [],
    [healthHistory],
  );

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
      </div>
    );
  }

  if (isError || !device) {
    return (
      <div className="text-center py-12">
        <p className="text-danger-600 mb-4">Device not found.</p>
        <Link to="/devices" className="text-primary-600 hover:underline">
          Back to devices
        </Link>
      </div>
    );
  }

  const isSeagate =
    device.model.startsWith('ST') || device.model.toLowerCase().includes('seagate');

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link
          to="/devices"
          className="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
        >
          &larr; Devices
        </Link>
        <span className="text-gray-300 dark:text-gray-600">/</span>
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
          {device.serial_number}
        </h2>
        <StatusBadge status={device.smart_status} />
        <StatusBadge status={device.compliance_status} />
      </div>

      {/* Device identity card */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">
            Device Information
          </h3>
          <InfoRow label="Model Number" value={device.model} />
          <InfoRow label="Serial Number" value={device.serial_number} />
          <InfoRow label="Firmware Revision" value={device.firmware_revision} />
          <InfoRow label="World Wide Name" value={device.world_wide_name} />
          <InfoRow label="Drive Capacity" value={formatCapacity(device.capacity_bytes)} />
          <InfoRow label="Native Capacity" value={`${device.native_capacity_tb} TB`} />
          <InfoRow label="Interface" value={device.interface_type} />
          <InfoRow label="Rotation Rate" value={device.rotation_rate ? `${device.rotation_rate} RPM` : 'SSD'} />
          <InfoRow label="Form Factor" value={device.form_factor} />
          <InfoRow label="Logical Sector Size" value={`${device.logical_sector_size} B`} />
          <InfoRow label="Physical Sector Size" value={`${device.physical_sector_size} B`} />
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">
            Status & Performance
          </h3>
          <InfoRow label="SMART Status" value={<StatusBadge status={device.smart_status} size="sm" />} />
          <InfoRow label="Current Temperature" value={`${device.temperature_current}°C`} />
          <InfoRow label="Highest Temperature" value={`${device.temperature_highest}°C`} />
          <InfoRow label="Lowest Temperature" value={`${device.temperature_lowest}°C`} />
          <InfoRow label="Power-On Hours" value={device.power_on_hours.toLocaleString()} />
          <InfoRow label="Max Interface Speed" value={device.max_interface_speed} />
          <InfoRow label="Negotiated Speed" value={device.negotiated_interface_speed} />
          <InfoRow label="Encryption Support" value={device.encryption_support} />
          <InfoRow label="FW Download Support" value={device.firmware_download_support} />
          <InfoRow label="Workload Rate" value={`${device.annualized_workload_rate_tb_yr} TB/yr`} />
          <InfoRow label="Cache Size" value={device.cache_size_mib > 0 ? `${device.cache_size_mib} MiB` : 'N/A'} />
        </div>
      </div>

      {/* Current health */}
      {currentHealth && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">
            Current Health Snapshot
          </h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Temperature</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{currentHealth.temperature_c}°C</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Power-On Hours</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{currentHealth.power_on_hours.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Power Cycles</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{currentHealth.power_cycle_count.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Workload Rate</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{currentHealth.workload_rate_tb_yr} TB/yr</p>
            </div>
          </div>
        </div>
      )}

      {/* Health history charts */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white">
            Health History
          </h3>
          <div className="flex gap-1">
            {(['30d', '90d', '365d'] as TimeRange[]).map((range) => (
              <button
                key={range}
                onClick={() => setTimeRange(range)}
                className={`px-3 py-1 text-xs rounded ${
                  timeRange === range
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
                }`}
              >
                {range}
              </button>
            ))}
          </div>
        </div>
        <div className="space-y-6">
          <HealthChart data={tempData} label="Temperature" unit="°C" color="#ef4444" />
          <HealthChart data={pohData} label="Power-On Hours" unit="hrs" color="#3b82f6" />
          <HealthChart data={workloadData} label="Workload Rate" unit="TB/yr" color="#22c55e" />
        </div>
      </div>

      {/* FARM reliability section */}
      {isSeagate && farm && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">
            FARM Reliability Metrics
          </h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Unrecoverable Read Errors</p>
              <p className={`text-lg font-semibold ${farm.unrecoverable_read_errors > 0 ? 'text-danger-600' : 'text-gray-900 dark:text-white'}`}>
                {farm.unrecoverable_read_errors}
              </p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Unrecoverable Write Errors</p>
              <p className={`text-lg font-semibold ${farm.unrecoverable_write_errors > 0 ? 'text-danger-600' : 'text-gray-900 dark:text-white'}`}>
                {farm.unrecoverable_write_errors}
              </p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Total Read Commands</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{farm.total_read_commands.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Total Write Commands</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{farm.total_write_commands.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Power-On Hours (FARM)</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{farm.power_on_hours.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Power Cycles (FARM)</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{farm.power_cycle_count}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Temperature (FARM)</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">{farm.current_temperature_c}°C</p>
            </div>
          </div>
        </div>
      )}

      {/* Firmware compliance */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-5">
        <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">
          Firmware Compliance
        </h3>
        <InfoRow label="Current Firmware" value={device.firmware_revision} />
        <InfoRow
          label="Compliance Status"
          value={<StatusBadge status={device.compliance_status} size="sm" />}
        />
      </div>

      {/* Alert history */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        <div className="px-5 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white">
            Alert History
          </h3>
        </div>
        <div className="divide-y divide-gray-200 dark:divide-gray-700">
          {deviceAlerts?.items.map((alert) => (
            <div key={alert.id} className="px-5 py-3 flex items-center gap-4">
              <StatusBadge status={alert.severity} size="sm" />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-gray-900 dark:text-white truncate">{alert.message}</p>
              </div>
              <StatusBadge status={alert.status} size="sm" />
              <span className="text-xs text-gray-400 whitespace-nowrap">
                {new Date(alert.timestamp).toLocaleString()}
              </span>
            </div>
          ))}
          {(!deviceAlerts?.items.length) && (
            <div className="px-5 py-8 text-center text-sm text-gray-500 dark:text-gray-400">
              No alerts recorded for this device.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
