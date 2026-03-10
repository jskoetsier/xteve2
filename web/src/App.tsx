import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Dashboard from './pages/Dashboard'
import Channels from './pages/Channels'
import Playlists from './pages/Playlists'
import EPG from './pages/EPG'
import Settings from './pages/Settings'
import Logs from './pages/Logs'

const queryClient = new QueryClient()

const navItems = [
  { to: '/', label: 'Dashboard' },
  { to: '/channels', label: 'Channels' },
  { to: '/playlists', label: 'Playlists' },
  { to: '/epg', label: 'EPG' },
  { to: '/settings', label: 'Settings' },
  { to: '/logs', label: 'Logs' },
]

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="flex h-screen">
          <nav className="w-48 bg-slate-900 text-white flex flex-col gap-1 p-4">
            <span className="font-bold text-lg mb-4">xTeVe</span>
            {navItems.map(({ to, label }) => (
              <NavLink
                key={to}
                to={to}
                end={to === '/'}
                className={({ isActive }) =>
                  `px-3 py-2 rounded text-sm ${isActive ? 'bg-slate-700' : 'hover:bg-slate-800'}`
                }
              >
                {label}
              </NavLink>
            ))}
          </nav>
          <main className="flex-1 overflow-auto p-6">
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/channels" element={<Channels />} />
              <Route path="/playlists" element={<Playlists />} />
              <Route path="/epg" element={<EPG />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="/logs" element={<Logs />} />
            </Routes>
          </main>
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
