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

export const api = {
  status: () => request<{ version: string; active_streams: number; tuner_count: number }>('/status'),
  getSettings: () => request<Record<string, unknown>>('/settings'),
  putSettings: (data: unknown) => request('/settings', { method: 'PUT', body: JSON.stringify(data) }),
  getChannels: () => request<unknown[]>('/channels'),
  putChannel: (id: string, data: unknown) =>
    request(`/channels/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  login: (password: string) =>
    request('/auth/login', { method: 'POST', body: JSON.stringify({ password }) }),
  logout: () => request('/auth/logout', { method: 'POST' }),
}
