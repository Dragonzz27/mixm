<template>
  <div class="dashboard">
    <n-card title="欢迎使用 MutiM">
      <template #header-extra>
        <n-tag type="success">系统运行中</n-tag>
      </template>
      <n-descriptions :columns="4" label-placement="left">
        <n-descriptions-item label="版本">1.0.0</n-descriptions-item>
        <n-descriptions-item label="用户总数">{{ stats.userCount }}</n-descriptions-item>
        <n-descriptions-item label="分类总数">{{ stats.categoryCount }}</n-descriptions-item>
        <n-descriptions-item label="内容总数">{{ stats.contentCount }}</n-descriptions-item>
      </n-descriptions>
    </n-card>

    <n-divider />

    <n-grid :cols="3" :x-gap="24" :y-gap="24">
      <n-gi>
        <n-card>
          <n-statistic label="用户总数" :value="stats.userCount" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card>
          <n-statistic label="分类总数" :value="stats.categoryCount" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card>
          <n-statistic label="内容总数" :value="stats.contentCount" />
        </n-card>
      </n-gi>
    </n-grid>

    <n-divider />

    <n-grid :cols="3" :x-gap="24" :y-gap="24">
      <n-gi>
        <n-card title="CPU使用率">
          <n-progress
            type="line"
            :percentage="system.cpu"
            :color="getCpuColor(system.cpu)"
          />
          <div class="stat-value">{{ system.cpu }}%</div>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="内存使用率">
          <n-progress
            type="line"
            :percentage="system.memoryPercent"
            :color="getMemoryColor(system.memoryPercent)"
          />
          <div class="stat-value">{{ system.memoryPercent }}% ({{ formatBytes(system.memoryUsed) }}/{{ formatBytes(system.memoryTotal) }})</div>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="磁盘使用率">
          <n-progress
            type="line"
            :percentage="system.diskPercent"
            :color="getDiskColor(system.diskPercent)"
          />
          <div class="stat-value">{{ system.diskPercent }}% ({{ formatBytes(system.diskUsed) }}/{{ formatBytes(system.diskTotal) }})</div>
        </n-card>
      </n-gi>
    </n-grid>

    <n-divider />

    <n-grid :cols="2" :x-gap="24">
      <n-gi>
        <n-card title="系统信息">
          <n-descriptions :column="1" label-placement="left">
            <n-descriptions-item label="进程数">{{ system.processes }}</n-descriptions-item>
            <n-descriptions-item label="更新时间">{{ system.timestamp }}</n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="系统负载">
          <n-descriptions :column="1" label-placement="left">
            <n-descriptions-item label="1分钟"> {{ load.load1?.toFixed(2) || '-' }}</n-descriptions-item>
            <n-descriptions-item label="5分钟"> {{ load.load5?.toFixed(2) || '-' }}</n-descriptions-item>
            <n-descriptions-item label="15分钟"> {{ load.load15?.toFixed(2) || '-' }}</n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { NGrid, NGi, NCard, NStatistic, NDivider, NProgress, NDescriptions, NDescriptionsItem, NTag } from 'naive-ui'
import { authApi, categoryApi, contentApi, systemApi } from '@/api'

const stats = ref({
  userCount: 0,
  contentCount: 0,
  categoryCount: 0
})

const system = ref({
  cpu: 0,
  memoryTotal: 0,
  memoryUsed: 0,
  memoryPercent: 0,
  diskTotal: 0,
  diskUsed: 0,
  diskPercent: 0,
  processes: 0,
  timestamp: '-'
})

const load = ref({
  load1: 0,
  load5: 0,
  load15: 0
})

let timer = null

function getCpuColor(percent) {
  if (percent < 50) return '#18a058'
  if (percent < 80) return '#f0a020'
  return '#d03050'
}

function getMemoryColor(percent) {
  if (percent < 70) return '#18a058'
  if (percent < 90) return '#f0a020'
  return '#d03050'
}

function getDiskColor(percent) {
  if (percent < 70) return '#18a058'
  if (percent < 90) return '#f0a020'
  return '#d03050'
}

function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

async function loadStats() {
  try {
    const [users, categories, contents] = await Promise.all([
      authApi.getUsers({ skip: 0, limit: 1 }),
      categoryApi.getList({}),
      contentApi.getList({})
    ])
    stats.value.userCount = users.total || 0
    stats.value.categoryCount = categories.length || 0
    stats.value.contentCount = contents.total || 0
  } catch (err) {
    console.error(err)
  }
}

async function loadSystemStats() {
  try {
    const res = await systemApi.getStats()
    system.value = {
      cpu: res.cpu || 0,
      memoryTotal: res.memory?.total || 0,
      memoryUsed: res.memory?.used || 0,
      memoryPercent: res.memory?.percent || 0,
      diskTotal: res.disk?.total || 0,
      diskUsed: res.disk?.used || 0,
      diskPercent: res.disk?.percent || 0,
      processes: res.processes || 0,
      timestamp: res.timestamp || '-'
    }
  } catch (err) {
    console.error(err)
  }
}

async function loadSystemLoad() {
  try {
    const res = await systemApi.getLoad()
    load.value = {
      load1: res.load_1 || 0,
      load5: res.load_5 || 0,
      load15: res.load_15 || 0
    }
  } catch (err) {
    console.error(err)
  }
}

onMounted(async () => {
  await loadStats()
  await loadSystemStats()
  await loadSystemLoad()
  timer = setInterval(async () => {
    await loadSystemStats()
    await loadSystemLoad()
  }, 5000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.dashboard {
  padding: 20px 0;
}

.stat-value {
  text-align: center;
  margin-top: 8px;
  font-size: 18px;
  font-weight: bold;
  color: #333;
}
</style>
