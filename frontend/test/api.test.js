import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import vm from 'node:vm'

// Evaluate the actual API declarations with a recording transport, without DOM or network.
function client() {
  const calls = []
  const request = Object.fromEntries(['get', 'post', 'patch', 'put', 'delete'].map(method => [method, (...args) => {
    calls.push(JSON.parse(JSON.stringify({ method, args })))
    return Promise.resolve({ recorded: true })
  }]))
  const source = fs.readFileSync(new URL('../src/api/index.js', import.meta.url), 'utf8')
    .replace(/^import .*$/gm, '').replace(/export const /g, 'const ')
  const api = vm.runInNewContext(`${source}\n;({authApi, instanceApi, assignmentApi, simulationApi, bindingApi, questionApi})`, { request, stream() {} })
  return { api, calls }
}

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
