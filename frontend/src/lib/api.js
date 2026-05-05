import { getAuthToken } from './auth'

export async function apiFetch(input, init = {}) {
  const headers = new Headers(init.headers ?? {})
  const token = await getAuthToken()

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  return fetch(input, {
    ...init,
    headers,
  })
}

export async function responseMessage(response, fallback) {
  const text = (await response.text()).trim()
  return text || fallback
}
