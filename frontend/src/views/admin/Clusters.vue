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
        <n-data-table :columns="podColumns" :data="pods" :bordered="false" />
      </n-tab-pane>
      <n-tab-pane name="deployments" tab="Deployment列表">
        <n-data-table :columns="deploymentColumns" :data="deployments" :bordered="false" />
      </n-tab-pane>

      <n-tab-pane name="pv" tab="PV列表">
        <n-data-table :columns="pvColumns" :data="pvs" :bordered="false" />
      </n-tab-pane>
      <n-tab-pane name="pvc" tab="PVC列表">
        <n-data-table :columns="pvcColumns" :data="pvcs" :bordered="false" />
      </n-tab-pane>

      <n-tab-pane name="nodes" tab="节点列表">
        <n-data-table :columns="nodeColumns" :data="nodes" :bordered="false" />
      </n-tab-pane>
    </n-tabs>

    <n-drawer v-model:show="showPodDrawer" :width="700" placement="right">
      <n-drawer-content :title="'Pod 详情 - ' + (selectedPod?.name || '')">
        <n-spin :show="loadingDescribe">
          <n-descriptions v-if="podDescribe" label-placement="left" bordered :column="1" style="margin-bottom: 16px;">
            <n-descriptions-item label="名称">{{ podDescribe.name }}</n-descriptions-item>
            <n-descriptions-item label="命名空间">{{ podDescribe.namespace }}</n-descriptions-item>
            <n-descriptions-item label="节点">{{ podDescribe.node }}</n-descriptions-item>
            <n-descriptions-item label="状态">{{ podDescribe.status }}</n-descriptions-item>
            <n-descriptions-item label="IP">{{ podDescribe.ip }}</n-descriptions-item>
            <n-descriptions-item label="标签">
              <n-space>
                <n-tag v-for="(v, k) in podDescribe.labels" :key="k" size="small">{{ k }}: {{ v }}</n-tag>
              </n-space>
            </n-descriptions-item>
          </n-descriptions>
          <n-divider>容器</n-divider>
          <n-card v-for="c in podDescribe?.containers" :key="c.name" size="small" style="margin-bottom: 8px;">
            <n-descriptions :column="1" label-placement="top">
              <n-descriptions-item label="名称">{{ c.name }}</n-descriptions-item>
              <n-descriptions-item label="镜像">{{ c.image }}</n-descriptions-item>
              <n-descriptions-item v-if="c.command" label="命令">{{ c.command }}</n-descriptions-item>
            </n-descriptions>
          </n-card>
          <n-divider>事件</n-divider>
          <n-data-table v-if="podDescribe?.events?.length" :columns="eventColumns" :data="podDescribe.events"
            :bordered="false" size="small" />
          <n-empty v-else description="无事件" />
          <n-divider>日志</n-divider>
          <n-space style="margin-bottom: 12px;">
            <n-select v-model:value="selectedContainer" :options="containerOptions" placeholder="选择容器"
              style="width: 200px;" />
            <n-button @click="loadPodLogs" size="small">刷新日志</n-button>
          </n-space>
          <n-spin :show="loadingLogs">
            <pre class="log-viewer">{{ podLogs }}</pre>
          </n-spin>
        </n-spin>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted, h, computed } from 'vue'
import { NGrid, NGi, NCard, NStatistic, NTabs, NTabPane, NDataTable, NProgress, NTag, NButton, useMessage, useDialog, NDrawer, NDrawerContent, NSpace, NSelect, NSpin, NDescriptions, NDescriptionsItem, NDivider, NEmpty } from 'naive-ui'
import { clusterApi } from '@/api'

const message = useMessage()
const dialog = useDialog()

const activeTab = ref('pods')
const pods = ref([])
const deployments = ref([])
const nodes = ref([])
const pvs = ref([])
const pvcs = ref([])
const summary = ref({
  total_pods: 0,
  running_pods: 0,
  total_nodes: 0
})

const showPodDrawer = ref(false)
const selectedPod = ref(null)
const selectedContainer = ref(null)
const podLogs = ref('')
const podDescribe = ref(null)
const loadingLogs = ref(false)
const loadingDescribe = ref(false)

function getStatusType(status) {
  switch (status) {
    case 'running': return 'success'
    case 'succeeded': return 'info'
    case 'failed': return 'error'
    case 'pending': return 'warning'
    default: return 'default'
  }
}

const containerOptions = computed(() => {
  if (!selectedPod.value?.containers) return []
  return selectedPod.value.containers.map(c => ({ label: c, value: c }))
})

const eventColumns = [
  { title: '类型', key: 'type', width: 80, render: (row) => h(NTag, { type: row.type === 'Warning' ? 'warning' : 'info', size: 'small' }, () => row.type) },
  { title: '原因', key: 'reason', width: 120 },
  { title: '消息', key: 'message' }
]

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
    width: 120,
    render: (row) => h(NSpace, { size: 'small' }, () => [
      h(NButton, { size: 'small', onClick: () => showPodDetail(row) }, () => '详情'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDeletePod(row) }, () => '停止')
    ])
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

const pvColumns = [
  { title: 'PV名称', key: 'name', width: 180 },
  { title: '容量', key: 'capacity', width: 100, render: (row) => row.capacity?.storage || '-' },
  { title: '访问模式', key: 'access_modes', width: 120, render: (row) => row.access_modes?.join(', ') || '-' },
  { title: '回收策略', key: 'reclaim_policy', width: 100 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => h(NTag, { type: row.status === 'Bound' ? 'success' : 'warning', size: 'small' }, () => row.status)
  },
  { title: '存储类', key: 'storage_class', width: 100 },
  { title: '声明', key: 'claim', width: 150 },
  { title: '运行时长', key: 'age', width: 100 },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) => h(NButton, { size: 'small', type: 'error', onClick: () => handleDeletePV(row) }, () => '删除')
  }
]

const pvcColumns = [
  { title: 'PVC名称', key: 'name', width: 180 },
  { title: '命名空间', key: 'namespace', width: 100 },
  { title: '容量', key: 'capacity', width: 100, render: (row) => row.capacity?.storage || '-' },
  { title: '访问模式', key: 'access_modes', width: 120, render: (row) => row.access_modes?.join(', ') || '-' },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => h(NTag, { type: row.status === 'Bound' ? 'success' : 'warning', size: 'small' }, () => row.status)
  },
  { title: '存储类', key: 'storage_class', width: 100 },
  { title: 'PV', key: 'volume', width: 150 },
  { title: '运行时长', key: 'age', width: 100 },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) => h(NButton, { size: 'small', type: 'error', onClick: () => handleDeletePVC(row) }, () => '删除')
  }
]

async function loadData() {
  try {
    const [podsRes, nodesRes, deploymentsRes, pvsRes, pvcsRes] = await Promise.all([
      clusterApi.getPods(),
      clusterApi.getNodes(),
      clusterApi.getDeployments(),
      clusterApi.getPV(),
      clusterApi.getPVC()
    ])
    pods.value = podsRes.items || []
    nodes.value = nodesRes.items || []
    deployments.value = deploymentsRes.items || []
    pvs.value = pvsRes.items || []
    pvcs.value = pvcsRes.items || []
    summary.value = {
      total_pods: podsRes.total || 0,
      running_pods: pods.value.filter(p => p.status === 'running').length,
      total_nodes: nodesRes.total || 0
    }
  } catch (err) {
    console.error(err)
  }
}

async function showPodDetail(pod) {
  selectedPod.value = pod
  selectedContainer.value = pod.containers?.[0] || null
  showPodDrawer.value = true
  loadingDescribe.value = true
  loadingLogs.value = true
  try {
    const [describeRes, logsRes] = await Promise.all([
      clusterApi.describePod(pod.namespace, pod.name),
      clusterApi.getPodLogs(pod.namespace, pod.name, selectedContainer.value, 200)
    ])
    if (describeRes.success) {
      podDescribe.value = describeRes.pod
    } else {
      message.error(describeRes.message)
    }
    if (logsRes.success) {
      podLogs.value = logsRes.logs || '(无日志)'
    } else {
      podLogs.value = `Error: ${logsRes.message}`
    }
  } catch (err) {
    message.error('获取详情失败')
  } finally {
    loadingDescribe.value = false
    loadingLogs.value = false
  }
}

async function loadPodLogs() {
  if (!selectedPod.value) return
  loadingLogs.value = true
  try {
    const res = await clusterApi.getPodLogs(
      selectedPod.value.namespace,
      selectedPod.value.name,
      selectedContainer.value,
      200
    )
    if (res.success) {
      podLogs.value = res.logs || '(无日志)'
    } else {
      podLogs.value = `Error: ${res.message}`
    }
  } catch (err) {
    podLogs.value = `Error: ${err.message}`
  } finally {
    loadingLogs.value = false
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

async function handleDeletePV(pv) {
  dialog.warning({
    title: '删除 PV',
    content: `确定要删除 PV "${pv.name}" 吗？`,
    positiveText: '确定删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await clusterApi.deletePV(pv.name)
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

async function handleDeletePVC(pvc) {
  dialog.warning({
    title: '删除 PVC',
    content: `确定要删除 PVC "${pvc.name}" 吗？`,
    positiveText: '确定删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await clusterApi.deletePVC(pvc.namespace, pvc.name)
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

<style scoped>
.log-viewer {
  background: #1a1a1a;
  color: #e0e0e0;
  padding: 12px;
  border-radius: 4px;
  max-height: 500px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.5;
  margin: 0;
}
</style>
