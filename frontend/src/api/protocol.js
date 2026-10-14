export function unwrapEnvelope(body) {
  if (!body || typeof body.code !== 'number' || typeof body.message !== 'string' || !Object.hasOwn(body, 'data')) {
    throw new Error('服务器返回了无效的响应格式')
  }
  if (body.code !== 0) throw new Error(body.message || '请求失败')
  return body.data
}
export function apiUrl(base, path) {
  return `${base.replace(/\/$/, '')}/${path.replace(/^\//, '')}`
}

export function resolveFileURLs(value, base) {
  if (Array.isArray(value)) return value.map(item => resolveFileURLs(item, base))
  if (!value || typeof value !== 'object') return value
  return Object.fromEntries(Object.entries(value).map(([key, field]) => {
    if (['file_url', 'url'].includes(key) && typeof field === 'string') {
      if (field.startsWith('/api/')) return [key, apiUrl(base, field.slice(5))]
      if (field.startsWith('/static/') && /^https?:\/\//.test(base)) return [key, new URL(field, base).href]
    }
    return [key, resolveFileURLs(field, base)]
  }))
}

export async function readEventStream(response, { onChunk, onDone }) {
  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let completed = false
  const consume = (event) => {
    const payload = event.split(/\r?\n/)
      .filter(line => line.startsWith('data:'))
      .map(line => line.slice(5).trimStart()).join('\n')
    if (!payload) return
    const value = JSON.parse(payload)
    if (value.type === 'chunk') onChunk(value.content)
    if (value.type === 'error') throw new Error(value.message || '生成失败')
    if (value.type === 'done') {
      completed = true
      onDone(value.data)
    }
  }
  try {
    while (!completed) {
      const { done, value } = await reader.read()
      buffer += done ? decoder.decode() : decoder.decode(value, { stream: true })
      const events = buffer.split(/\r?\n\r?\n/)
      buffer = events.pop()
      for (const event of events) {
        consume(event)
        if (completed) break
      }
      if (done) {
        if (!completed && buffer.trim()) consume(buffer)
        if (!completed) throw new Error('生成流提前结束')
        break
      }
    }
  } finally {
    await reader.cancel().catch(() => {})
    reader.releaseLock()
  }
}
