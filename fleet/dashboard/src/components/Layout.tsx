import { useState, useEffect } from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { useAlerts } from '../hooks/useAlerts';
import { toggleTheme, isDarkMode } from '../theme';

const NAV_ITEMS = [
  { to: '/', label: 'Fleet Overview', icon: '⊞' },
  { to: '/devices', label: 'Devices', icon: '⊟' },
  { to: '/firmware', label: 'Firmware Compliance', icon: '⊠' },
  { to: '/alerts', label: 'Alerts', icon: '⊡' },
];

export default function Layout() {
  const [dark, setDark] = useState(isDarkMode);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const { data: alertData } = useAlerts({ status: 'firing' });
  const firingCount = alertData?.total ?? 0;

  useEffect(() => {
    setDark(isDarkMode());
  }, []);

  function handleToggleTheme() {
    toggleTheme();
    setDark(isDarkMode());
  }

  return (
    <div className="flex h-screen bg-gray-50 dark:bg-gray-900">
      {/* Sidebar */}
      <aside
        className={`${
          sidebarOpen ? 'w-60' : 'w-16'
        } flex-shrink-0 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 transition-all duration-200 flex flex-col`}
      >
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-primary-600 rounded flex items-center justify-center text-white font-bold text-sm">
              SC
            </div>
            {sidebarOpen && (
              <div>
                <h1 className="text-sm font-semibold text-gray-900 dark:text-white">
                  Fleet Dashboard
                </h1>
                <p className="text-xs text-gray-500 dark:text-gray-400">openSeaChest</p>
              </div>
            )}
          </div>
        </div>

        <nav className="flex-1 p-2">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm mb-0.5 transition-colors ${
                  isActive
                    ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-400 font-medium'
                    : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
                }`
              }
            >
              <span className="text-lg">{item.icon}</span>
              {sidebarOpen && (
                <span className="flex-1">{item.label}</span>
              )}
              {item.to === '/alerts' && firingCount > 0 && sidebarOpen && (
                <span className="bg-danger-500 text-white text-xs font-medium px-1.5 py-0.5 rounded-full">
                  {firingCount}
                </span>
              )}
            </NavLink>
          ))}
        </nav>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-3 flex items-center justify-between">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>

          <div className="flex items-center gap-3">
            <button
              onClick={handleToggleTheme}
              className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400"
              title={dark ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {dark ? (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
              ) : (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                </svg>
              )}
            </button>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              Fleet Admin
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
