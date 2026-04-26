export async function requestJSON<T>(input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const response = await fetch(input, init)
  const text = await response.text()
  const payload = text ? JSON.parse(text) : null

  if (!response.ok) {
    const message = payload?.error || payload?.message || `HTTP ${response.status}`
    throw new Error(message)
  }

  return payload as T
}
