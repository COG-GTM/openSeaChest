import { useState, useCallback } from 'react';
import { useAlerts, useAcknowledgeAlert } from '../hooks/useAlerts';
import StatusBadge from '../components/StatusBadge';
import Pagination from '../components/Pagination';
import type { AlertSeverity, AlertStatus, AlertQueryParams } from '../api/types';

export default function Alerts() {
  const [severity, setSeverity] = useState<AlertSeverity | ''>('');
  const [device, setDevice] = useState('');
  const [host, setHost] = useState('');
  const [status, setStatus] = useState<AlertStatus | ''>('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [ackModal, setAckModal] = useState<string | null>(null);
  const [annotation, setAnnotation] = useState('');

  const params: AlertQueryParams = {
    page,
    page_size: pageSize,
    ...(severity && { severity }),
    ...(device && { device }),
    ...(host && { host }),
    ...(status && { status }),
  };

  const { data, isLoading, isError } = useAlerts(params);
  const ackMutation = useAcknowledgeAlert();

  const handleAcknowledge = useCallback(() => {
    if (!ackModal) return;
    ackMutation.mutate(
      { id: ackModal, annotation },
      {
        onSuccess: () => {
          setAckModal(null);
          setAnnotation('');
        },
      },
    );
  }, [ackModal, annotation, ackMutation]);

  const inputClass =
    'rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm px-3 py-1.5 w-full';

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Alerts</h2>

      {/* Filters */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div>
            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
              Severity
            </label>
            <select
              value={severity}
              onChange={(e) => {
                setSeverity(e.target.value as AlertSeverity | '');
                setPage(1);
              }}
              className={inputClass}
            >
              <option value="">All</option>
              <option value="critical">Critical</option>
              <option value="warning">Warning</option>
              <option value="info">Info</option>
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
              Device
            </label>
            <input
              type="text"
              value={device}
              onChange={(e) => {
                setDevice(e.target.value);
                setPage(1);
              }}
              placeholder="Serial number..."
              className={inputClass}
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
              Host
            </label>
            <input
              type="text"
              value={host}
              onChange={(e) => {
                setHost(e.target.value);
                setPage(1);
              }}
              placeholder="Hostname..."
              className={inputClass}
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
              Status
            </label>
            <select
              value={status}
              onChange={(e) => {
                setStatus(e.target.value as AlertStatus | '');
                setPage(1);
              }}
              className={inputClass}
            >
              <option value="">All</option>
              <option value="firing">Firing</option>
              <option value="acknowledged">Acknowledged</option>
              <option value="resolved">Resolved</option>
            </select>
          </div>
        </div>
      </div>

      {/* Alerts table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
          </div>
        ) : isError ? (
          <div className="p-8 text-center text-danger-600">
            Failed to load alerts. Please try again.
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-800">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Timestamp
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Severity
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Device
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Host
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Metric
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Message
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Status
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                  {data?.items.map((alert) => (
                    <tr key={alert.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                      <td className="px-4 py-3 text-xs text-gray-500 whitespace-nowrap">
                        {new Date(alert.timestamp).toLocaleString()}
                      </td>
                      <td className="px-4 py-3">
                        <StatusBadge status={alert.severity} size="sm" />
                      </td>
                      <td className="px-4 py-3 text-sm font-mono text-gray-700 dark:text-gray-300">
                        {alert.device_serial}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                        {alert.host}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                        {alert.metric}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700 dark:text-gray-300 max-w-xs truncate">
                        {alert.message}
                      </td>
                      <td className="px-4 py-3">
                        <StatusBadge status={alert.status} size="sm" />
                      </td>
                      <td className="px-4 py-3">
                        {alert.status === 'firing' && (
                          <button
                            onClick={() => setAckModal(alert.id)}
                            className="text-xs px-2 py-1 bg-warning-500 text-white rounded hover:bg-warning-600 transition-colors"
                          >
                            Acknowledge
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                  {data?.items.length === 0 && (
                    <tr>
                      <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500">
                        No alerts matching the current filters.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
            <Pagination
              page={page}
              pageSize={pageSize}
              total={data?.total ?? 0}
              onPageChange={setPage}
              onPageSizeChange={(size) => {
                setPageSize(size);
                setPage(1);
              }}
            />
          </>
        )}
      </div>

      {/* Acknowledge modal */}
      {ackModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Acknowledge Alert
            </h3>
            <textarea
              value={annotation}
              onChange={(e) => setAnnotation(e.target.value)}
              placeholder="Add an annotation (optional)..."
              className="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 p-3 text-sm min-h-[100px]"
            />
            <div className="flex justify-end gap-2 mt-4">
              <button
                onClick={() => {
                  setAckModal(null);
                  setAnnotation('');
                }}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 dark:text-gray-400"
              >
                Cancel
              </button>
              <button
                onClick={handleAcknowledge}
                disabled={ackMutation.isPending}
                className="px-4 py-2 text-sm bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50"
              >
                {ackMutation.isPending ? 'Saving...' : 'Acknowledge'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
