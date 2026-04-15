import type { Device } from "../api/types";

const MODELS = [
  "ST8000NM000A",
  "ST4000NM002A",
  "ST16000NM001G",
  "ST12000NM001G",
  "ST2000NM000A",
  "ZA960NM10001",
  "XP3840SE70005",
];

const HOSTS = [
  "server-rack-01",
  "server-rack-02",
  "server-rack-03",
  "nas-bay-01",
  "nas-bay-02",
  "workstation-lab",
];

const FIRMWARE_VERSIONS = ["SN03", "SN04", "SN05", "SN06", "EN01", "EN02"];
const DEVICE_TYPES: Device["device_type"][] = ["HDD", "SSD", "NVMe"];
const SMART_STATUSES: Device["smart_status"][] = [
  "PASSED",
  "PASSED",
  "PASSED",
  "PASSED",
  "FAILED",
  "UNKNOWN",
];

function pick<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

function generateSerial(index: number): string {
  return `WDG${String(index).padStart(4, "0")}${Math.random().toString(36).slice(2, 6).toUpperCase()}`;
}

function generateWWN(): string {
  const hex = () =>
    Math.floor(Math.random() * 0xffff)
      .toString(16)
      .padStart(4, "0");
  return `5000C500${hex()}${hex()}`;
}

function createDevice(index: number): Device {
  const deviceType = pick(DEVICE_TYPES);
  return {
    serial_number: generateSerial(index),
    model: pick(MODELS),
    device_type: deviceType,
    host: pick(HOSTS),
    firmware_version: pick(FIRMWARE_VERSIONS),
    smart_status: pick(SMART_STATUSES),
    temperature_c: Math.floor(Math.random() * 30) + 25,
    power_on_hours: Math.floor(Math.random() * 50000) + 1000,
    wwn: generateWWN(),
    capacity_gb:
      deviceType === "NVMe"
        ? pick([960, 1920, 3840])
        : pick([2000, 4000, 8000, 12000, 16000]),
    interface_type:
      deviceType === "NVMe" ? "NVMe" : pick(["SATA", "SAS"]),
    features: deviceType === "SSD" ? ["TRIM", "NCQ"] : ["NCQ", "APM", "EPC"],
  };
}

// Seed a consistent dataset
const seededRandom = (() => {
  let seed = 42;
  return () => {
    seed = (seed * 16807) % 2147483647;
    return (seed - 1) / 2147483646;
  };
})();

// Override Math.random temporarily for consistent data
const origRandom = Math.random;
Math.random = seededRandom;

export const mockDevices: Device[] = Array.from({ length: 120 }, (_, i) =>
  createDevice(i)
);

Math.random = origRandom;
