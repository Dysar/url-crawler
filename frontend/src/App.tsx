import { useEffect, useState } from 'react'
import { getApiUrl } from './apiUrl'

export function App() {
  const [health, setHealth] = useState<string>('...')

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
    </div>
  )
}


