export type LoginResponse = {
  token: string
  username: string
}

export type JobRun = {
  id: number
  status: string
  startedAt: string
  finishedAt: string | null
  errorMessage: string | null
  extractedCount: number
}

export type JobsResponse = {
  cronSchedule: string
  runs: JobRun[]
}

export type ExpenseRow = {
  id: number
  shop: string
  item: string
  expense: string
  sourceId: number | null
  jobRunId: number | null
  createdAt: string
}

export type SourceRow = {
  id: number
  name: string
  websiteUrl: string
  isEnabled: boolean
  createdAt: string
  updatedAt: string
}

export type ForwardingState = {
  targetApiUrl: string
  lastStatus: string
  lastResponseCode: number | null
  lastError: string
  lastAt: string | null
}

const TOKEN_KEY = 'agent-scraper-token'
const USER_KEY = 'agent-scraper-user'

export const API_BASE = (() => {
  const raw = import.meta.env.VITE_API_BASE_URL as string | undefined
  if (raw === undefined || raw === '') return ''
  return raw.replace(/\/$/, '')
})()

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setSession(token: string, username: string): void {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, username)
}

export function clearSession(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

async function apiFetch<T>(path: string, options: RequestInit = {}, auth = false): Promise<T> {
  const headers = new Headers(options.headers)
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json')
  }
  if (auth) {
    const token = getToken()
    if (!token) throw new Error('Not authenticated')
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_BASE}${path}`, { ...options, headers })
  if (response.status === 204) {
    return undefined as T
  }
  if (!response.ok) {
    let message = `Request failed (${response.status})`
    try {
      const data = (await response.json()) as { error?: string }
      if (data.error) message = data.error
    } catch {
      /* ignore */
    }
    throw new Error(message)
  }
  return response.json() as Promise<T>
}

export function login(body: { username: string; password: string }): Promise<LoginResponse> {
  return apiFetch<LoginResponse>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function fetchJobs(): Promise<JobsResponse> {
  return apiFetch<JobsResponse>('/api/v1/jobs', {}, true)
}

export function runJobNow(): Promise<{ ok: boolean; runs?: JobRun[] }> {
  return apiFetch<{ ok: boolean; runs?: JobRun[] }>('/api/v1/jobs/run', { method: 'POST' }, true)
}

export function fetchExpenses(): Promise<{ expenses: ExpenseRow[] }> {
  return apiFetch<{ expenses: ExpenseRow[] }>('/api/v1/expenses', {}, true)
}

export function fetchSources(): Promise<{ sources: SourceRow[] }> {
  return apiFetch<{ sources: SourceRow[] }>('/api/v1/sources', {}, true)
}

export function createSource(body: {
  name: string
  websiteUrl: string
  isEnabled: boolean
}): Promise<SourceRow> {
  return apiFetch<SourceRow>(
    '/api/v1/sources',
    { method: 'POST', body: JSON.stringify(body) },
    true,
  )
}

export function updateSource(
  id: number,
  body: { name: string; websiteUrl: string; isEnabled: boolean },
): Promise<SourceRow> {
  return apiFetch<SourceRow>(
    `/api/v1/sources/${id}`,
    { method: 'PATCH', body: JSON.stringify(body) },
    true,
  )
}

export function deleteSource(id: number): Promise<void> {
  return apiFetch<void>(`/api/v1/sources/${id}`, { method: 'DELETE' }, true)
}

export function fetchForwarding(): Promise<ForwardingState> {
  return apiFetch<ForwardingState>('/api/v1/forwarding', {}, true)
}

export function saveForwarding(targetApiUrl: string): Promise<ForwardingState> {
  return apiFetch<ForwardingState>(
    '/api/v1/forwarding',
    { method: 'PUT', body: JSON.stringify({ targetApiUrl }) },
    true,
  )
}
