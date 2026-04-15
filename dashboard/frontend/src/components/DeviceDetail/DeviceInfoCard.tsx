import type { Device } from "../../api/types";

interface DeviceInfoCardProps {
  device: Device;
}

export default function DeviceInfoCard({ device }: DeviceInfoCardProps) {
  const fields: { label: string; value: string }[] = [
    { label: "Model", value: device.model },
    { label: "Serial Number", value: device.serial_number },
    { label: "Firmware", value: device.firmware_version },
    { label: "WWN", value: device.wwn },
    { label: "Capacity", value: `${device.capacity_gb.toLocaleString()} GB` },
    { label: "Interface", value: device.interface_type },
    { label: "Device Type", value: device.device_type },
    { label: "Host", value: device.host },
    { label: "Features", value: device.features.join(", ") },
    { label: "Temperature", value: `${device.temperature_c}°C` },
    {
      label: "Power-On Hours",
      value: `${device.power_on_hours.toLocaleString()}`,
    },
    { label: "SMART Status", value: device.smart_status },
  ];

  return (
    <div className="bg-white rounded-lg shadow border border-gray-200 p-6">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">
        Device Information
      </h3>
      <dl className="grid grid-cols-2 md:grid-cols-3 gap-x-6 gap-y-3">
        {fields.map((f) => (
          <div key={f.label}>
            <dt className="text-xs font-medium text-gray-500 uppercase tracking-wider">
              {f.label}
            </dt>
            <dd className="mt-1 text-sm text-gray-900 font-mono">{f.value}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
