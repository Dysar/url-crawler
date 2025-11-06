import { useState } from 'react'
import { createUrl } from '../services/api'

export function UrlForm({ onCreated }: { onCreated: () => void }) {
  const [url, setUrl] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      await createUrl(url)
      setUrl('')
      onCreated()
    } catch (e) {
      setError('Failed to add URL')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={onSubmit} style={{ display: 'flex', gap: 8 }}>
      <input
        type="url"
        required
        value={url}
        onChange={e => setUrl(e.target.value)}
        placeholder="https://example.com"
        style={{ flex: 1 }}
      />
      <button type="submit" disabled={loading}>Add URL</button>
      {error && <span style={{ color: 'red' }}>{error}</span>}
    </form>
  )
}


