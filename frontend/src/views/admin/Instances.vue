<template>
  <div class="instances">
    <n-card title="实例管理">
      <template #header-extra>
        <n-space>
          <n-button type="primary" @click="showStartModal = true">启动新实例</n-button>
          <n-button @click="loadData">刷新</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="instances"
        :pagination="pagination"
        :bordered="false"
      />
    </n-card>
    <StartInstanceModal
      v-model:show="showStartModal"
      :enable-user-select="true"
      @success="loadData"
    />
    <InstanceConfigModal
      v-model:show="showConfigModal"
      :config="currentConfig"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NCard, NButton, NDataTable, NSpace, NTag, useMessage } from 'naive-ui'
import { instanceApi } from '@/api'
import StartInstanceModal from '@/components/StartInstanceModal.vue'
import InstanceConfigModal from '@/components/InstanceConfigModal.vue'

const message = useMessage()
const instances = ref([])
const showStartModal = ref(false)
const showConfigModal = ref(false)
const currentConfig = ref(null)
const pagination = ref({ pageSize: 10 })

function parseConfig(config) {
  if (!config) return {}
  if (typeof config === 'string') {
    try {
      return JSON.parse(config)
    } catch (err) {
      return {}
    }
  }
  return config
}

function resolveGpuLimit(row) {
  if (row.gpu_limit !== undefined && row.gpu_limit !== null) {
    return Number(row.gpu_limit) || 0
  }
  const config = parseConfig(row.config)
  return Number(config.gpu_limit) || 0
}

function resolveGpuMemoryLimit(row) {
  if (row.gpu_memory_limit !== undefined && row.gpu_memory_limit !== null) {
    return Number(row.gpu_memory_limit) || 0
  }
  const config = parseConfig(row.config)
  return Number(config.gpu_memory_limit) || 0
}

function formatGpuMemory(mb) {
  const memoryMb = Number(mb) || 0
  if (memoryMb <= 0) return '-'
  if (memoryMb % 1024 === 0) {
    return `${memoryMb / 1024}Gi`
  }
  return `${memoryMb}Mi`
}

function getStatusType(status) {
  switch (status) {
    case 'running': return 'success'
    case 'stopped': return 'default'
    case 'failed': return 'error'
    case 'pending': return 'warning'
    default: return 'default'
  }
}

const columns = [
  { title: 'ID', key: 'id', width: 160 },
  { title: '实例名称', key: 'name', width: 160 },
  { title: '镜像', key: 'image', width: 200, render: (row) => h('span', row.image || '-') },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => h(NTag, { type: getStatusType(row.status), size: 'small' }, () => row.status)
  },
  { title: '命名空间', key: 'namespace', width: 100 },
  { title: '副本数', key: 'replicas', width: 80 },
  {
    title: 'CPU',
    key: 'cpu_limit',
    width: 100,
    render: (row) => h('span', `${row.cpu_limit || 1}核`)
  },
  {
    title: '内存',
    key: 'memory_limit',
    width: 140,
    render: (row) => h('span', row.memory_limit || '-')
  },
  {
    title: 'GPU数量',
    key: 'gpu_limit',
    width: 100,
    render: (row) => h('span', `${resolveGpuLimit(row)}张`)
  },
  {
    title: '单卡显存',
    key: 'gpu_memory_limit',
    width: 120,
    render: (row) => h('span', formatGpuMemory(resolveGpuMemoryLimit(row)))
  },
  { title: '绑定用户', key: 'username', width: 100 },
  { title: '启动时间', key: 'started_at', width: 150 },
  {
    title: '操作',
    key: 'actions',
    width: 240,
    render: (row) => h(NSpace, {}, () => [
      row.status === 'running' 
        ? h(NButton, { size: 'small', type: 'warning', onClick: () => handleStop(row.id) }, () => '停止')
        : row.status === 'stopped'
        ? h(NButton, { size: 'small', type: 'success', onClick: () => handleStart(row.id) }, () => '启动')
        : null,
      h(NButton, { size: 'small', onClick: () => handleViewConfig(row) }, () => '配置'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
]

async function loadData() {
  try {
    const instancesRes = await instanceApi.getList({})
    instances.value = instancesRes.instances || []
    for (const instance of instances.value) {
      if (instance.started_at) {
        instance.name = instance.name.split('-').slice(0, -1).join('-')
        instance.image = instance.image.split('/').pop()
        instance.started_at = new Date(instance.started_at).toLocaleString()
      }
    }
  } catch (err) {
    console.error(err)
  }
}

async function handleStart(id) {
  try {
    await instanceApi.start(id)
    message.success('实例已启动')
    loadData()
  } catch (err) {
    console.error(err)
  }
}

async function handleStop(id) {
  try {
    await instanceApi.stop(id)
    message.success('实例已停止')
    loadData()
  } catch (err) {
    console.error(err)
  }
}

async function handleDelete(id) {
  try {
    await instanceApi.delete(id)
    message.success('实例已删除')
    loadData()
  } catch (err) {
    console.error(err)
  }
}

function handleViewConfig(row) {
  const config = parseConfig(row.config)
  const sshHost = row.hostIp || ''
  const sshPort = row.ssh_port || ''
  const sshUsername = row.ssh_username || ''
  const sshConnection = sshUsername && sshPort ? `ssh ${sshUsername}@${sshHost} -p ${sshPort}` : ''
  const endpoint = row.endpoint || `/${row.id}`
  const accessPort = row.port || 80
  const codeServerPassword = config.password || ''
  
  currentConfig.value = {
    id: row.id,
    name: row.name,
    image: row.image || '-',
    cpuLimit: row.cpu_limit || 1,
    memoryLimit: row.memory_limit || '1Gi',
    storageLimit: row.storage_limit || 1,
    gpuLimit: resolveGpuLimit(row),
    gpuMemoryLimit: resolveGpuMemoryLimit(row),
    namespace: row.namespace,
    sshHost: sshHost,
    sshPort: sshPort,
    sshUsername: sshUsername,
    sshPassword: row.ssh_password || '',
    sshConnection: sshConnection,
    endpoint: endpoint,
    accessPort: accessPort,
    codeServerPassword: codeServerPassword
  }
  showConfigModal.value = true
}

onMounted(loadData)
</script>
