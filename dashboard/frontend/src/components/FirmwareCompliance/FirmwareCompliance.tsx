import { useState } from "react";
import { useFirmwareCompliance } from "../../hooks/useFirmwareCompliance";
import CompliancePieChart from "./CompliancePieChart";
import NonCompliantList from "./NonCompliantList";

export default function FirmwareCompliance() {
  const [targetFw, setTargetFw] = useState("");
  const { result, loading, error, checkCompliance } = useFirmwareCompliance();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (targetFw.trim()) {
      checkCompliance(targetFw.trim());
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900">
          Firmware Compliance
        </h2>
        <p className="text-sm text-gray-500 mt-1">
          Check how many devices are running the target firmware version.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="flex gap-3 items-end">
        <div className="flex-1 max-w-sm">
          <label
            htmlFor="target-fw"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Target Firmware Version
          </label>
          <input
            id="target-fw"
            type="text"
            value={targetFw}
            onChange={(e) => setTargetFw(e.target.value)}
            placeholder="e.g. SN04"
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
          />
        </div>
        <button
          type="submit"
          disabled={loading || !targetFw.trim()}
          className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {loading ? "Checking..." : "Check Compliance"}
        </button>
      </form>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4 text-red-700">
          Error: {error}
        </div>
      )}

      {result && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="bg-white rounded-lg shadow border border-gray-200 p-4 text-center">
              <div className="text-3xl font-bold text-gray-900">
                {result.compliant.length + result.non_compliant.length}
              </div>
              <div className="text-sm text-gray-500">Total Devices</div>
            </div>
            <div className="bg-white rounded-lg shadow border border-green-200 p-4 text-center">
              <div className="text-3xl font-bold text-green-600">
                {result.compliant.length}
              </div>
              <div className="text-sm text-gray-500">Compliant</div>
            </div>
            <div className="bg-white rounded-lg shadow border border-red-200 p-4 text-center">
              <div className="text-3xl font-bold text-red-600">
                {result.non_compliant.length}
              </div>
              <div className="text-sm text-gray-500">Non-Compliant</div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow border border-gray-200 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Compliance Distribution
            </h3>
            <CompliancePieChart
              compliantCount={result.compliant.length}
              nonCompliantCount={result.non_compliant.length}
            />
          </div>

          <div className="bg-white rounded-lg shadow border border-gray-200 overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">
                Non-Compliant Devices ({result.non_compliant.length})
              </h3>
            </div>
            <NonCompliantList
              devices={result.non_compliant}
              targetFirmware={targetFw}
            />
          </div>
        </div>
      )}
    </div>
  );
}
