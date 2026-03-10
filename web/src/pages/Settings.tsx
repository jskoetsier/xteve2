import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { api } from '@/lib/api'

export default function Settings() {
  const qc = useQueryClient()
  const { data, isLoading } = useQuery({ queryKey: ['settings'], queryFn: api.getSettings })
  const [form, setForm] = useState<Record<string, unknown>>({})

  useEffect(() => {
    if (data) setForm(data)
  }, [data])

  const save = useMutation({
    mutationFn: () => api.putSettings(form),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })

  if (isLoading) return <p>Loading...</p>

  return (
    <div className="max-w-lg">
      <h1 className="text-2xl font-bold mb-6">Settings</h1>
      <div className="space-y-4">
        <Field label="Port" value={String(form.port ?? '')}
          onChange={(v) => setForm({ ...form, port: Number(v) })} />
        <Field label="Tuner Count" value={String(form.tuner_count ?? '')}
          onChange={(v) => setForm({ ...form, tuner_count: Number(v) })} />
        <Field label="FFmpeg Path" value={String(form.ffmpeg_path ?? '')}
          onChange={(v) => setForm({ ...form, ffmpeg_path: v })} />
        <Field label="Buffer Type" value={String(form.buffer_type ?? '')}
          onChange={(v) => setForm({ ...form, buffer_type: v })} />
        <button
          onClick={() => save.mutate()}
          className="px-4 py-2 bg-slate-900 text-white rounded text-sm hover:bg-slate-700"
        >
          {save.isPending ? 'Saving...' : 'Save Settings'}
        </button>
      </div>
    </div>
  )
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-700 mb-1">{label}</label>
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full border rounded px-3 py-2 text-sm"
      />
    </div>
  )
}
