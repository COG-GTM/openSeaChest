import { useState, useCallback, useMemo } from 'react';
import { useDevices } from '../hooks/useDevices';
import { getHosts } from '../api/client';
import { useQuery } from '@tanstack/react-query';
import DeviceTable from '../components/DeviceTable';
import FilterBar from '../components/FilterBar';
import Pagination from '../components/Pagination';
import { EMPTY_FILTERS } from '../components/filterTypes';
import type { FilterValues } from '../components/filterTypes';
import type { DeviceQueryParams } from '../api/types';

export default function DeviceList() {
  const [filters, setFilters] = useState<FilterValues>(EMPTY_FILTERS);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [sortBy, setSortBy] = useState('model');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');

  const { data: hosts } = useQuery({
    queryKey: ['hosts'],
    queryFn: getHosts,
  });

  const hostMap = useMemo(() => {
    const map: Record<string, string> = {};
    if (hosts) {
      for (const h of hosts) {
        map[h.id] = h.hostname;
      }
    }
    return map;
  }, [hosts]);

  const queryParams: DeviceQueryParams = {
    page,
    page_size: pageSize,
    sort_by: sortBy,
    sort_order: sortOrder,
    ...(filters.model && { model: filters.model }),
    ...(filters.firmware && { firmware: filters.firmware }),
    ...(filters.interface_type && { interface_type: filters.interface_type }),
    ...(filters.host && { host: filters.host }),
    ...(filters.health_status && { health_status: filters.health_status }),
    ...(filters.compliance_status && { compliance_status: filters.compliance_status }),
  };

  const { data, isLoading, isError } = useDevices(queryParams);

  const handleSort = useCallback(
    (key: string) => {
      if (key === sortBy) {
        setSortOrder((prev) => (prev === 'asc' ? 'desc' : 'asc'));
      } else {
        setSortBy(key);
        setSortOrder('asc');
      }
      setPage(1);
    },
    [sortBy],
  );

  const handleFilterChange = useCallback((newFilters: FilterValues) => {
    setFilters(newFilters);
    setPage(1);
  }, []);

  const handleReset = useCallback(() => {
    setFilters(EMPTY_FILTERS);
    setPage(1);
  }, []);

  function exportCsv() {
    if (!data?.items.length) return;
    const headers = [
      'Host', 'Model', 'Serial Number', 'FW Rev', 'Interface',
      'Capacity (bytes)', 'SMART Status', 'Temperature (C)', 'Power-On Hours', 'Compliance',
    ];
    const rows = data.items.map((d) => [
      hostMap[d.host_id] ?? d.host_id,
      d.model,
      d.serial_number,
      d.firmware_revision,
      d.interface_type,
      d.capacity_bytes,
      d.smart_status,
      d.temperature_current,
      d.power_on_hours,
      d.compliance_status,
    ]);
    const escapeCsv = (val: unknown) => {
      const s = String(val);
      return s.includes(',') || s.includes('"') || s.includes('\n')
        ? '"' + s.replace(/"/g, '""') + '"'
        : s;
    };
    const csv = [headers.map(escapeCsv).join(','), ...rows.map((r) => r.map(escapeCsv).join(','))].join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'fleet-devices.csv';
    a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
          Devices
        </h2>
        <button
          onClick={exportCsv}
          className="px-3 py-1.5 text-sm bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          Export CSV
        </button>
      </div>

      <FilterBar
        filters={filters}
        onChange={handleFilterChange}
        onReset={handleReset}
      />

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
          </div>
        ) : isError ? (
          <div className="p-8 text-center text-danger-600">
            Failed to load devices. Please try again.
          </div>
        ) : (
          <>
            <DeviceTable
              devices={data?.items ?? []}
              sortBy={sortBy}
              sortOrder={sortOrder}
              onSort={handleSort}
              hostMap={hostMap}
            />
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
    </div>
  );
}
