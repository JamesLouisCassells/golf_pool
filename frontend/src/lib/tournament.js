import { apiFetch, responseMessage } from './api'

export async function fetchActiveConfig(options = {}) {
  const { admin = false } = options
  const path = admin ? '/api/admin/config/active' : '/api/config/active'
  const response = await apiFetch(path)

  if (!response.ok) {
    const error = new Error(await responseMessage(response, 'Failed to load the active tournament config.'))
    error.status = response.status
    throw error
  }

  return await response.json()
}
