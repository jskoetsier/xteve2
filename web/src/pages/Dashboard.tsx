import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Activity, Tv, Zap, Server, RefreshCw, Clock } from 'lucide-react'
import { useMemo } from 'react'

export default function Dashboard() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['status'],
    queryFn: api.status,
    refetchInterval: 5000,
  })

  const { data: channels } = useQuery({
    queryKey: ['channels'],
    queryFn: api.getChannels,
  })

  const channelStats = useMemo(() => {
    if (!channels) return { total: 0, enabled: 0, mapped: 0 }
    const total = channels.length
    const enabled = channels.filter((c: any) => c.enabled).length
    const mapped = channels.filter((c: any) => c.custom_name || c.epg_channel || c.channel_num).length
    return { total, enabled, mapped }
  }, [channels])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
        <p className="text-red-600 font-medium">Failed to load server status</p>
        <p className="text-sm text-red-500 mt-1">Check your connection and try again</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          icon={Server}
          label="Server Version"
          value={data?.version ?? '—'}
          sublabel="xTeVe 2.0"
          color="blue"
        />
        <StatCard
          icon={Tv}
          label="Total Channels"
          value={channelStats.total}
          sublabel={`${channelStats.enabled} enabled`}
          color="purple"
        />
        <StatCard
          icon={Activity}
          label="Active Streams"
          value={`${data?.active_streams ?? 0}`}
          sublabel={`of ${data?.tuner_count ?? 0} tuners`}
          color={data?.active_streams && data.active_streams > 0 ? 'green' : 'gray'}
        />
        <StatCard
          icon={Zap}
          label="Channel Mappings"
          value={channelStats.mapped}
          sublabel={`${channelStats.total > 0 ? Math.round(channelStats.mapped / channelStats.total * 100) : 0}% configured`}
          color="orange"
        />
      </div>

      {/* Quick Actions & Info */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Server Info */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100 bg-gradient-to-r from-slate-50 to-white">
            <h3 className="font-semibold text-gray-900">Server Information</h3>
          </div>
          <div className="p-6">
            <div className="space-y-4">
              <InfoRow label="Version" value={data?.version ?? '—'} />
              <InfoRow label="Tuner Count" value={`${data?.tuner_count ?? 0}`} />
              <InfoRow label="Active Streams" value={`${data?.active_streams ?? 0}`} />
              <InfoRow label="Server Time" value={new Date().toLocaleTimeString()} />
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100 bg-gradient-to-r from-slate-50 to-white">
            <h3 className="font-semibold text-gray-900">Quick Actions</h3>
          </div>
          <div className="p-6">
            <div className="space-y-3">
              <button 
                onClick={() => refetch()}
                className="w-full flex items-center justify-between p-4 bg-gray-50 hover:bg-gray-100 rounded-lg transition-colors group"
              >
                <div className="flex items-center gap-3">
                  <RefreshCw className="w-5 h-5 text-gray-500" />
                  <span className="font-medium text-gray-700">Refresh Status</span>
                </div>
                <span className="text-gray-400 group-hover:translate-x-1 transition-transform">→</span>
              </button>
              
              <a 
                href="/playlists"
                className="block w-full flex items-center justify-between p-4 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors group"
              >
                <div className="flex items-center gap-3">
                  <Tv className="w-5 h-5 text-blue-500" />
                  <span className="font-medium text-blue-700">Configure Playlists</span>
                </div>
                <span className="text-blue-400 group-hover:translate-x-1 transition-transform">→</span>
              </a>

              <a 
                href="/epg"
                className="block w-full flex items-center justify-between p-4 bg-purple-50 hover:bg-purple-100 rounded-lg transition-colors group"
              >
                <div className="flex items-center gap-3">
                  <Activity className="w-5 h-5 text-purple-500" />
                  <span className="font-medium text-purple-700">Manage EPG Mappings</span>
                </div>
                <span className="text-purple-400 group-hover:translate-x-1 transition-transform">→</span>
              </a>
            </div>
          </div>
        </div>
      </div>

      {/* System Status */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-100 bg-gradient-to-r from-slate-50 to-white">
          <h3 className="font-semibold text-gray-900">System Status</h3>
        </div>
        <div className="p-6">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-sm font-medium text-green-700">All Systems Operational</span>
            </div>
            <div className="flex-1"></div>
            <div className="flex items-center gap-2 text-sm text-gray-500">
              <Clock className="w-4 h-4" />
              <span>Last checked: {new Date().toLocaleTimeString()}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function StatCard({ 
  icon: Icon, 
  label, 
  value, 
  sublabel,
  color = 'blue' 
}: { 
  icon: any
  label: string
  value: number | string
  sublabel?: string
  color?: 'blue' | 'green' | 'purple' | 'orange' | 'gray'
}) {
  const colorClasses = {
    blue: 'bg-blue-50 text-blue-600',
    green: 'bg-green-50 text-green-600',
    purple: 'bg-purple-50 text-purple-600',
    orange: 'bg-orange-50 text-orange-600',
    gray: 'bg-gray-50 text-gray-600',
  }

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-5 hover:shadow-md transition-shadow">
      <div className="flex items-center gap-4">
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${colorClasses[color]}`}>
          <Icon className="w-6 h-6" />
        </div>
        <div>
          <p className="text-sm text-gray-500">{label}</p>
          <p className="text-2xl font-bold text-gray-900">{value}</p>
          {sublabel && <p className="text-xs text-gray-400 mt-0.5">{sublabel}</p>}
        </div>
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0">
      <span className="text-sm text-gray-500">{label}</span>
      <span className="text-sm font-medium text-gray-900">{value}</span>
    </div>
  )
}
