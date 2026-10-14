<template>
  <n-modal v-model:show="showModal" preset="dialog" title="创建队列" style="width: 600px;">
    <n-spin :show="loading">
      <n-space vertical :size="16">
        <n-space justify="space-between" align="center">
          <n-tag type="info">我的待处理: {{ queueStatus.pending }}</n-tag>
          <n-tag type="warning">我的处理中: {{ queueStatus.processing }}</n-tag>
          <n-tag type="success">我的已完成: {{ queueStatus.completed }}</n-tag>
          <n-tag type="error">我的失败: {{ queueStatus.failed }}</n-tag>
        </n-space>
        
        <n-alert v-if="tasks.length === 0" type="info">
          当前账号没有创建任务
        </n-alert>
        
        <n-list v-else bordered>
          <n-list-item v-for="task in tasks" :key="task.task_id">
            <n-space vertical :size="8">
              <n-space justify="space-between" align="center">
                <n-text strong>{{ task.instance_name }}</n-text>
                <n-tag :type="getStatusType(task.status)">
                  {{ getStatusText(task.status) }}
                </n-tag>
              </n-space>
              <n-space :size="12">
                <n-text depth="3">用户: {{ task.username || '-' }}</n-text>
                <n-text depth="3">创建时间: {{ formatTime(task.created_at) }}</n-text>
              </n-space>
              <n-text v-if="task.error" depth="3" type="error">
                错误: {{ task.error }}
              </n-text>
              <n-text v-if="task.instance_id && task.status === 'completed'" depth="3" type="success">
                实例ID: {{ task.instance_id }}
              </n-text>
            </n-space>
          </n-list-item>
        </n-list>
      </n-space>
    </n-spin>
    <template #action>
      <n-button @click="handleRefresh">刷新</n-button>
      <n-button @click="showModal = false">关闭</n-button>
    </template>
  </n-modal>
</template>

<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import { NModal, NSpace, NTag, NAlert, NList, NListItem, NText, NSpin, NButton } from 'naive-ui'
import { instanceApi } from '@/api'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:show'])

const loading = ref(false)
const showModal = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})

const queueStatus = ref({
  pending: 0,
  processing: 0,
  completed: 0,
  failed: 0,
  total: 0
})

const tasks = ref([])
let pollTimer = null

async function loadData() {
  loading.value = true
  try {
    const [statusRes, tasksRes] = await Promise.all([
      instanceApi.getQueueStatus(),
      instanceApi.getQueueTasks()
    ])
    queueStatus.value = {
      pending: statusRes.pending || 0,
      processing: statusRes.processing || 0,
      completed: statusRes.completed || 0,
      failed: statusRes.failed || 0,
      total: statusRes.total || 0
    }
    tasks.value = tasksRes.items || []
  } catch (err) {
    console.error('Failed to load queue data:', err)
  } finally {
    loading.value = false
  }
}

function startPolling() {
  stopPolling()
  loadData()
  pollTimer = setInterval(() => {
    loadData()
  }, 30000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function getStatusType(status) {
  const typeMap = {
    pending: 'info',
    processing: 'warning',
    completed: 'success',
    failed: 'error'
  }
  return typeMap[status] || 'default'
}

function getStatusText(status) {
  const textMap = {
    pending: '等待中',
    processing: '创建中',
    completed: '已完成',
    failed: '失败'
  }
  return textMap[status] || status
}

function formatTime(timeStr) {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN')
}

function handleRefresh() {
  loadData()
}

watch(showModal, (val) => {
  if (val) {
    startPolling()
  } else {
    stopPolling()
  }
})

function handleQueueUpdated() {
  if (showModal.value) {
    loadData()
  }
}

window.addEventListener('queue-updated', handleQueueUpdated)

onUnmounted(() => {
  stopPolling()
  window.removeEventListener('queue-updated', handleQueueUpdated)
})
</script>
