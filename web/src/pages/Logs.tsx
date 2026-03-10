import { useEffect, useRef, useState } from 'react'

interface LogMessage {
  type: string
  msg: string
}

export default function Logs() {
  const [logs, setLogs] = useState<string[]>([])
  const [connected, setConnected] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const wsURL = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`
    const ws = new WebSocket(wsURL)

    ws.onopen = () => setConnected(true)
    ws.onclose = () => setConnected(false)
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data) as LogMessage
        setLogs((prev) => [...prev.slice(-499), msg.msg])
      } catch {
        setLogs((prev) => [...prev.slice(-499), e.data])
      }
    }

    return () => ws.close()
  }, [])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Logs</h1>
        <span className={`text-xs px-2 py-1 rounded ${connected ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
          {connected ? 'Connected' : 'Disconnected'}
        </span>
      </div>
      <div className="flex-1 bg-slate-900 text-green-400 font-mono text-xs p-4 rounded overflow-auto">
        {logs.map((line, i) => <div key={i}>{line}</div>)}
        {logs.length === 0 && <span className="text-slate-500">Waiting for logs...</span>}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
