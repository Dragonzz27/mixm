<template>
  <div class="instances">
    <n-card title="实例管理">
      <template #header-extra>
        <n-space>
          <n-select
            v-model:value="filterTemplateId"
            :options="templateOptions"
            clearable
            placeholder="按模板筛选"
            style="width: 180px;"
            @update:value="filterInstances"
          />
          <n-button type="primary" @click="showStartModal = true">启动新实例</n-button>
          <n-button @click="loadData">刷新</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="filteredInstances"
        :pagination="pagination"
        :bordered="false"
      />
    </n-card>
    <StartInstanceModal
      v-model:show="showStartModal"
      :template="null"
      @success="loadData"
    />
    <InstanceConfigModal
      v-model:show="showConfigModal"
      :config="currentConfig"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { NCard, NButton, NDataTable, NSpace, NSelect, NTag, useMessage } from 'naive-ui'
import { instanceApi, templateApi } from '@/api'
import StartInstanceModal from '@/components/StartInstanceModal.vue'
import InstanceConfigModal from '@/components/InstanceConfigModal.vue'

const message = useMessage()
const instances = ref([])
const templates = ref([])
const filteredInstances = ref([])
const showStartModal = ref(false)
const showConfigModal = ref(false)
const currentConfig = ref(null)
const filterTemplateId = ref(null)
const pagination = ref({ pageSize: 10 })

function getStatusType(status) {
  switch (status) {
    case 'running': return 'success'
    case 'stopped': return 'default'
    case 'failed': return 'error'
    case 'pending': return 'warning'
    default: return 'default'
  }
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatMB(mb) {
  if (!mb || mb === 0) return '0 MB'
  if (mb >= 1024) {
    return parseFloat((mb / 1024).toFixed(1)) + ' GB'
  }
  return mb + ' MB'
}

const templateOptions = computed(() => {
  return templates.value.map(t => ({ label: t.name, value: t.id }))
})

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '实例名称', key: 'name', width: 160 },
  { title: '模板', key: 'template_name', width: 120 },
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
    key: 'cpu_usage',
    width: 100,
    render: (row) => h('span', `${row.cpu_usage}%`)
  },
  {
    title: '内存',
    key: 'memory_usage',
    width: 140,
    render: (row) => h('span', `${formatMB(row.memory_usage)} / ${formatMB(row.memory_limit)}`)
  },
  { title: '绑定用户', key: 'username', width: 100 },
  { title: '启动时间', key: 'started_at', width: 180 },
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
    const [instancesRes, templatesRes] = await Promise.all([
      instanceApi.getList({}),
      templateApi.getList({})
    ])
    instances.value = instancesRes.instances || []
    templates.value = templatesRes.templates || []
    filterInstances()
  } catch (err) {
    console.error(err)
  }
}

function filterInstances() {
  if (filterTemplateId.value) {
    filteredInstances.value = instances.value.filter(i => i.template_id === filterTemplateId.value)
  } else {
    filteredInstances.value = instances.value
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
  const config = row.config ? JSON.parse(row.config) : {}
  const hostIp = window.location.hostname
  currentConfig.value = {
    id: row.id,
    name: row.name,
    image: 'codercom/code-server:latest',
    accessUrl: `http://${hostIp}:18800/${row.id}/`,
    password: config.password || '无',
    cpuLimit: '1',
    memoryLimit: formatMB(row.memory_limit),
    namespace: row.namespace,
    portMapping: '18800 -> 80'
  }
  showConfigModal.value = true
}

onMounted(loadData)
</script>
