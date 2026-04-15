import { useState, useMemo, useCallback, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useVirtualizer } from "@tanstack/react-virtual";
import type { Device } from "../../api/types";
import { useDevices } from "../../hooks/useDevices";
import FilterBar from "./FilterBar";
import DeviceRow from "./DeviceRow";

type SortKey = keyof Pick<
  Device,
  | "serial_number"
  | "model"
  | "device_type"
  | "host"
  | "firmware_version"
  | "smart_status"
  | "temperature_c"
  | "power_on_hours"
>;

const COLUMNS: { key: SortKey; label: string }[] = [
  { key: "serial_number", label: "Serial" },
  { key: "model", label: "Model" },
  { key: "device_type", label: "Type" },
  { key: "host", label: "Host" },
  { key: "firmware_version", label: "Firmware" },
  { key: "smart_status", label: "SMART" },
  { key: "temperature_c", label: "Temp" },
  { key: "power_on_hours", label: "POH" },
];

export default function FleetInventory() {
  const { devices, loading, error } = useDevices();
  const navigate = useNavigate();

  const [search, setSearch] = useState("");
  const [deviceType, setDeviceType] = useState("");
  const [smartStatus, setSmartStatus] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("serial_number");
  const [sortAsc, setSortAsc] = useState(true);

  const parentRef = useRef<HTMLDivElement>(null);

  const filtered = useMemo(() => {
    let result = devices;
    if (search) {
      const q = search.toLowerCase();
      result = result.filter(
        (d) =>
          d.serial_number.toLowerCase().includes(q) ||
          d.model.toLowerCase().includes(q)
      );
    }
    if (deviceType) {
      result = result.filter((d) => d.device_type === deviceType);
    }
    if (smartStatus) {
      result = result.filter((d) => d.smart_status === smartStatus);
    }
    return result;
  }, [devices, search, deviceType, smartStatus]);

  const sorted = useMemo(() => {
    const copy = [...filtered];
    copy.sort((a, b) => {
      const av = a[sortKey];
      const bv = b[sortKey];
      if (typeof av === "number" && typeof bv === "number") {
        return sortAsc ? av - bv : bv - av;
      }
      const as = String(av);
      const bs = String(bv);
      return sortAsc ? as.localeCompare(bs) : bs.localeCompare(as);
    });
    return copy;
  }, [filtered, sortKey, sortAsc]);

  const rowVirtualizer = useVirtualizer({
    count: sorted.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 44,
    overscan: 20,
  });

  const handleSort = useCallback(
    (key: SortKey) => {
      if (sortKey === key) {
        setSortAsc((prev) => !prev);
      } else {
        setSortKey(key);
        setSortAsc(true);
      }
    },
    [sortKey]
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4 text-red-700">
        Error loading devices: {error}
      </div>
    );
  }

  return (
    <div>
      <div className="mb-4">
        <h2 className="text-2xl font-bold text-gray-900">Fleet Inventory</h2>
        <p className="text-sm text-gray-500 mt-1">
          {sorted.length} of {devices.length} devices
        </p>
      </div>

      <FilterBar
        search={search}
        onSearchChange={setSearch}
        deviceType={deviceType}
        onDeviceTypeChange={setDeviceType}
        smartStatus={smartStatus}
        onSmartStatusChange={setSmartStatus}
      />

      <div className="bg-white rounded-lg shadow border border-gray-200 overflow-hidden">
        {/* Header */}
        <div className="grid grid-cols-8 gap-2 px-4 py-3 bg-gray-50 border-b border-gray-200 text-xs font-medium text-gray-500 uppercase tracking-wider">
          {COLUMNS.map((col) => (
            <button
              key={col.key}
              onClick={() => handleSort(col.key)}
              className="text-left hover:text-gray-700 flex items-center gap-1"
            >
              {col.label}
              {sortKey === col.key && (
                <span className="text-blue-600">{sortAsc ? "▲" : "▼"}</span>
              )}
            </button>
          ))}
        </div>

        {/* Virtualized rows */}
        <div
          ref={parentRef}
          className="overflow-auto"
          style={{ height: "600px" }}
        >
          <div
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`,
              width: "100%",
              position: "relative",
            }}
          >
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const device = sorted[virtualRow.index];
              return (
                <DeviceRow
                  key={device.serial_number}
                  device={device}
                  style={{
                    position: "absolute",
                    top: 0,
                    left: 0,
                    width: "100%",
                    height: `${virtualRow.size}px`,
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                  onClick={() =>
                    navigate(`/devices/${device.serial_number}`)
                  }
                />
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
