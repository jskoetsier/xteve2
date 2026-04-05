import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { api, type Channel, type Program } from '@/lib/api'
import { Search, RefreshCw, Save, Tv, Zap, Clock, ChevronRight, AlertCircle, CheckCircle } from 'lucide-react'

export default function EPG() {
  const qc = useQueryClient()
  const [selectedChannelId, setSelectedChannelId] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [mappingForm, setMappingForm] = useState({ custom_name: '', epg_channel: '', channel_num: '' })
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'saved'>('idle')

  const { data: channels, isLoading: channelsLoading } = useQuery({
    queryKey: ['channels'],
    queryFn: api.getChannels,
  })

  const { data: programs, isLoading: programsLoading } = useQuery({
    queryKey: ['epg-programs', selectedChannelId],
    queryFn: () => (selectedChannelId ? api.getEPGPrograms(selectedChannelId) : null),
    enabled: !!selectedChannelId,
  })

  const refreshEPG = useMutation({
    mutationFn: api.refreshEPG,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['epg-programs'] }),
  })

  const updateMapping = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { custom_name?: string; epg_channel?: string; channel_num?: number } }) =>
      api.updateChannelMapping(id, data),
    onSuccess: () => {
      setSaveStatus('saved')
      qc.invalidateQueries({ queryKey: ['channels'] })
      setTimeout(() => setSaveStatus('idle'), 2000)
    },
    onMutate: () => setSaveStatus('saving'),
  })

  const selectedChannel = useMemo(() => 
    channels?.find((c: Channel) => c.id === selectedChannelId),
    [channels, selectedChannelId]
  )

  const filteredChannels = useMemo(() => {
    if (!channels) return []
    if (!searchQuery) return channels
    const q = searchQuery.toLowerCase()
    return channels.filter((ch: Channel) => 
      ch.channel.name.toLowerCase().includes(q) ||
      ch.channel.tvg_id?.toLowerCase().includes(q) ||
      ch.custom_name?.toLowerCase().includes(q) ||
      ch.epg_channel?.toLowerCase().includes(q)
    )
  }, [channels, searchQuery])

  const stats = useMemo(() => {
    if (!channels) return { total: 0, mapped: 0, withEPG: 0 }
    const total = channels.length
    const mapped = channels.filter((c: Channel) => c.custom_name || c.epg_channel || c.channel_num).length
    const withEPG = channels.filter((c: Channel) => c.epg_channel).length
    return { total, mapped, withEPG }
  }, [channels])

  const handleChannelSelect = (ch: Channel) => {
    setSelectedChannelId(ch.id)
    setMappingForm({
      custom_name: ch.custom_name || '',
      epg_channel: ch.epg_channel || '',
      channel_num: ch.channel_num ? String(ch.channel_num) : '',
    })
  }

  const handleMappingSave = () => {
    if (!selectedChannelId) return
    updateMapping.mutate({
      id: selectedChannelId,
      data: {
        custom_name: mappingForm.custom_name || undefined,
        epg_channel: mappingForm.epg_channel || undefined,
        channel_num: mappingForm.channel_num ? Number(mappingForm.channel_num) : undefined,
      },
    })
  }

  const formatTime = (iso: string) => {
    try {
      const date = new Date(iso)
      return date.toLocaleTimeString('nl-NL', { hour: '2-digit', minute: '2-digit' })
    } catch {
      return iso
    }
  }

  const formatDate = (iso: string) => {
    try {
      const date = new Date(iso)
      return date.toLocaleDateString('nl-NL', { weekday: 'short', day: 'numeric', month: 'short' })
    } catch {
      return ''
    }
  }

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard 
          icon={Tv} 
          label="Total Channels" 
          value={stats.total} 
          color="blue"
        />
        <StatCard 
          icon={CheckCircle} 
          label="Mapped" 
          value={stats.mapped} 
          sublabel={`${Math.round(stats.mapped / stats.total * 100) || 0}% of total`}
          color="green"
        />
        <StatCard 
          icon={Zap} 
          label="With EPG Link" 
          value={stats.withEPG} 
          sublabel="Channels linked to XMLTV"
          color="purple"
        />
        <StatCard 
          icon={Clock} 
          label="Last Refresh" 
          value="--" 
          sublabel="Click to refresh EPG"
          color="gray"
          onClick={() => refreshEPG.mutate()}
          loading={refreshEPG.isPending}
        />
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Channel List */}
        <div className="lg:col-span-2 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-200 bg-gradient-to-r from-slate-50 to-white">
            <div className="flex items-center justify-between gap-4">
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search channels..."
                  className="w-full pl-10 pr-4 py-2.5 bg-white border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                />
              </div>
              <span className="text-sm text-gray-500">
                {filteredChannels.length} channels
              </span>
            </div>
          </div>

          <div className="divide-y divide-gray-100 max-h-[calc(100vh-380px)] overflow-y-auto">
            {channelsLoading ? (
              <div className="p-8 text-center text-gray-500">
                <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2" />
                Loading channels...
              </div>
            ) : filteredChannels.length === 0 ? (
              <div className="p-8 text-center text-gray-500">
                No channels found
              </div>
            ) : (
              filteredChannels.map((ch: Channel) => (
                <div
                  key={ch.id}
                  onClick={() => handleChannelSelect(ch)}
                  className={`p-4 cursor-pointer transition-all hover:bg-blue-50 ${
                    selectedChannelId === ch.id ? 'bg-blue-50 border-l-4 border-blue-500' : ''
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        {ch.channel.tvg_logo && (
                          <img 
                            src={ch.channel.tvg_logo} 
                            alt=""
                            className="w-8 h-8 object-contain rounded"
                            onError={(e) => e.currentTarget.style.display = 'none'}
                          />
                        )}
                        <div>
                          <p className="font-medium text-gray-900 truncate">
                            {ch.custom_name || ch.channel.name}
                          </p>
                          <p className="text-xs text-gray-500 truncate">
                            {ch.channel.name !== (ch.custom_name || '') && (
                              <span className="text-gray-400">{ch.channel.name}</span>
                            )}
                            {ch.channel.tvg_id && <span className="ml-1">• {ch.channel.tvg_id}</span>}
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {ch.channel_num > 0 && (
                        <span className="px-2 py-1 bg-gray-100 text-gray-700 text-xs font-medium rounded">
                          #{ch.channel_num}
                        </span>
                      )}
                      {ch.epg_channel && (
                        <span className="px-2 py-1 bg-green-100 text-green-700 text-xs font-medium rounded">
                          EPG
                        </span>
                      )}
                      <ChevronRight className="w-4 h-4 text-gray-400" />
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Channel Details & Mapping */}
        <div className="space-y-6">
          {/* Mapping Card */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
            <div className="p-4 border-b border-gray-200 bg-gradient-to-r from-slate-50 to-white">
              <h3 className="font-semibold text-gray-900">Channel Mapping</h3>
              <p className="text-sm text-gray-500 mt-0.5">Configure EPG and display settings</p>
            </div>

            {selectedChannel ? (
              <div className="p-4 space-y-4">
                {/* Selected channel preview */}
                <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  {selectedChannel.channel.tvg_logo && (
                    <img 
                      src={selectedChannel.channel.tvg_logo} 
                      alt=""
                      className="w-12 h-12 object-contain rounded"
                      onError={(e) => e.currentTarget.style.display = 'none'}
                    />
                  )}
                  <div>
                    <p className="font-medium text-gray-900">
                      {selectedChannel.custom_name || selectedChannel.channel.name}
                    </p>
                    <p className="text-xs text-gray-500">
                      {selectedChannel.channel.group_title || 'No group'}
                    </p>
                  </div>
                </div>

                {/* Form fields */}
                <div className="space-y-3">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                      Custom Name
                    </label>
                    <input
                      value={mappingForm.custom_name}
                      onChange={(e) => setMappingForm({ ...mappingForm, custom_name: e.target.value })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                      placeholder="Override display name"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                      EPG Channel ID
                    </label>
                    <input
                      value={mappingForm.epg_channel}
                      onChange={(e) => setMappingForm({ ...mappingForm, epg_channel: e.target.value })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                      placeholder="XMLTV channel identifier"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                      Channel Number
                    </label>
                    <input
                      type="number"
                      value={mappingForm.channel_num}
                      onChange={(e) => setMappingForm({ ...mappingForm, channel_num: e.target.value })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                      placeholder="Position in guide"
                    />
                  </div>
                </div>

                <button
                  onClick={handleMappingSave}
                  disabled={updateMapping.isPending}
                  className={`w-full py-2.5 rounded-lg font-medium text-sm transition-all flex items-center justify-center gap-2 ${
                    saveStatus === 'saved'
                      ? 'bg-green-500 text-white'
                      : 'bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50'
                  }`}
                >
                  {saveStatus === 'saving' ? (
                    <>
                      <RefreshCw className="w-4 h-4 animate-spin" />
                      Saving...
                    </>
                  ) : saveStatus === 'saved' ? (
                    <>
                      <CheckCircle className="w-4 h-4" />
                      Saved!
                    </>
                  ) : (
                    <>
                      <Save className="w-4 h-4" />
                      Save Mapping
                    </>
                  )}
                </button>
              </div>
            ) : (
              <div className="p-8 text-center">
                <div className="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-3">
                  <Tv className="w-6 h-6 text-gray-400" />
                </div>
                <p className="text-sm text-gray-500">Select a channel to configure</p>
              </div>
            )}
          </div>

          {/* Program Guide */}
          {selectedChannel && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
              <div className="p-4 border-b border-gray-200 bg-gradient-to-r from-slate-50 to-white">
                <h3 className="font-semibold text-gray-900">Program Guide</h3>
                <p className="text-sm text-gray-500 mt-0.5">Upcoming broadcasts</p>
              </div>

              <div className="max-h-80 overflow-y-auto">
                {programsLoading ? (
                  <div className="p-8 text-center text-gray-500">
                    <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2" />
                    Loading programs...
                  </div>
                ) : programs && programs.length > 0 ? (
                  <div className="divide-y divide-gray-100">
                    {programs.slice(0, 20).map((p: Program, i: number) => (
                      <div key={i} className="p-4 hover:bg-gray-50 transition-colors">
                        <div className="flex items-start gap-3">
                          <div className="text-right min-w-[60px]">
                            <p className="text-xs font-medium text-gray-500">
                              {formatDate(p.start)}
                            </p>
                            <p className="text-sm font-semibold text-blue-600">
                              {formatTime(p.start)}
                            </p>
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="font-medium text-gray-900">{p.title}</p>
                            {p.category && (
                              <span className="inline-block mt-1 px-2 py-0.5 bg-purple-100 text-purple-700 text-xs font-medium rounded">
                                {p.category}
                              </span>
                            )}
                            {p.desc && (
                              <p className="text-sm text-gray-500 mt-1 line-clamp-2">{p.desc}</p>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="p-8 text-center">
                    <div className="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-3">
                      <AlertCircle className="w-6 h-6 text-gray-400" />
                    </div>
                    <p className="text-sm text-gray-500">No program data available</p>
                    <p className="text-xs text-gray-400 mt-1">
                      Configure an XMLTV source to load program listings
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}
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
  color = 'blue',
  onClick,
  loading 
}: { 
  icon: any
  label: string
  value: number | string
  sublabel?: string
  color?: 'blue' | 'green' | 'purple' | 'gray'
  onClick?: () => void
  loading?: boolean
}) {
  const colorClasses = {
    blue: 'bg-blue-50 text-blue-600',
    green: 'bg-green-50 text-green-600',
    purple: 'bg-purple-50 text-purple-600',
    gray: 'bg-gray-50 text-gray-600',
  }

  return (
    <div 
      onClick={onClick}
      className={`bg-white rounded-xl shadow-sm border border-gray-200 p-5 ${onClick ? 'cursor-pointer hover:shadow-md transition-shadow' : ''}`}
    >
      <div className="flex items-center gap-4">
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${colorClasses[color]}`}>
          <Icon className={`w-6 h-6 ${loading ? 'animate-spin' : ''}`} />
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
