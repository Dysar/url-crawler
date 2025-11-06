import { useEffect, useState } from 'react'
import { getApiUrl } from './apiUrl'
import { login } from './services/api'
import { UrlForm } from './components/UrlForm'
import { UrlTable } from './components/UrlTable'

export function App() {
  const [health, setHealth] = useState<string>('...')
  const [token, setToken] = useState<string | null>(localStorage.getItem('auth_token'))
  const [authError, setAuthError] = useState<string | null>(null)
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('password')
  const [reload, setReload] = useState(0)

  useEffect(() => {
    fetch(getApiUrl('/health'))
      .then(r => r.json())
      .then(d => setHealth(JSON.stringify(d)))
      .catch(() => setHealth('unreachable'))
  }, [])

  return (
    <div style={{ padding: 16, fontFamily: 'sans-serif' }}>
      <h1>URL Crawler</h1>
      <p>Backend health: {health}</p>
      {!token ? (
        <form onSubmit={async e => { e.preventDefault(); setAuthError(null); try { const t = await login(username, password); setToken(t) } catch { setAuthError('Login failed') } }} style={{ display: 'flex', gap: 8, margin: '12px 0' }}>
          <input value={username} onChange={e => setUsername(e.target.value)} placeholder="username" />
          <input value={password} type="password" onChange={e => setPassword(e.target.value)} placeholder="password" />
          <button type="submit">Login</button>
          {authError && <span style={{ color: 'red' }}>{authError}</span>}
        </form>
      ) : (
        <>
          <UrlForm onCreated={() => { setReload(x => x + 1) }} />
          <div style={{ height: 8 }} />
          <UrlTable reload={reload} />
        </>
      )}
    </div>
  )
}


