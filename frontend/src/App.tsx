import { useEffect, useState } from 'react'
import { getApiUrl } from './apiUrl'
import { login } from './services/api'
import { UrlForm } from './components/UrlForm'
import { UrlTable } from './components/UrlTable'

export function App() {
  const [health, setHealth] = useState<'checking' | 'ok' | 'error'>('checking')
  const [token, setToken] = useState<string | null>(localStorage.getItem('auth_token'))
  const [authError, setAuthError] = useState<string | null>(null)
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('password')
  const [reload, setReload] = useState(0)

  useEffect(() => {
    fetch(getApiUrl('/health'))
      .then(r => r.json())
      .then(d => {
        if (d.data?.ok === true) {
          setHealth('ok')
        } else {
          setHealth('error')
        }
      })
      .catch(() => setHealth('error'))
  }, [])

  function handleLogout() {
    localStorage.removeItem('auth_token')
    setToken(null)
  }

  return (
    <div style={{ padding: 16, fontFamily: 'sans-serif' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h1 style={{ margin: 0 }}>URL Crawler</h1>
        {token && <button onClick={handleLogout}>Logout</button>}
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
        <span style={{ fontSize: '14px', color: '#666' }}>Backend:</span>
        {health === 'checking' && (
          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: '#999', display: 'inline-block' }}></span>
            <span style={{ fontSize: '14px', color: '#666' }}>Checking...</span>
          </span>
        )}
        {health === 'ok' && (
          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: '#4CAF50', display: 'inline-block' }}></span>
            <span style={{ fontSize: '14px', color: '#4CAF50', fontWeight: 500 }}>Connected</span>
          </span>
        )}
        {health === 'error' && (
          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: '#f44336', display: 'inline-block' }}></span>
            <span style={{ fontSize: '14px', color: '#f44336', fontWeight: 500 }}>Unreachable</span>
          </span>
        )}
      </div>
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


