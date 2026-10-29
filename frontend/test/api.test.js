import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import vm from 'node:vm'

// Evaluate the actual API declarations with a recording transport, without DOM or network.
function client(respond) {
  const calls = []
  const request = Object.fromEntries(['get', 'post', 'patch', 'put', 'delete'].map(method => [method, (...args) => {
    calls.push(JSON.parse(JSON.stringify({ method, args })))
    return Promise.resolve(respond ? respond(method, args) : { recorded: true })
  }]))
  const source = fs.readFileSync(new URL('../src/api/index.js', import.meta.url), 'utf8')
    .replace(/^import .*$/gm, '').replace(/export const /g, 'const ')
  const api = vm.runInNewContext(`${source}\n;({authApi, instanceApi, imageApi, assignmentApi, simulationApi, bindingApi, questionApi, clusterApi})`, { request, stream() {} })
  return { api, calls }
}

test('image catalog synchronization uses the protected POST route without repository names', async () => {
  const { api, calls } = client()
  await api.imageApi.sync()
  assert.deepEqual(calls, [{ method: 'post', args: ['/images/sync'] }])
  const routes = JSON.parse(fs.readFileSync(new URL('../../backend/internal/router/routes.json', import.meta.url), 'utf8'))
  assert.ok(routes.includes('POST /api/images/sync'))
})

test('article course editing paginates bindings for the requested article', async () => {
  const { api, calls } = client()
  await api.bindingApi.getByContent('article', { offset: 1000, limit: 1000 })
  assert.deepEqual(calls, [{ method: 'get', args: ['/bindings', { params: { offset: 1000, limit: 1000, content_id: 'article' } }] }])
})

test('image tag selectors read every API page instead of stopping at the default page', async () => {
  const all = Array.from({ length: 1003 }, (_, index) => ({ name: `tag-${index}` }))
  const { api, calls } = client((method, args) => {
    const { offset, limit } = args[1].params
    return { items: all.slice(offset, offset + limit), total: all.length, offset, limit }
  })
  const page = await api.imageApi.getAllTags('repo')
  assert.equal(page.items.length, 1003)
  assert.equal(page.items.at(-1).name, 'tag-1002')
  assert.deepEqual(calls.map(call => call.args[1].params.offset), [0, 1000])
})

test('incomplete tag pagination fails instead of silently losing labels or looping', async () => {
  const { api } = client(() => ({ items: [], total: 1, offset: 0, limit: 1000 }))
  await assert.rejects(api.imageApi.getAllTags('repo'), /镜像标签分页响应不完整/)
})

test('cluster resource requests match registered backend routes', async () => {
  const { api, calls } = client()
  const routes = new Set(JSON.parse(fs.readFileSync(new URL('../../backend/internal/router/routes.json', import.meta.url), 'utf8')))
  await api.clusterApi.getPods()
  await api.clusterApi.getNodes()
  await api.clusterApi.getDeployments()
  await api.clusterApi.getPV()
  await api.clusterApi.getPVC()
  for (const { method, args } of calls) {
    const route = `${method.toUpperCase()} /api${args[0]}`
    assert.ok(routes.has(route), `cluster API route is missing: ${route}`)
  }
})

test('instance state, queue status and tasks use the new REST routes', async () => {
  const { api, calls } = client()
  assert.deepEqual(await api.instanceApi.getQueueStatus(), { recorded: true })
  await api.instanceApi.start('instance')
  await api.instanceApi.stop('instance')
  await api.instanceApi.getQueueTasks()
  assert.deepEqual(calls, [
    { method: 'get', args: ['/instance-tasks/summary'] },
    { method: 'patch', args: ['/instances/instance', { status: 'running' }] },
    { method: 'patch', args: ['/instances/instance', { status: 'stopped' }] },
    { method: 'get', args: ['/instance-tasks'] }
  ])
})
test('submission, simulation and template IDs are moved from payloads to paths', async () => {
  const { api, calls } = client()
  await api.assignmentApi.submit({ assignment_id: 'assignment', content: 'answer' })
  await api.simulationApi.submit({ attempt_id: 'attempt', answers: [] })
  await api.instanceApi.startBatchByTemplate({ template_id: 'template', user_ids: ['user'] })
  assert.deepEqual(calls, [
    { method: 'post', args: ['/assignments/assignment/submissions', { content: 'answer' }] },
    { method: 'post', args: ['/simulation-attempts/attempt/submissions', { answers: [] }] },
    { method: 'post', args: ['/templates/template/instances', { user_ids: ['user'] }] }
  ])
})
test('user creation, password, binding counts and question reviews use canonical methods', async () => {
  const { api, calls } = client()
  await api.authApi.register({ username: 'student' })
  await api.authApi.changePassword('user', { new_password: 'test' })
  await api.bindingApi.getCounts({ category_ids: ['a', 'b'] })
  await api.questionApi.review('question', { review_status: 'approved' })
  assert.deepEqual(calls, [
    { method: 'post', args: ['/users', { username: 'student' }] },
    { method: 'put', args: ['/users/user/password', { new_password: 'test' }] },
    { method: 'get', args: ['/binding-counts', { params: { category_ids: 'a,b' } }] },
    { method: 'put', args: ['/questions/question/review', { review_status: 'approved' }] }
  ])
})
