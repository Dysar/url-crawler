import { getApiUrl } from '../apiUrl'

function buildHeaders(init?: Record<string, string>): Headers {
  const h = new Headers(init)
  const token = localStorage.getItem('auth_token')
  if (token) h.set('Authorization', `Bearer ${token}`)
  return h
}

export async function login(username: string, password: string): Promise<string> {
  const res = await fetch(getApiUrl('/api/v1/auth/login'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  if (!res.ok) throw new Error('Login failed')
  const data = await res.json()
  const token = data.token as string
  localStorage.setItem('auth_token', token)
  return token
}

export type URLItem = { id: number; url: string }

export async function createUrl(url: string): Promise<URLItem> {
  const headers = buildHeaders({ 'Content-Type': 'application/json' })
  const res = await fetch(getApiUrl('/api/v1/urls'), {
    method: 'POST',
    headers,
    body: JSON.stringify({ url }),
  })
  if (!res.ok) throw new Error('Create URL failed')
  const data = await res.json()
  return data.data as URLItem
}

export async function listUrls(page = 1, limit = 20): Promise<{ data: URLItem[]; total: number; page: number; limit: number }> {
  const headers = buildHeaders()
  const res = await fetch(getApiUrl(`/api/v1/urls?page=${page}&limit=${limit}`), { headers })
  if (!res.ok) throw new Error('List URLs failed')
  return res.json()
}

export async function startJobs(urlIds: number[]): Promise<any> {
  const headers = buildHeaders({ 'Content-Type': 'application/json' })
  const res = await fetch(getApiUrl('/api/v1/jobs/start'), {
    method: 'POST',
    headers,
    body: JSON.stringify({ url_ids: urlIds }),
  })
  if (!res.ok) throw new Error('Start jobs failed')
  return res.json()
}

export async function jobStatus(jobId: number): Promise<{ data: { id: number; status: string } }> {
  const headers = buildHeaders()
  const res = await fetch(getApiUrl(`/api/v1/jobs/${jobId}/status`), { headers })
  if (!res.ok) throw new Error('Job status failed')
  return res.json()
}

export type Result = {
  id: number
  url_id: number
  html_version: string | null
  title: string | null
  headings_h1: number
  headings_h2: number
  headings_h3: number
  headings_h4: number
  headings_h5: number
  headings_h6: number
  internal_links_count: number
  external_links_count: number
  inaccessible_links_count: number
  has_login_form: boolean
}

export async function getResult(urlId: number): Promise<Result> {
  const headers = buildHeaders()
  const res = await fetch(getApiUrl(`/api/v1/results/${urlId}`), { headers })
  if (!res.ok) throw new Error('Get result failed')
  const data = await res.json()
  return data.data as Result
}


