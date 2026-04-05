import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { api } from '@/lib/api'
import { Save, RefreshCw, ExternalLink, CheckCircle, AlertCircle, Globe, Clock } from 'lucide-react'

export default function Playlists() {
  const qc = useQueryClient()
  const [form, setForm] = useState({ m3u_url: '', xmltv_url: '', m3u_refresh_mins: 15 })

  const { data: settings, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: api.getSettings,
  })

  useEffect(() => {
    if (settings) {
      setForm({
        m3u_url: settings.m3u_url || '',
        xmltv_url: settings.xmltv_url || '',
        m3u_refresh_mins: settings.m3u_refresh_mins || 15,
      })
    }
  }, [settings])

  const save = useMutation({
    mutationFn: () => api.putSettings(form),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })

  const refreshPlaylist = useMutation({
    mutationFn: api.refreshPlaylist,
  })

  const refreshEPG = useMutation({
    mutationFn: api.refreshEPG,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold text-gray-900">Playlists & EPG</h1>
        <p className="text-sm text-gray-500 mt-0.5">Configure M3U and XMLTV sources for IPTV streaming</p>
      </div>

      {/* M3U Playlist Card */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-blue-50 to-white">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
              <Globe className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">M3U Playlist</h3>
              <p className="text-sm text-gray-500">IPTV channel source</p>
            </div>
          </div>
        </div>

        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Playlist URL</label>
            <input
              value={form.m3u_url}
              onChange={(e) => setForm({ ...form, m3u_url: e.target.value })}
              className="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="https://example.com/playlist.m3u"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Refresh Interval (minutes)</label>
            <input
              type="number"
              value={form.m3u_refresh_mins}
              onChange={(e) => setForm({ ...form, m3u_refresh_mins: Number(e.target.value) })}
              className="w-40 px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              min="1"
            />
          </div>

          <div className="flex items-center gap-3 pt-2">
            <button
              onClick={() => save.mutate()}
              disabled={save.isPending}
              className="flex items-center gap-2 px-4 py-2.5 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 transition-colors"
            >
              {save.isPending ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <Save className="w-4 h-4" />
              )}
              Save Settings
            </button>

            <button
              onClick={() => refreshPlaylist.mutate()}
              disabled={refreshPlaylist.isPending || !form.m3u_url}
              className="flex items-center gap-2 px-4 py-2.5 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 transition-colors"
            >
              {refreshPlaylist.isPending ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <RefreshCw className="w-4 h-4" />
              )}
              Refresh Now
            </button>
          </div>

          {save.isSuccess && (
            <div className="flex items-center gap-2 text-green-600 text-sm">
              <CheckCircle className="w-4 h-4" />
              Settings saved successfully
            </div>
          )}

          {refreshPlaylist.isSuccess && (
            <div className="flex items-center gap-2 text-green-600 text-sm">
              <CheckCircle className="w-4 h-4" />
              Playlist refreshed successfully
            </div>
          )}

          {refreshPlaylist.isError && (
            <div className="flex items-center gap-2 text-red-600 text-sm">
              <AlertCircle className="w-4 h-4" />
              Failed to refresh playlist
            </div>
          )}
        </div>
      </div>

      {/* XMLTV EPG Card */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-purple-50 to-white">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
              <Clock className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">XMLTV EPG Source</h3>
              <p className="text-sm text-gray-500">Electronic Program Guide data</p>
            </div>
          </div>
        </div>

        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">XMLTV URL</label>
            <input
              value={form.xmltv_url}
              onChange={(e) => setForm({ ...form, xmltv_url: e.target.value })}
              className="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-purple-500"
              placeholder="https://example.com/xmltv.xml"
            />
          </div>

          <div className="flex items-center gap-3 pt-2">
            <button
              onClick={() => save.mutate()}
              disabled={save.isPending}
              className="flex items-center gap-2 px-4 py-2.5 bg-purple-500 text-white rounded-lg hover:bg-purple-600 disabled:opacity-50 transition-colors"
            >
              {save.isPending ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <Save className="w-4 h-4" />
              )}
              Save Settings
            </button>

            <button
              onClick={() => refreshEPG.mutate()}
              disabled={refreshEPG.isPending || !form.xmltv_url}
              className="flex items-center gap-2 px-4 py-2.5 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 transition-colors"
            >
              {refreshEPG.isPending ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <RefreshCw className="w-4 h-4" />
              )}
              Refresh EPG
            </button>
          </div>

          {refreshEPG.isSuccess && (
            <div className="flex items-center gap-2 text-green-600 text-sm">
              <CheckCircle className="w-4 h-4" />
              EPG refreshed successfully
            </div>
          )}

          {refreshEPG.isError && (
            <div className="flex items-center gap-2 text-red-600 text-sm">
              <AlertCircle className="w-4 h-4" />
              Failed to refresh EPG
            </div>
          )}
        </div>
      </div>

      {/* Info Box */}
      <div className="bg-gradient-to-r from-gray-50 to-slate-50 rounded-xl border border-gray-200 p-6">
        <h4 className="font-medium text-gray-900 mb-2">About Playlists</h4>
        <p className="text-sm text-gray-600">
          xTeVe supports a single M3U playlist URL and a single XMLTV URL. Enter your IPTV subscription URLs above, 
          save the settings, and then click "Refresh Now" to fetch the channels and programme data. 
          The EPG data will be used to populate programme information for each channel in the guide.
        </p>
        <div className="mt-4 flex items-center gap-4 text-xs text-gray-500">
          <span>Auto-refresh every {form.m3u_refresh_mins} minutes</span>
          {form.m3u_url && (
            <a 
              href={form.m3u_url} 
              target="_blank" 
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-blue-500 hover:text-blue-600"
            >
              <ExternalLink className="w-3 h-3" />
              View current playlist
            </a>
          )}
        </div>
      </div>
    </div>
  )
}
