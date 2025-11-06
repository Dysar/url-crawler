import { useEffect, useState } from 'react'
import { listUrls, startJobs, URLItem, jobStatus, getResult, Result } from '../services/api'

type RowWithStatus = URLItem & { jobId?: number; status?: string; result?: Result }

export function UrlTable({ reload }: { reload: number }) {
  const [rows, setRows] = useState<RowWithStatus[]>([])
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [jobMap, setJobMap] = useState<Map<number, number>>(new Map()) // url_id -> job_id

  async function load() {
    const data = await listUrls(1, 50)
    setRows(data.data.map(r => ({ ...r })))
  }

  useEffect(() => { load() }, [reload])

  useEffect(() => {
    const interval = setInterval(() => {
      jobMap.forEach((jobId, urlId) => {
        jobStatus(jobId).then(res => {
          setRows(prev => prev.map(r => r.id === urlId ? { ...r, status: res.data.status } : r))
          if (res.data.status === 'completed') {
            getResult(urlId).then(result => {
              setRows(prev => prev.map(r => r.id === urlId ? { ...r, result } : r))
            }).catch(() => {})
          }
        }).catch(() => {})
      })
    }, 2000)
    return () => clearInterval(interval)
  }, [jobMap])

  function toggle(id: number) {
    const s = new Set(selected)
    if (s.has(id)) s.delete(id); else s.add(id)
    setSelected(s)
  }

  async function startSelected() {
    setLoading(true)
    setMessage(null)
    try {
      const res = await startJobs([...selected])
      const newMap = new Map(jobMap)
      res.data.forEach((item: any) => {
        newMap.set(item.url_id, item.job_id)
        setRows(prev => prev.map(r => r.id === item.url_id ? { ...r, jobId: item.job_id, status: 'queued' } : r))
      })
      setJobMap(newMap)
      setMessage('Jobs started')
    } catch {
      setMessage('Failed to start jobs')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, margin: '8px 0' }}>
        <button onClick={startSelected} disabled={loading || selected.size === 0}>Start</button>
        {message && <span>{message}</span>}
      </div>
      <table width="100%" cellPadding={6} style={{ borderCollapse: 'collapse', border: '1px solid #ccc' }}>
        <thead>
          <tr>
            <th></th>
            <th>ID</th>
            <th>URL</th>
            <th>Status</th>
            <th>Title</th>
            <th>H1-H6</th>
            <th>Links</th>
            <th>Login</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(r => (
            <tr key={r.id}>
              <td><input type="checkbox" checked={selected.has(r.id)} onChange={() => toggle(r.id)} /></td>
              <td>{r.id}</td>
              <td>{r.url}</td>
              <td>{r.status || '-'}</td>
              <td>{r.result?.title || '-'}</td>
              <td>{r.result ? `${r.result.headings_h1}/${r.result.headings_h2}/${r.result.headings_h3}` : '-'}</td>
              <td>{r.result ? `I:${r.result.internal_links_count} E:${r.result.external_links_count}` : '-'}</td>
              <td>{r.result?.has_login_form ? 'Yes' : '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}


