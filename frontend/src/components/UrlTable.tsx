import { useEffect, useState } from 'react'
import { listUrls, startJobs, stopJobs, URLItem, jobStatus, getResult, Result } from '../services/api'

type RowWithStatus = URLItem & { jobId?: number; status?: string; result?: Result }

type SortField = 'id' | 'url' | 'created_at' | 'updated_at'
type SortOrder = 'asc' | 'desc'

function StatusBadge({ status }: { status?: string }) {
  if (!status) return <span style={{ color: '#666' }}>-</span>
  const colors: Record<string, { bg: string; color: string }> = {
    queued: { bg: '#e0e0e0', color: '#333' },
    running: { bg: '#2196F3', color: '#fff' },
    completed: { bg: '#4CAF50', color: '#fff' },
    failed: { bg: '#f44336', color: '#fff' },
    stopped: { bg: '#FF9800', color: '#fff' },
  }
  const style = colors[status] || { bg: '#999', color: '#fff' }
  return (
    <span style={{
      padding: '2px 8px',
      borderRadius: '4px',
      fontSize: '12px',
      backgroundColor: style.bg,
      color: style.color,
    }}>
      {status}
    </span>
  )
}

export function UrlTable({ reload }: { reload: number }) {
  const [rows, setRows] = useState<RowWithStatus[]>([])
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [jobMap, setJobMap] = useState<Map<number, number>>(new Map())
  const [page, setPage] = useState(1)
  const [limit] = useState(20)
  const [total, setTotal] = useState(0)
  const [sortBy, setSortBy] = useState<SortField>('created_at')
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc')

  async function load() {
    try {
      const data = await listUrls(page, limit, sortBy, sortOrder)
      const baseRows = data.data.map(r => ({ ...r }))
      setRows(baseRows)
      setTotal(data.total)

      // After initial load, fetch latest results for each URL so refresh shows persisted data
      Promise.all(
        baseRows.map(async (r) => {
          try {
            const res = await getResult(r.id)
            setRows(prev => prev.map(p => p.id === r.id ? { ...p, result: res, status: p.status || 'completed' } : p))
          } catch {
            // result may not exist yet; ignore
          }
        })
      )
      .catch(() => {})
    } catch (err) {
      setMessage('Failed to load URLs')
    }
  }

  useEffect(() => { load() }, [reload, page, limit, sortBy, sortOrder])

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

  function toggleAll() {
    if (selected.size === rows.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(rows.map(r => r.id)))
    }
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
      setSelected(new Set())
    } catch {
      setMessage('Failed to start jobs')
    } finally {
      setLoading(false)
    }
  }

  async function stopSelected() {
    setLoading(true)
    setMessage(null)
    try {
      await stopJobs([...selected])
      setMessage('Jobs stopped')
      setSelected(new Set())
      load() // Reload to update status
    } catch {
      setMessage('Failed to stop jobs')
    } finally {
      setLoading(false)
    }
  }

  function handleSort(field: SortField) {
    if (sortBy === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
    } else {
      setSortBy(field)
      setSortOrder('desc')
    }
  }

  const totalPages = Math.ceil(total / limit)

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, margin: '8px 0', flexWrap: 'wrap' }}>
        <button onClick={startSelected} disabled={loading || selected.size === 0}>Start</button>
        <button onClick={stopSelected} disabled={loading || selected.size === 0}>Stop</button>
        {message && <span style={{ color: message.includes('Failed') ? 'red' : 'green' }}>{message}</span>}
      </div>
      <table width="100%" cellPadding={8} style={{ borderCollapse: 'collapse', border: '1px solid #ccc', marginTop: 8 }}>
        <thead>
          <tr style={{ backgroundColor: '#f5f5f5' }}>
            <th><input type="checkbox" checked={selected.size === rows.length && rows.length > 0} onChange={toggleAll} /></th>
            <th style={{ cursor: 'pointer' }} onClick={() => handleSort('id')}>
              ID {sortBy === 'id' && (sortOrder === 'asc' ? '↑' : '↓')}
            </th>
            <th style={{ cursor: 'pointer' }} onClick={() => handleSort('url')}>
              URL {sortBy === 'url' && (sortOrder === 'asc' ? '↑' : '↓')}
            </th>
            <th>Status</th>
            <th>Title</th>
            <th>H1-H6</th>
            <th>Links</th>
            <th>Login</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(r => (
            <tr key={r.id} style={{ borderBottom: '1px solid #eee' }}>
              <td><input type="checkbox" checked={selected.has(r.id)} onChange={() => toggle(r.id)} /></td>
              <td>{r.id}</td>
              <td style={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.url}</td>
              <td><StatusBadge status={r.status} /></td>
              <td style={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.result?.title || '-'}</td>
              <td>{r.result ? `${r.result.headings_h1}/${r.result.headings_h2}/${r.result.headings_h3}` : '-'}</td>
              <td>{r.result ? `I:${r.result.internal_links_count} E:${r.result.external_links_count}` : '-'}</td>
              <td>{r.result?.has_login_form ? 'Yes' : '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {totalPages > 1 && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 16, justifyContent: 'center' }}>
          <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1}>Previous</button>
          <span>Page {page} of {totalPages} (Total: {total})</span>
          <button onClick={() => setPage(p => Math.min(totalPages, p + 1))} disabled={page === totalPages}>Next</button>
        </div>
      )}
    </div>
  )
}
