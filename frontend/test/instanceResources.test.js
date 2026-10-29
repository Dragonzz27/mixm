import test from 'node:test'
import assert from 'node:assert/strict'
import { getInstanceResources, formatGpuMemory } from '../src/utils/instanceResources.js'

test('resource display reads API fields independently of runtime config', () => {
  const instance = {
    cpu_limit: 2.5,
    memory_limit: '3Gi',
    storage_limit: 20,
    gpu_limit: 2,
    gpu_memory_limit: 3072,
    config: JSON.stringify({ password: 'test-only', working_dir: '/workspace', image: 'mixm-codeserver:test' })
  }
  assert.deepEqual(getInstanceResources(instance), {
    cpuLimit: 2.5,
    memoryLimit: '3Gi',
    storageLimit: 20,
    gpuLimit: 2,
    gpuMemoryLimit: 3072
  })
})

test('CPU-only instances preserve zero GPU values and ignore conflicting metadata', () => {
  const instance = {
    cpu_limit: 0.5,
    memory_limit: '512Mi',
    storage_limit: 5,
    gpu_limit: 0,
    gpu_memory_limit: 0,
    config: JSON.stringify({ cpu_limit: 8, memory_limit: '32Gi', storage_limit: 40, gpu_limit: 4, gpu_memory_limit: 8192 })
  }
  assert.deepEqual(getInstanceResources(instance), {
    cpuLimit: 0.5,
    memoryLimit: '512Mi',
    storageLimit: 5,
    gpuLimit: 0,
    gpuMemoryLimit: 0
  })
  assert.equal(formatGpuMemory(0), '-')
})

test('missing resources stay unknown instead of displaying invented limits', () => {
  const resource = getInstanceResources({ config: '{}' })
  assert.equal(resource.cpuLimit, '-')
  assert.equal(resource.memoryLimit, '-')
  assert.equal(resource.storageLimit, '-')
  assert.equal(resource.gpuLimit, '-')
})

test('GPU memory formats Mi without rounding non-integral Gi', () => {
  assert.equal(formatGpuMemory(4096), '4Gi')
  assert.equal(formatGpuMemory(1536), '1536Mi')
  assert.equal(formatGpuMemory(null), '-')
})
