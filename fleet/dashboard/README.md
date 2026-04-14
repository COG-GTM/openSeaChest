# Fleet Dashboard

A React-based web dashboard for visualizing fleet inventory, health status, firmware compliance, and alerts for storage devices managed by openSeaChest.

## Tech Stack

- **React 18** with TypeScript
- **Vite** for development and build
- **TanStack React Query** for data fetching and caching
- **Recharts** for charts and visualizations
- **Tailwind CSS** for styling (light/dark mode)
- **MSW (Mock Service Worker)** for API mocking during development

## Development Setup

### Prerequisites

- Node.js 18+
- npm 9+

### Install Dependencies

```bash
cd fleet/dashboard
npm install
```

### Start Development Server

```bash
npm run dev
```

The dashboard will be available at [http://localhost:3000](http://localhost:3000).

In development mode, MSW (Mock Service Worker) is enabled by default, providing realistic mock data for all API endpoints. No backend server is required.

### Build for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

### Preview Production Build

```bash
npm run preview
```

### Lint

```bash
npm run lint
```

### Type Check

```bash
npm run typecheck
```

## Mock API

MSW is enabled automatically in development mode. The mock handlers (`src/mocks/handlers.ts`) intercept all API calls and return realistic data:

- 75 mock devices across 10 hosts with varied health statuses and firmware versions
- Time-series health data (30 days of 15-minute snapshots) for chart testing
- Mock firmware policies and compliance reports
- Mock alerts with various severities and statuses
- Mock FARM reliability metrics modeled after real Seagate device output

Mock data is seeded deterministically, so the data is consistent across page reloads.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VITE_API_BASE_URL` | `/api` | Base URL for the fleet REST API |

## Pages

### Fleet Overview (`/`)
Dashboard home page with stat cards, health status pie chart, firmware compliance bar chart, and recent alerts.

### Devices (`/devices`)
Filterable, sortable device table with server-side pagination. Supports filtering by model, firmware, interface type, host, health status, and compliance status. Includes CSV export.

### Device Detail (`/devices/:serial`)
Single device view showing device information, current health snapshot, health history charts (temperature, power-on hours, workload rate) with 30d/90d/365d time range selection, FARM reliability metrics (for Seagate drives), firmware compliance status, and alert history.

### Firmware Compliance (`/firmware`)
Compliance summary table grouped by model with progress bars. Click a model to drill down into non-compliant devices. Lists firmware policies.

### Alerts (`/alerts`)
Alert feed with filtering by severity, device, host, and status. Acknowledge button on firing alerts opens a modal for annotation.

## Docker

Build and run with Docker:

```bash
docker build -t fleet-dashboard .
docker run -p 80:80 fleet-dashboard
```

The nginx configuration proxies `/api/*` requests to the `fleet-api` backend service on port 8080.

## Project Structure

```
src/
  main.tsx              # Entry point with MSW initialization
  App.tsx               # Router and provider setup
  index.css             # Tailwind CSS imports
  api/
    client.ts           # Axios-based API client
    types.ts            # TypeScript interfaces for API schemas
  pages/
    FleetOverview.tsx   # Dashboard home page
    DeviceList.tsx      # Filterable device table
    DeviceDetail.tsx    # Single device view with charts
    FirmwareCompliance.tsx # Compliance report view
    Alerts.tsx          # Alert feed with acknowledge/resolve
  components/
    Layout.tsx          # App shell with sidebar navigation
    DeviceTable.tsx     # Sortable/filterable device table
    HealthChart.tsx     # Time-series line chart (Recharts)
    StatusBadge.tsx     # Health/compliance status badges
    PieChart.tsx        # Fleet health breakdown pie chart
    Pagination.tsx      # Pagination controls
    FilterBar.tsx       # Filter controls
  hooks/
    useDevices.ts       # React Query hooks for device data
    useHealth.ts        # React Query hooks for health data
    useAlerts.ts        # React Query hooks for alerts
  theme/
    index.ts            # Light/dark mode theme config
  mocks/
    handlers.ts         # MSW mock handlers for all API endpoints
    data.ts             # Mock data fixtures
    browser.ts          # MSW browser worker setup
```
