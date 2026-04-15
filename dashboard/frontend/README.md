# OpenSeaChest Fleet Dashboard — Frontend

A React + TypeScript web dashboard for monitoring and managing a fleet of storage devices via the OpenSeaChest backend API.

## Tech Stack

- **Vite** — fast build tool and dev server
- **React 18** — UI framework
- **TypeScript** — strict mode enabled
- **React Router v6** — client-side routing
- **Recharts** — time-series and pie charts
- **@tanstack/react-virtual** — virtual scrolling for large device lists
- **Tailwind CSS** — utility-first styling
- **MSW (Mock Service Worker)** — API mocking for development

## Getting Started

### Prerequisites

- **Node.js** >= 18
- **npm** >= 9

### Install Dependencies

```bash
cd dashboard/frontend
npm install
```

### Start Development Server

```bash
npm run dev
```

The app will be available at [http://localhost:5173](http://localhost:5173).

In development mode, MSW intercepts API calls and returns mock data so the frontend works without a running backend.

### Build for Production

```bash
npm run build
```

Output is written to `dist/`.

### Preview Production Build

```bash
npm run preview
```

## Project Structure

```
src/
  main.tsx              # Entry point (bootstraps MSW in dev)
  App.tsx               # Router configuration
  index.css             # Tailwind imports
  api/
    client.ts           # Typed fetch wrappers for the backend API
    types.ts            # TypeScript interfaces matching the backend OpenAPI spec
  components/
    Layout/             # App shell with navigation header
    FleetInventory/     # Sortable, filterable device table with virtual scrolling
    DeviceDetail/       # Device info card + Recharts time-series charts
    FirmwareCompliance/ # Target FW input, pie chart, non-compliant drill-down
    Alerts/             # Fired alerts list + CRUD for alert rules
  hooks/
    useDevices.ts       # Fetch all devices / single device
    useDeviceHealth.ts  # Fetch health snapshot history
    useFirmwareCompliance.ts  # Firmware compliance check
    useAlerts.ts        # Fired alerts + alert rule CRUD
  mocks/
    browser.ts          # MSW worker setup
    handlers.ts         # MSW request handlers
    devices.ts          # Mock device data (120 devices)
    health.ts           # Mock health snapshot data
```

## Views

| Route | View | Description |
|---|---|---|
| `/` | Fleet Inventory | Sortable/filterable table of all devices |
| `/devices/:serial` | Device Detail | Device info + temperature/POH charts |
| `/firmware` | Firmware Compliance | Target FW check with pie chart |
| `/alerts` | Alerts | Fired alerts + alert rule management |

## API Endpoints (Mocked)

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/devices` | List all devices |
| GET | `/api/v1/devices/:serial` | Get device by serial |
| GET | `/api/v1/devices/:serial/health/history` | Health snapshot history |
| GET | `/api/v1/firmware/compliance?target_firmware=X` | Firmware compliance check |
| GET | `/api/v1/alerts` | List fired alerts |
| GET | `/api/v1/alerts/rules` | List alert rules |
| POST | `/api/v1/alerts/rules` | Create alert rule |
| PUT | `/api/v1/alerts/rules/:id` | Update alert rule |
| DELETE | `/api/v1/alerts/rules/:id` | Delete alert rule |
