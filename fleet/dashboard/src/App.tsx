import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Layout from './components/Layout';
import FleetOverview from './pages/FleetOverview';
import DeviceList from './pages/DeviceList';
import DeviceDetail from './pages/DeviceDetail';
import FirmwareCompliance from './pages/FirmwareCompliance';
import Alerts from './pages/Alerts';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 2,
      refetchOnWindowFocus: false,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<FleetOverview />} />
            <Route path="/devices" element={<DeviceList />} />
            <Route path="/devices/:serial" element={<DeviceDetail />} />
            <Route path="/firmware" element={<FirmwareCompliance />} />
            <Route path="/alerts" element={<Alerts />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
