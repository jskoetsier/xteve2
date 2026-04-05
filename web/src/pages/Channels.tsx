import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api, type Channel } from '@/lib/api'
import { Search, RefreshCw, Tv, ToggleLeft, ToggleRight, Filter } from 'lucide-react'
import { useState, useMemo } from 'react'

export default function Channels() {
  const qc = useQueryClient()
  const [searchQuery, setSearchQuery] = useState('')
  const [groupFilter, setGroupFilter] = useState<string | null>(null)

  const { data: channels = [], isLoading, refetch } = useQuery({
    queryKey: ['channels'],
    queryFn: api.getChannels,
  })

  const toggle = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api.putChannel(id, { enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['channels'] }),
  })

  const groups = useMemo(() => {
    if (!channels) return []
    const groupSet = new Set<string>()
    channels.forEach((ch: Channel) => {
      if (ch.channel.group_title) groupSet.add(ch.channel.group_title)
    })
    return Array.from(groupSet).sort()
  }, [channels])

  const filteredChannels = useMemo(() => {
    if (!channels) return []
    let filtered = channels
    if (searchQuery) {
      const q = searchQuery.toLowerCase()
      filtered = filtered.filter((ch: Channel) => 
        ch.channel.name.toLowerCase().includes(q) ||
        ch.channel.tvg_id?.toLowerCase().includes(q) ||
        ch.custom_name?.toLowerCase().includes(q)
      )
    }
    if (groupFilter) {
      filtered = filtered.filter((ch: Channel) => ch.channel.group_title === groupFilter)
    }
    return filtered
  }, [channels, searchQuery, groupFilter])

  const enabledCount = useMemo(() => 
    channels.filter((ch: Channel) => ch.enabled).length,
    [channels]
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">Channels</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            {enabledCount} of {channels.length} channels enabled
          </p>
        </div>
        <button
          onClick={() => refetch()}
          className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
        >
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search channels..."
            className="w-full pl-10 pr-4 py-2.5 bg-white border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-gray-400" />
          <select
            value={groupFilter || ''}
            onChange={(e) => setGroupFilter(e.target.value || null)}
            className="px-3 py-2.5 bg-white border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Groups</option>
            {groups.map((g: string) => (
              <option key={g} value={g}>{g}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-gradient-to-r from-slate-50 to-white border-b border-gray-200">
                <th className="text-left px-6 py-4 text-xs font-semibold text-gray-500 uppercase tracking-wider">Status</th>
                <th className="text-left px-6 py-4 text-xs font-semibold text-gray-500 uppercase tracking-wider">Channel</th>
                <th className="text-left px-6 py-4 text-xs font-semibold text-gray-500 uppercase tracking-wider">Group</th>
                <th className="text-left px-6 py-4 text-xs font-semibold text-gray-500 uppercase tracking-wider">TVG ID</th>
                <th className="text-left px-6 py-4 text-xs font-semibold text-gray-500 uppercase tracking-wider">Custom Name</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {isLoading ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center">
                    <RefreshCw className="w-6 h-6 animate-spin mx-auto text-gray-400" />
                    <p className="mt-2 text-sm text-gray-500">Loading channels...</p>
                  </td>
                </tr>
              ) : filteredChannels.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center">
                    <Tv className="w-12 h-12 mx-auto text-gray-300" />
                    <p className="mt-2 text-sm text-gray-500">No channels found</p>
                  </td>
                </tr>
              ) : (
                filteredChannels.map((ch: Channel) => (
                  <tr key={ch.id} className="hover:bg-blue-50/50 transition-colors">
                    <td className="px-6 py-4">
                      <button
                        onClick={() => toggle.mutate({ id: ch.id, enabled: !ch.enabled })}
                        className="group"
                      >
                        {ch.enabled ? (
                          <ToggleRight className="w-8 h-8 text-green-500 group-hover:text-green-600" />
                        ) : (
                          <ToggleLeft className="w-8 h-8 text-gray-300 group-hover:text-gray-400" />
                        )}
                      </button>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        {ch.channel.tvg_logo && (
                          <img 
                            src={ch.channel.tvg_logo} 
                            alt=""
                            className="w-10 h-10 object-contain rounded bg-gray-50"
                            onError={(e) => e.currentTarget.style.display = 'none'}
                          />
                        )}
                        <span className="font-medium text-gray-900">
                          {ch.custom_name || ch.channel.name}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
                        {ch.channel.group_title || 'Ungrouped'}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <code className="text-xs font-mono text-gray-500 bg-gray-50 px-2 py-1 rounded">
                        {ch.channel.tvg_id || '—'}
                      </code>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-sm text-gray-500">
                        {ch.custom_name ? (
                          <span className="text-blue-600">{ch.custom_name}</span>
                        ) : '—'}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 bg-gradient-to-r from-slate-50 to-white">
          <p className="text-sm text-gray-500">
            Showing {filteredChannels.length} of {channels.length} channels
            {groupFilter && ` in "${groupFilter}"`}
          </p>
        </div>
      </div>
    </div>
  )
}
