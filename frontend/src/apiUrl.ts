export function getApiUrl(endpoint: string): string {
  const base = import.meta.env.VITE_API_URL || 'http://localhost:8080'
  return `${base}${endpoint}`
}


