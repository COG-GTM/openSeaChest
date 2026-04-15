import type { Device } from "../../api/types";

interface DeviceRowProps {
  device: Device;
  style: React.CSSProperties;
  onClick: () => void;
}

function smartBadge(status: Device["smart_status"]) {
  const colors: Record<Device["smart_status"], string> = {
    PASSED: "bg-green-100 text-green-800",
    FAILED: "bg-red-100 text-red-800",
    UNKNOWN: "bg-yellow-100 text-yellow-800",
  };
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${colors[status]}`}
    >
      {status}
    </span>
  );
}

function tempColor(temp: number): string {
  if (temp >= 50) return "text-red-600 font-semibold";
  if (temp >= 40) return "text-yellow-600";
  return "text-gray-700";
}

export default function DeviceRow({ device, style, onClick }: DeviceRowProps) {
  return (
    <div
      style={style}
      onClick={onClick}
      className="grid grid-cols-8 gap-2 items-center px-4 py-2 border-b border-gray-100 hover:bg-blue-50 cursor-pointer text-sm transition-colors"
      role="row"
    >
      <div className="font-mono text-xs truncate" title={device.serial_number}>
        {device.serial_number}
      </div>
      <div className="truncate" title={device.model}>
        {device.model}
      </div>
      <div>
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700">
          {device.device_type}
        </span>
      </div>
      <div className="truncate text-gray-600">{device.host}</div>
      <div className="font-mono text-xs">{device.firmware_version}</div>
      <div>{smartBadge(device.smart_status)}</div>
      <div className={tempColor(device.temperature_c)}>
        {device.temperature_c}°C
      </div>
      <div className="text-gray-600">
        {device.power_on_hours.toLocaleString()}h
      </div>
    </div>
  );
}
