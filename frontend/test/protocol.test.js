import test from 'node:test'
import assert from 'node:assert/strict'
import { unwrapEnvelope, apiUrl, resolveFileURLs, readEventStream } from '../src/api/protocol.js'

test('unwraps tokens and paginated data, preserves zero and null', () => {
  const tokens = { access_token: 'access', refresh_token: 'refresh' }
  assert.deepEqual(unwrapEnvelope({ code: 0, message: '成功', data: tokens }), tokens)
  assert.equal(unwrapEnvelope({ code: 0, message: '成功', data: null }), null)
  const page = { items: [], total: 0, offset: 100, limit: 20 }
  assert.deepEqual(unwrapEnvelope({ code: 0, message: '成功', data: page }), page)
  assert.throws(() => unwrapEnvelope({ code: 403, message: '没有权限', data: null }), /没有权限/)
  assert.throws(() => unwrapEnvelope({ detail: 'legacy' }), /无效/)
})
test('uses configured API base with consistent separators', () => {
  assert.equal(apiUrl('https://example.test/api/', '/contents/files'), 'https://example.test/api/contents/files')
  assert.deepEqual(resolveFileURLs({ items: [{ file_url: '/api/assignments/file/legacy.txt' }], url: '/static/export/grades.csv' }, 'https://example.test/custom/api'), {
    items: [{ file_url: 'https://example.test/custom/api/assignments/file/legacy.txt' }], url: 'https://example.test/static/export/grades.csv'
  })
})
function response(text, width = 1) {
  const bytes = new TextEncoder().encode(text)
  return new Response(new ReadableStream({ start(controller) {
    for (let offset = 0; offset < bytes.length; offset += width) controller.enqueue(bytes.slice(offset, offset + width))
    controller.close()
  }}))
}
test('SSE handles split UTF-8 chunks and completion without trailing newline', async () => {
  const chunks = []; let result
  await readEventStream(response('data: {"type":"chunk","content":"中文"}\r\n\r\ndata: {"type":"done","data":{"title":"题目"}}'), {
    onChunk: value => chunks.push(value), onDone: value => { result = value }
  })
  assert.deepEqual(chunks, ['中文']); assert.deepEqual(result, { title: '题目' })
})
test('SSE reports upstream errors and incomplete responses', async () => {
  const callbacks = { onChunk() {}, onDone() {} }
  await assert.rejects(readEventStream(response('data: {"type":"error","message":"服务不可用"}\n\n'), callbacks), /服务不可用/)
  await assert.rejects(readEventStream(response('data: {"type":"chunk","content":"partial"}\n\n'), callbacks), /提前结束/)
})
