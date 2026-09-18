let cachedPort: number | null = null

export async function serverBase(): Promise<string> {
  if (cachedPort == null) {
    cachedPort = window.api ? await window.api.getServerPort() : 18080
  }
  return `http://127.0.0.1:${cachedPort}`
}

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const base = await serverBase()
  const full = path.startsWith('/api') ? path : `/api${path}`
  const res = await fetch(`${base}${full}`, init)
  let json: ApiResponse<T> | null = null
  try {
    json = (await res.json()) as ApiResponse<T>
  } catch {
    /* ignore */
  }
  if (!res.ok || !json || json.code !== 0) {
    throw new Error(json?.message || `请求失败 HTTP ${res.status}`)
  }
  return json.data
}

export async function apiGet<T>(path: string): Promise<T> {
  return request<T>(path)
}

export async function apiPost<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export async function apiPut<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export async function apiDelete<T>(path: string): Promise<T> {
  return request<T>(path, { method: 'DELETE' })
}

export async function checkHealth(): Promise<boolean> {
  try {
    const res = await fetch(`${await serverBase()}/healthz`)
    return res.ok
  } catch {
    return false
  }
}
