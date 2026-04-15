import { http, HttpResponse } from "msw";
import type { AlertRule, FiredAlert } from "../api/types";
import { mockDevices } from "./devices";
import { mockHealthData } from "./health";

/** In-memory store for alert rules (mutable). */
let alertRules: AlertRule[] = [
  {
    id: "rule-1",
    name: "SMART Failure Alert",
    rule_type: "smart_tripped",
    threshold_value: null,
    enabled: true,
    created_at: new Date(Date.now() - 86400_000 * 7).toISOString(),
  },
  {
    id: "rule-2",
    name: "High Temperature",
    rule_type: "temperature_threshold",
    threshold_value: 50,
    enabled: true,
    created_at: new Date(Date.now() - 86400_000 * 3).toISOString(),
  },
  {
    id: "rule-3",
    name: "POH Threshold",
    rule_type: "poh_threshold",
    threshold_value: 40000,
    enabled: false,
    created_at: new Date(Date.now() - 86400_000).toISOString(),
  },
];

const firedAlerts: FiredAlert[] = [
  {
    id: "fa-1",
    device_serial: mockDevices[4]?.serial_number ?? "UNKNOWN",
    rule_id: "rule-1",
    rule_name: "SMART Failure Alert",
    fired_at: new Date(Date.now() - 3600_000 * 2).toISOString(),
    message: `SMART status FAILED on device ${mockDevices[4]?.serial_number ?? "UNKNOWN"}`,
    resolved: false,
  },
  {
    id: "fa-2",
    device_serial: mockDevices[10]?.serial_number ?? "UNKNOWN",
    rule_id: "rule-2",
    rule_name: "High Temperature",
    fired_at: new Date(Date.now() - 3600_000 * 5).toISOString(),
    message: `Temperature 53°C exceeds threshold 50°C on device ${mockDevices[10]?.serial_number ?? "UNKNOWN"}`,
    resolved: true,
  },
  {
    id: "fa-3",
    device_serial: mockDevices[20]?.serial_number ?? "UNKNOWN",
    rule_id: "rule-2",
    rule_name: "High Temperature",
    fired_at: new Date(Date.now() - 3600_000).toISOString(),
    message: `Temperature 52°C exceeds threshold 50°C on device ${mockDevices[20]?.serial_number ?? "UNKNOWN"}`,
    resolved: false,
  },
];

let nextRuleId = 4;

export const handlers = [
  // GET all devices
  http.get("/api/v1/devices", () => {
    return HttpResponse.json(mockDevices);
  }),

  // GET single device by serial
  http.get("/api/v1/devices/:serial", ({ params }) => {
    const device = mockDevices.find(
      (d) => d.serial_number === params.serial
    );
    if (!device) {
      return new HttpResponse(null, { status: 404 });
    }
    return HttpResponse.json(device);
  }),

  // GET device health history
  http.get("/api/v1/devices/:serial/health/history", ({ params }) => {
    const serial = params.serial as string;
    const history = mockHealthData[serial] ?? [];
    return HttpResponse.json(history);
  }),

  // GET firmware compliance
  http.get("/api/v1/firmware/compliance", ({ request }) => {
    const url = new URL(request.url);
    const targetFw = url.searchParams.get("target_firmware") ?? "";
    const compliant = mockDevices.filter(
      (d) => d.firmware_version === targetFw
    );
    const nonCompliant = mockDevices.filter(
      (d) => d.firmware_version !== targetFw
    );
    return HttpResponse.json({
      compliant,
      non_compliant: nonCompliant,
    });
  }),

  // GET fired alerts
  http.get("/api/v1/alerts", () => {
    return HttpResponse.json(firedAlerts);
  }),

  // GET alert rules
  http.get("/api/v1/alerts/rules", () => {
    return HttpResponse.json(alertRules);
  }),

  // POST create alert rule
  http.post("/api/v1/alerts/rules", async ({ request }) => {
    const body = (await request.json()) as Omit<AlertRule, "id" | "created_at">;
    const newRule: AlertRule = {
      ...body,
      id: `rule-${nextRuleId++}`,
      created_at: new Date().toISOString(),
    };
    alertRules.push(newRule);
    return HttpResponse.json(newRule, { status: 201 });
  }),

  // PUT update alert rule
  http.put("/api/v1/alerts/rules/:id", async ({ params, request }) => {
    const body = (await request.json()) as Partial<AlertRule>;
    const index = alertRules.findIndex((r) => r.id === params.id);
    if (index === -1) {
      return new HttpResponse(null, { status: 404 });
    }
    alertRules[index] = { ...alertRules[index], ...body };
    return HttpResponse.json(alertRules[index]);
  }),

  // DELETE alert rule
  http.delete("/api/v1/alerts/rules/:id", ({ params }) => {
    const index = alertRules.findIndex((r) => r.id === params.id);
    if (index === -1) {
      return new HttpResponse(null, { status: 404 });
    }
    alertRules = alertRules.filter((r) => r.id !== params.id);
    return new HttpResponse(null, { status: 204 });
  }),
];
