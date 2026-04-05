const BASE = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  if (res.status === 204) return undefined as T
  return res.json()
}

export interface Channel {
  id: string
  channel: {
    name: string
    tvg_id: string
    tvg_name: string
    tvg_logo: string
    group_title: string
    url: string
  }
  enabled: boolean
  custom_name: string
  epg_channel: string
  channel_num: number
}

export interface Program {
  channel: string
  start: string
  stop: string
  title: string
  desc: string
  category: string
  icon: string
  episode: string
}

export interface Settings {
  port: number
  tuner_count: number
  auth_enabled: boolean
  auth_password: string
  ffmpeg_path: string
  vlc_path: string
  buffer_type: string
  epg_refresh_hour: number
  m3u_url: string
  xmltv_url: string
  m3u_refresh_mins: number
}

export const api = {
  status: () =>
    request<{ version: string; active_streams: number; tuner_count: number }>('/status'),

  getSettings: () => request<Settings>('/settings'),

  putSettings: (data: unknown) =>
    request('/settings', { method: 'PUT', body: JSON.stringify(data) }),

  getChannels: () => request<Channel[]>('/channels'),

  putChannel: (id: string, data: Partial<Channel>) =>
    request(`/channels/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  login: (password: string) =>
    request('/auth/login', { method: 'POST', body: JSON.stringify({ password }) }),

  logout: () => request('/auth/logout', { method: 'POST' }),

  refreshPlaylist: () =>
    request<{ status: string }>('/playlists/refresh', { method: 'POST' }),

  refreshEPG: () =>
    request<{ status: string }>('/epg/refresh', { method: 'POST' }),

  getEPGPrograms: (channelId: string) =>
    request<Program[]>(`/epg/programs?channel_id=${encodeURIComponent(channelId)}`),

  updateChannelMapping: (id: string, data: { custom_name?: string; epg_channel?: string; channel_num?: number }) =>
    request(`/channels/${id}/mapping`, { method: 'PUT', body: JSON.stringify(data) }),
}
