// 实例资源来自接口顶层字段；config 仅保存访问凭据等运行信息。
export function getInstanceResources(instance) {
  return {
    cpuLimit: instance.cpu_limit ?? '-',
    memoryLimit: instance.memory_limit ?? '-',
    storageLimit: instance.storage_limit ?? '-',
    gpuLimit: instance.gpu_limit ?? '-',
    gpuMemoryLimit: instance.gpu_memory_limit ?? 0
  }
}

// 后端单卡显存以 Mi 为单位，展示时按整 Gi 换算。
export function formatGpuMemory(value) {
  const memoryMi = Number(value)
  if (!Number.isFinite(memoryMi) || memoryMi <= 0) return '-'
  if (memoryMi % 1024 === 0) return `${memoryMi / 1024}Gi`
  return `${memoryMi}Mi`
}
