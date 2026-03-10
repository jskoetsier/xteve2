import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'

export default function Dashboard() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['status'],
    queryFn: api.status,
    refetchInterval: 5000,
  })

  if (isLoading) return <p>Loading...</p>
  if (error) return <p className="text-red-500">Failed to load status</p>

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>
      <div className="grid grid-cols-3 gap-4">
        <StatCard label="Version" value={data?.version ?? '—'} />
        <StatCard label="Active Streams" value={String(data?.active_streams ?? 0)} />
        <StatCard label="Tuners Available" value={`${data?.active_streams ?? 0} / ${data?.tuner_count ?? 0}`} />
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="border rounded-lg p-4 bg-white shadow-sm">
      <p className="text-sm text-slate-500">{label}</p>
      <p className="text-2xl font-semibold mt-1">{value}</p>
    </div>
  )
}
