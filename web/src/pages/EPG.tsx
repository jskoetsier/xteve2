import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { api, type Channel, type Program } from '@/lib/api'

export default function EPG() {
  const qc = useQueryClient()
  const [selectedChannel, setSelectedChannel] = useState<string | null>(null)
  const [mappingForm, setMappingForm] = useState({ custom_name: '', epg_channel: '', channel_num: '' })

  const { data: channels } = useQuery({
    queryKey: ['channels'],
    queryFn: api.getChannels,
  })

  const { data: programs } = useQuery({
    queryKey: ['epg-programs', selectedChannel],
    queryFn: () => (selectedChannel ? api.getEPGPrograms(selectedChannel) : null),
    enabled: !!selectedChannel,
  })

  const refreshEPG = useMutation({
    mutationFn: api.refreshEPG,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['epg-programs'] }),
  })

  const updateMapping = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { custom_name?: string; epg_channel?: string; channel_num?: number } }) =>
      api.updateChannelMapping(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['channels'] })
    },
  })

  const handleChannelSelect = (ch: Channel) => {
    setSelectedChannel(ch.id)
    setMappingForm({
      custom_name: ch.custom_name || '',
      epg_channel: ch.epg_channel || '',
      channel_num: ch.channel_num ? String(ch.channel_num) : '',
    })
  }

  const handleMappingSave = () => {
    if (!selectedChannel) return
    updateMapping.mutate({
      id: selectedChannel,
      data: {
        custom_name: mappingForm.custom_name || undefined,
        epg_channel: mappingForm.epg_channel || undefined,
        channel_num: mappingForm.channel_num ? Number(mappingForm.channel_num) : undefined,
      },
    })
  }

  const formatTime = (iso: string) => {
    try {
      return new Date(iso).toLocaleString()
    } catch {
      return iso
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">EPG</h1>
        <button
          onClick={() => refreshEPG.mutate()}
          disabled={refreshEPG.isPending}
          className="px-4 py-2 bg-slate-900 text-white rounded text-sm hover:bg-slate-700 disabled:opacity-50"
        >
          {refreshEPG.isPending ? 'Refreshing...' : 'Refresh EPG'}
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-4">Channels</h2>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {channels?.map((ch) => (
              <div
                key={ch.id}
                onClick={() => handleChannelSelect(ch)}
                className={`p-3 rounded cursor-pointer border ${
                  selectedChannel === ch.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:bg-gray-50'
                }`}
              >
                <div className="font-medium">{ch.custom_name || ch.channel.name}</div>
                <div className="text-xs text-gray-500">
                  {ch.channel.tvg_id && <span>tvg-id: {ch.channel.tvg_id}</span>}
                  {ch.epg_channel && <span> | EPG: {ch.epg_channel}</span>}
                </div>
              </div>
            ))}
            {channels?.length === 0 && <p className="text-gray-500 text-sm">No channels found. Add an M3U playlist first.</p>}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-4">Channel Mapping</h2>
          {selectedChannel ? (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Custom Name</label>
                <input
                  value={mappingForm.custom_name}
                  onChange={(e) => setMappingForm({ ...mappingForm, custom_name: e.target.value })}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="Override display name"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">EPG Channel ID</label>
                <input
                  value={mappingForm.epg_channel}
                  onChange={(e) => setMappingForm({ ...mappingForm, epg_channel: e.target.value })}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="XMLTV channel ID to map to"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Channel Number</label>
                <input
                  type="number"
                  value={mappingForm.channel_num}
                  onChange={(e) => setMappingForm({ ...mappingForm, channel_num: e.target.value })}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="Channel number"
                />
              </div>
              <button
                onClick={handleMappingSave}
                disabled={updateMapping.isPending}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {updateMapping.isPending ? 'Saving...' : 'Save Mapping'}
              </button>
            </div>
          ) : (
            <p className="text-gray-500 text-sm">Select a channel to configure its EPG mapping.</p>
          )}
        </div>
      </div>

      {programs && programs.length > 0 && (
        <div className="mt-6 bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-4">Program Guide</h2>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {programs.map((p: Program, i: number) => (
              <div key={i} className="border-b border-gray-100 pb-2 mb-2">
                <div className="font-medium">{p.title}</div>
                <div className="text-xs text-gray-500">
                  {formatTime(p.start)} — {formatTime(p.stop)}
                  {p.category && <span className="ml-2 text-blue-600">{p.category}</span>}
                  {p.episode && <span className="ml-2 text-gray-600">[{p.episode}]</span>}
                </div>
                {p.desc && <p className="text-sm text-gray-600 mt-1 line-clamp-2">{p.desc}</p>}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
