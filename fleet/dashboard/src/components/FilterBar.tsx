import type { FilterValues } from './filterTypes';

interface FilterBarProps {
  filters: FilterValues;
  onChange: (filters: FilterValues) => void;
  onReset: () => void;
}

export default function FilterBar({ filters, onChange, onReset }: FilterBarProps) {
  function update(key: keyof FilterValues, value: string) {
    onChange({ ...filters, [key]: value });
  }

  const inputClass =
    'rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm px-3 py-1.5 w-full';

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 mb-4">
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Model
          </label>
          <input
            type="text"
            value={filters.model}
            onChange={(e) => update('model', e.target.value)}
            placeholder="Model match..."
            className={inputClass}
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Firmware
          </label>
          <input
            type="text"
            value={filters.firmware}
            onChange={(e) => update('firmware', e.target.value)}
            placeholder="FW revision..."
            className={inputClass}
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Interface
          </label>
          <select
            value={filters.interface_type}
            onChange={(e) => update('interface_type', e.target.value)}
            className={inputClass}
          >
            <option value="">All</option>
            <option value="SATA">SATA</option>
            <option value="SAS">SAS</option>
            <option value="NVMe">NVMe</option>
          </select>
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Host
          </label>
          <input
            type="text"
            value={filters.host}
            onChange={(e) => update('host', e.target.value)}
            placeholder="Hostname..."
            className={inputClass}
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Health Status
          </label>
          <select
            value={filters.health_status}
            onChange={(e) => update('health_status', e.target.value)}
            className={inputClass}
          >
            <option value="">All</option>
            <option value="good">Good</option>
            <option value="warning">Warning</option>
            <option value="tripped">Tripped</option>
            <option value="unknown">Unknown</option>
          </select>
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Compliance
          </label>
          <select
            value={filters.compliance_status}
            onChange={(e) => update('compliance_status', e.target.value)}
            className={inputClass}
          >
            <option value="">All</option>
            <option value="compliant">Compliant</option>
            <option value="non_compliant">Non-Compliant</option>
            <option value="no_policy">No Policy</option>
          </select>
        </div>
      </div>
      <div className="mt-3 flex justify-end">
        <button
          onClick={onReset}
          className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400"
        >
          Reset Filters
        </button>
      </div>
    </div>
  );
}
