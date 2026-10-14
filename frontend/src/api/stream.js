import { API_BASE_URL } from './config'
import { apiUrl, readEventStream } from './protocol'

export async function stream(path, data, callbacks) {
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(apiUrl(API_BASE_URL, path), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data)
    })
    if (!response.ok) {
      const failure = await response.json().catch(() => ({}))
      throw new Error(failure.message || `HTTP ${response.status}`)
    }
    await readEventStream(response, callbacks)
  } catch (error) {
    callbacks.onError(error.message)
  }
}
