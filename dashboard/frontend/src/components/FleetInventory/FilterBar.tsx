interface FilterBarProps {
  search: string;
  onSearchChange: (value: string) => void;
  deviceType: string;
  onDeviceTypeChange: (value: string) => void;
  smartStatus: string;
  onSmartStatusChange: (value: string) => void;
}

export default function FilterBar({
  search,
  onSearchChange,
  deviceType,
  onDeviceTypeChange,
  smartStatus,
  onSmartStatusChange,
}: FilterBarProps) {
  return (
    <div className="flex flex-wrap gap-4 mb-4">
      <div className="flex-1 min-w-[200px]">
        <input
          type="text"
          placeholder="Search by serial or model..."
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
        />
      </div>
      <select
        value={deviceType}
        onChange={(e) => onDeviceTypeChange(e.target.value)}
        className="px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
      >
        <option value="">All Device Types</option>
        <option value="HDD">HDD</option>
        <option value="SSD">SSD</option>
        <option value="NVMe">NVMe</option>
      </select>
      <select
        value={smartStatus}
        onChange={(e) => onSmartStatusChange(e.target.value)}
        className="px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
      >
        <option value="">All SMART Status</option>
        <option value="PASSED">PASSED</option>
        <option value="FAILED">FAILED</option>
        <option value="UNKNOWN">UNKNOWN</option>
      </select>
    </div>
  );
}
