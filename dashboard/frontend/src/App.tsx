import { BrowserRouter, Routes, Route } from "react-router-dom";
import Layout from "./components/Layout/Layout";
import FleetInventory from "./components/FleetInventory/FleetInventory";
import DeviceDetail from "./components/DeviceDetail/DeviceDetail";
import FirmwareCompliance from "./components/FirmwareCompliance/FirmwareCompliance";
import Alerts from "./components/Alerts/Alerts";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<FleetInventory />} />
          <Route path="/devices/:serial" element={<DeviceDetail />} />
          <Route path="/firmware" element={<FirmwareCompliance />} />
          <Route path="/alerts" element={<Alerts />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
