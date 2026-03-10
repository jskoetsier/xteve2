import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'

interface Channel {
  id: string
  channel: { name: string; tvg_id: string; group_title: string; url: string }
  enabled: boolean
  custom_name?: string
}

export default function Channels() {
  const qc = useQueryClient()
  const { data: channels = [], isLoading } = useQuery({
    queryKey: ['channels'],
    queryFn: api.getChannels as () => Promise<Channel[]>,
  })

  const toggle = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api.putChannel(id, { enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['channels'] }),
  })

  if (isLoading) return <p>Loading...</p>

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Channels</h1>
      <div className="border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b">
            <tr>
              <th className="text-left p-3">Enabled</th>
              <th className="text-left p-3">Name</th>
              <th className="text-left p-3">Group</th>
              <th className="text-left p-3">TVG ID</th>
            </tr>
          </thead>
          <tbody>
            {channels.map((ch) => (
              <tr key={ch.id} className="border-b hover:bg-slate-50">
                <td className="p-3">
                  <input
                    type="checkbox"
                    checked={ch.enabled}
                    onChange={(e) => toggle.mutate({ id: ch.id, enabled: e.target.checked })}
                  />
                </td>
                <td className="p-3 font-medium">{ch.custom_name || ch.channel.name}</td>
                <td className="p-3 text-slate-500">{ch.channel.group_title}</td>
                <td className="p-3 text-slate-400 font-mono text-xs">{ch.channel.tvg_id}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {channels.length === 0 && (
          <p className="p-6 text-center text-slate-400">No channels. Add a playlist first.</p>
        )}
      </div>
    </div>
  )
}
