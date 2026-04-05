import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { api } from '@/lib/api'

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

  const handleSave = () => {
    save.mutate()
  }

  if (isLoading) return <p>Loading...</p>

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Playlists</h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-4">M3U Playlist</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">M3U URL</label>
              <input
                value={form.m3u_url}
                onChange={(e) => setForm({ ...form, m3u_url: e.target.value })}
                className="w-full border rounded px-3 py-2 text-sm"
                placeholder="https://example.com/playlist.m3u"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Refresh Interval (minutes)</label>
              <input
                type="number"
                value={form.m3u_refresh_mins}
                onChange={(e) => setForm({ ...form, m3u_refresh_mins: Number(e.target.value) })}
                className="w-full border rounded px-3 py-2 text-sm"
                min="1"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleSave}
                disabled={save.isPending}
                className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
              >
                {save.isPending ? 'Saving...' : 'Save'}
              </button>
              <button
                onClick={() => refreshPlaylist.mutate()}
                disabled={refreshPlaylist.isPending || !form.m3u_url}
                className="px-4 py-2 bg-slate-900 text-white rounded text-sm hover:bg-slate-700 disabled:opacity-50"
              >
                {refreshPlaylist.isPending ? 'Refreshing...' : 'Refresh Now'}
              </button>
            </div>
            {refreshPlaylist.isSuccess && (
              <p className="text-sm text-green-600">Playlist refreshed successfully.</p>
            )}
            {refreshPlaylist.isError && (
              <p className="text-sm text-red-600">Failed to refresh playlist.</p>
            )}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-4">
          <h2 className="text-lg font-semibold mb-4">XMLTV EPG Source</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">XMLTV URL</label>
              <input
                value={form.xmltv_url}
                onChange={(e) => setForm({ ...form, xmltv_url: e.target.value })}
                className="w-full border rounded px-3 py-2 text-sm"
                placeholder="https://example.com/xmltv.xml"
              />
            </div>
            <button
              onClick={() => refreshEPG.mutate()}
              disabled={refreshEPG.isPending || !form.xmltv_url}
              className="px-4 py-2 bg-slate-900 text-white rounded text-sm hover:bg-slate-700 disabled:opacity-50"
            >
              {refreshEPG.isPending ? 'Refreshing...' : 'Refresh EPG'}
            </button>
            {refreshEPG.isSuccess && (
              <p className="text-sm text-green-600">EPG refreshed successfully.</p>
            )}
            {refreshEPG.isError && (
              <p className="text-sm text-red-600">Failed to refresh EPG.</p>
            )}
          </div>
        </div>
      </div>

      <div className="mt-6 bg-white rounded-lg shadow p-4">
        <h2 className="text-lg font-semibold mb-4">About Playlists</h2>
        <p className="text-sm text-gray-600">
          xTeVe supports a single M3U playlist URL. Enter your IPTV subscription M3U URL above,
          save the settings, and then click &quot;Refresh Now&quot; to fetch the channels.
          The EPG data will be used to populate programme information for each channel.
        </p>
      </div>
    </div>
  )
}
