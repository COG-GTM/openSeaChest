import { useParams, Link } from "react-router-dom";
import { useDevice } from "../../hooks/useDevices";
import { useDeviceHealth } from "../../hooks/useDeviceHealth";
import DeviceInfoCard from "./DeviceInfoCard";
import HealthChart from "./HealthChart";

export default function DeviceDetail() {
  const { serial } = useParams<{ serial: string }>();
  const { device, loading: deviceLoading, error: deviceError } = useDevice(
    serial ?? ""
  );
  const { history, loading: healthLoading, error: healthError } =
    useDeviceHealth(serial ?? "");

  if (deviceLoading || healthLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    );
  }

  if (deviceError || healthError) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4 text-red-700">
        Error: {deviceError ?? healthError}
      </div>
    );
  }

  if (!device) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4 text-yellow-700">
        Device not found.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link
          to="/"
          className="text-blue-600 hover:text-blue-800 text-sm font-medium"
        >
          &larr; Back to Fleet
        </Link>
        <h2 className="text-2xl font-bold text-gray-900">
          Device: {device.serial_number}
        </h2>
      </div>

      <DeviceInfoCard device={device} />

      <HealthChart data={history} />

      {/* Health Snapshot History Table */}
      <div className="bg-white rounded-lg shadow border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">
            Health Snapshot History
          </h3>
        </div>
        <div className="overflow-x-auto max-h-96 overflow-y-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50 sticky top-0">
              <tr>
                {[
                  "Timestamp",
                  "Temp (°C)",
                  "POH",
                  "SMART",
                  "Reallocated",
                  "Pending",
                  "CRC Errors",
                ].map((h) => (
                  <th
                    key={h}
                    className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                  >
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {history.map((snap) => (
                <tr key={snap.id} className="hover:bg-gray-50">
                  <td className="px-4 py-2 text-sm text-gray-600">
                    {new Date(snap.timestamp).toLocaleString()}
                  </td>
                  <td className="px-4 py-2 text-sm">{snap.temperature_c}</td>
                  <td className="px-4 py-2 text-sm">
                    {snap.power_on_hours.toLocaleString()}
                  </td>
                  <td className="px-4 py-2 text-sm">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                        snap.smart_status === "PASSED"
                          ? "bg-green-100 text-green-800"
                          : snap.smart_status === "FAILED"
                            ? "bg-red-100 text-red-800"
                            : "bg-yellow-100 text-yellow-800"
                      }`}
                    >
                      {snap.smart_status}
                    </span>
                  </td>
                  <td className="px-4 py-2 text-sm">
                    {snap.reallocated_sectors}
                  </td>
                  <td className="px-4 py-2 text-sm">
                    {snap.pending_sectors}
                  </td>
                  <td className="px-4 py-2 text-sm">{snap.crc_errors}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
