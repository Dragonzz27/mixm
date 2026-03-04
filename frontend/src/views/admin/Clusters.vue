<template>
  <div class="clusters">
    <n-grid :cols="3" :x-gap="24" :y-gap="24" style="margin-bottom: 24px;">
      <n-gi>
        <n-card>
          <n-statistic label="总Pod数" :value="summary.total_pods" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card>
          <n-statistic label="运行中" :value="summary.running_pods">
            <template #prefix>
              <span style="color: #18a058;">●</span>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card>
          <n-statistic label="节点数" :value="summary.total_nodes">
            <template #prefix>
              <span style="color: #2080f0;">●</span>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
    </n-grid>

    <n-tabs v-model:value="activeTab" type="card">
      <n-tab-pane name="pods" tab="Pod列表">
        <n-data-table
          :columns="podColumns"
          :data="pods"
          :bordered="false"
        />
      </n-tab-pane>
      <n-tab-pane name="deployments" tab="Deployment列表">
        <n-data-table
          :columns="deploymentColumns"
          :data="deployments"
          :bordered="false"
        />
      </n-tab-pane>
      <n-tab-pane name="nodes" tab="节点列表">
        <n-data-table
          :columns="nodeColumns"
          :data="nodes"
          :bordered="false"
        />
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NGrid, NGi, NCard, NStatistic, NTabs, NTabPane, NDataTable, NProgress, NTag, NButton, useMessage, useDialog } from 'naive-ui'
import { clusterApi } from '@/api'

const message = useMessage()
const dialog = useDialog()

const activeTab = ref('pods')
const pods = ref([])
const deployments = ref([])
const nodes = ref([])
const summary = ref({
  total_pods: 0,
  running_pods: 0,
  total_nodes: 0
})

function getStatusType(status) {
  switch (status) {
    case 'running': return 'success'
    case 'succeeded': return 'info'
    case 'failed': return 'error'
    case 'pending': return 'warning'
    default: return 'default'
  }
}

const podColumns = [
  { title: 'Pod名称', key: 'name', width: 220 },
  { title: '命名空间', key: 'namespace', width: 100 },
  { title: '节点', key: 'node', width: 140 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => h(NTag, { type: getStatusType(row.status), size: 'small' }, () => row.status)
  },
  { title: '容器', key: 'containers', width: 200, render: (row) => row.containers?.join(', ') || '-' },
  { title: '总重启次数', key: 'restarts', width: 100 },
  { title: '运行时长', key: 'age', width: 100 },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) => h(NButton, { size: 'small', type: 'error', onClick: () => handleDeletePod(row) }, () => '强制停止')
  }
]

const deploymentColumns = [
  { title: 'Deployment名称', key: 'name', width: 200 },
  { title: '命名空间', key: 'namespace', width: 100 },
  { title: '副本数', key: 'replicas', width: 80 },
  { title: '可用副本', key: 'available_replicas', width: 90 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => h(NTag, { type: row.status === '运行中' ? 'success' : 'warning', size: 'small' }, () => row.status)
  },
  { title: '运行时长', key: 'age', width: 100 },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) => h(NButton, { size: 'small', type: 'error', onClick: () => handleDeleteDeployment(row) }, () => '删除')
  }
]

const nodeColumns = [
  { title: '节点名称', key: 'name', width: 180 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => h(NTag, { type: row.status === 'Ready' ? 'success' : 'error', size: 'small' }, () => row.status)
  },
  {
    title: 'CPU',
    key: 'cpu',
    width: 180,
    render: (row) => {
      const used = (row.cpu_usage / 1000).toFixed(1)
      const total = (row.cpu_capacity / 1000).toFixed(1)
      return h('div', [
        h(NProgress, { type: 'line', percentage: row.cpu_capacity ? (row.cpu_usage / row.cpu_capacity * 100).toFixed(0) : 0, showInfo: false, style: { marginBottom: '4px' } }),
        h('span', { style: 'font-size: 12px;' }, `${used} / ${total} 核`)
      ])
    }
  },
  {
    title: '内存',
    key: 'memory',
    width: 200,
    render: (row) => {
      const used = (row.memory_usage / 1024 / 1024 / 1024).toFixed(1)
      const total = (row.memory_capacity / 1024 / 1024 / 1024).toFixed(1)
      return h('div', [
        h(NProgress, { type: 'line', percentage: row.memory_capacity ? (row.memory_usage / row.memory_capacity * 100).toFixed(0) : 0, showInfo: false, style: { marginBottom: '4px' } }),
        h('span', { style: 'font-size: 12px;' }, `${used} / ${total} GB`)
      ])
    }
  },
  { title: 'Pod数量', key: 'pods_count', width: 90 }
]

async function loadData() {
  try {
    const [podsRes, nodesRes, deploymentsRes] = await Promise.all([
      clusterApi.getPods(),
      clusterApi.getNodes(),
      clusterApi.getDeployments()
    ])
    pods.value = podsRes.pods || []
    nodes.value = nodesRes.nodes || []
    deployments.value = deploymentsRes.deployments || []
    summary.value = {
      total_pods: podsRes.total || 0,
      running_pods: pods.value.filter(p => p.status === 'running').length,
      total_nodes: nodesRes.total || 0
    }
  } catch (err) {
    console.error(err)
  }
}

async function handleDeletePod(pod) {
  dialog.warning({
    title: '强制停止 Pod',
    content: `确定要强制停止 Pod "${pod.name}" 吗？`,
    positiveText: '确定停止',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await clusterApi.deletePod(pod.namespace, pod.name)
        if (res.success) {
          message.success(res.message)
          setTimeout(loadData, 2000)
        } else {
          message.error(res.message)
        }
      } catch (err) {
        message.error('操作失败')
      }
    }
  })
}

async function handleDeleteDeployment(deployment) {
  dialog.warning({
    title: '删除 Deployment',
    content: `确定要删除 Deployment "${deployment.name}" 吗？此操作将同时删除其关联的Pod。`,
    positiveText: '确定删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await clusterApi.deleteDeployment(deployment.namespace, deployment.name)
        if (res.success) {
          message.success(res.message)
          setTimeout(loadData, 2000)
        } else {
          message.error(res.message)
        }
      } catch (err) {
        message.error('操作失败')
      }
    }
  })
}

onMounted(loadData)
</script>
