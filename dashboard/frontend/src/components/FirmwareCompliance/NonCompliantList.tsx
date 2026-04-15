import { useNavigate } from "react-router-dom";
import type { Device } from "../../api/types";

interface NonCompliantListProps {
  devices: Device[];
  targetFirmware: string;
}

export default function NonCompliantList({
  devices,
  targetFirmware,
}: NonCompliantListProps) {
  const navigate = useNavigate();

  if (devices.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        All devices are running firmware {targetFirmware}.
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {["Serial", "Model", "Type", "Host", "Current Firmware"].map(
              (h) => (
                <th
                  key={h}
                  className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  {h}
                </th>
              )
            )}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {devices.map((device) => (
            <tr
              key={device.serial_number}
              onClick={() => navigate(`/devices/${device.serial_number}`)}
              className="hover:bg-blue-50 cursor-pointer"
            >
              <td className="px-4 py-2 text-sm font-mono">
                {device.serial_number}
              </td>
              <td className="px-4 py-2 text-sm">{device.model}</td>
              <td className="px-4 py-2 text-sm">{device.device_type}</td>
              <td className="px-4 py-2 text-sm text-gray-600">{device.host}</td>
              <td className="px-4 py-2 text-sm">
                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800">
                  {device.firmware_version}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
