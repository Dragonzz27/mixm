<template>
  <div class="dashboard">
    <section class="dashboard-intro">
      <div>
        <p class="eyebrow">MIXM / OVERVIEW</p>
        <h1 class="editorial-title">教学与实践，一目了然。</h1>
        <p class="intro-description">连接学习者、知识与实验，为每一次成长创造空间。</p>
      </div>
      <span class="dashboard-edition">教学与实验平台<span>WORKSPACE / 1.0</span></span>
    </section>

    <section class="overview-grid" aria-label="平台数据概览">
      <article v-for="stat in statCards" :key="stat.path" class="overview-card">
        <div class="overview-caption">
          <span>{{ stat.index }} / {{ stat.label }}</span
          ><span>{{ stat.title }}</span>
        </div>
        <span class="overview-number">{{ stat.value }}</span>
        <router-link :to="stat.path" class="overview-link"
          >{{ stat.link }} <span aria-hidden="true">↗</span></router-link
        >
      </article>
    </section>

    <section class="resource-section">
      <div class="section-header">
        <div>
          <p class="eyebrow">SYSTEM RESOURCES</p>
          <h2>系统资源</h2>
        </div>
        <span class="refresh-note"><span></span>每 30 秒更新</span>
      </div>
      <div class="resource-grid">
        <n-card v-for="resource in resourceCards" :key="resource.label" class="resource-card">
          <div class="resource-caption">
            <span>{{ resource.title }}</span
            ><span class="eyebrow">{{ resource.label }}</span>
          </div>
          <p class="resource-number">{{ resource.value.toFixed(2) }}<span>%</span></p>
          <n-progress
            type="line"
            :percentage="resource.value"
            :color="resource.color"
            :show-indicator="false"
            :height="4"
          />
          <p class="resource-detail">{{ resource.detail }}</p>
        </n-card>
      </div>
    </section>

    <section class="system-details">
      <div class="system-info">
        <p class="eyebrow">SYSTEM INFORMATION</p>
        <h2>运行概况</h2>
        <dl>
          <div>
            <dt>进程数量</dt>
            <dd>{{ system.processes }}</dd>
          </div>
          <div>
            <dt>最近采集</dt>
            <dd class="timestamp">{{ system.timestamp }}</dd>
          </div>
        </dl>
      </div>
      <div class="system-load">
        <p class="eyebrow">SYSTEM LOAD</p>
        <h2>系统负载</h2>
        <div class="load-values">
          <div>
            <span>过去 1 分钟</span><strong>{{ load.load1?.toFixed(2) ?? '—' }}</strong>
          </div>
          <div>
            <span>过去 5 分钟</span><strong>{{ load.load5?.toFixed(2) ?? '—' }}</strong>
          </div>
          <div>
            <span>过去 15 分钟</span><strong>{{ load.load15?.toFixed(2) ?? '—' }}</strong>
          </div>
        </div>
      </div>
    </section>
    <footer class="dashboard-footer">
      <span>让知识流动，让探索发生。</span
      ><router-link to="/app">前往学习空间 <span aria-hidden="true">↗</span></router-link>
    </footer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { NCard, NProgress } from 'naive-ui'
import { palette } from '@/styles/theme'
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

const statCards = computed(() => [
  {
    title: '平台用户',
    value: stats.value.userCount,
    index: '01',
    label: 'PEOPLE',
    path: '/admin/users',
    link: '管理用户'
  },
  {
    title: '知识分类',
    value: stats.value.categoryCount,
    index: '02',
    label: 'KNOWLEDGE',
    path: '/admin/categories',
    link: '整理分类'
  },
  {
    title: '教学内容',
    value: stats.value.contentCount,
    index: '03',
    label: 'CONTENTS',
    path: '/admin/contents',
    link: '查看内容'
  }
])
const resourceCards = computed(() => [
  {
    title: 'CPU 使用率',
    label: 'COMPUTE',
    value: system.value.cpu,
    color: getCpuColor(system.value.cpu),
    detail: '计算资源使用情况'
  },
  {
    title: '内存使用率',
    label: 'MEMORY',
    value: system.value.memoryPercent,
    color: getMemoryColor(system.value.memoryPercent),
    detail: `${formatBytes(system.value.memoryUsed)} / ${formatBytes(system.value.memoryTotal)}`
  },
  {
    title: '磁盘使用率',
    label: 'STORAGE',
    value: system.value.diskPercent,
    color: getDiskColor(system.value.diskPercent),
    detail: `${formatBytes(system.value.diskUsed)} / ${formatBytes(system.value.diskTotal)}`
  }
])

let timer = null

function getCpuColor(percent) {
  if (percent < 50) return palette.success
  if (percent < 80) return palette.warning
  return palette.error
}

function getMemoryColor(percent) {
  if (percent < 70) return palette.success
  if (percent < 90) return palette.warning
  return palette.error
}

function getDiskColor(percent) {
  if (percent < 70) return palette.success
  if (percent < 90) return palette.warning
  return palette.error
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
      authApi.getUsers({ offset: 0, limit: 1 }),
      categoryApi.getList({}),
      contentApi.getList({})
    ])
    stats.value.userCount = users.total || 0
    stats.value.categoryCount = categories.total || 0
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
  }, 30000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.dashboard-intro {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 8px 0 32px;
  border-bottom: 1px solid var(--mixm-border);
}
.dashboard-intro .eyebrow {
  margin-bottom: 24px;
  font-size: 9px;
}
.dashboard-intro h1 {
  font-size: clamp(30px, 3.2vw, 44px);
  margin-bottom: 16px;
}
.intro-description {
  color: var(--mixm-muted);
  font-size: 13px;
}
.dashboard-edition {
  display: flex;
  flex-direction: column;
  gap: 8px;
  text-align: right;
  font-family: var(--mixm-font-display);
  font-size: 15px;
  color: var(--mixm-muted);
  flex-shrink: 0;
}
.dashboard-edition span {
  font-family: var(--mixm-font-body);
  font-size: 8px;
  letter-spacing: 0.12em;
}
.overview-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 28px 0 38px;
  border: 1px solid var(--mixm-border);
  background: var(--mixm-surface);
  border-radius: 5px;
}
.overview-card {
  padding: 24px 28px;
  min-width: 0;
}
.overview-card + .overview-card {
  border-left: 1px solid var(--mixm-border);
}
.overview-caption {
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-size: 12px;
}
.overview-caption > span:first-child {
  color: var(--mixm-faint);
  font-size: 9px;
  letter-spacing: 0.09em;
}
.overview-number {
  display: block;
  font-family: var(--mixm-font-display);
  font-size: 56px;
  letter-spacing: -0.04em;
  line-height: 1.3;
  padding: 14px 0 18px;
}
.overview-link {
  display: flex;
  justify-content: space-between;
  border-top: 1px solid var(--mixm-border);
  padding-top: 15px;
  font-size: 11px;
  color: var(--mixm-muted);
}
.overview-link span {
  color: var(--mixm-accent);
  font-size: 17px;
  line-height: 1;
}
.overview-link:hover {
  color: var(--mixm-accent);
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 20px;
}
.section-header .eyebrow {
  font-size: 9px;
  margin-bottom: 5px;
}
.section-header h2,
.system-details h2 {
  font-family: var(--mixm-font-display);
  font-size: 24px;
  font-weight: 400;
}
.refresh-note {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 10px;
  color: var(--mixm-muted);
}
.refresh-note > span {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--mixm-accent);
}
.resource-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}
.resource-caption {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
}
.resource-caption .eyebrow {
  font-size: 8px;
  color: var(--mixm-faint);
}
.resource-number {
  font-family: var(--mixm-font-display);
  font-size: 38px;
  line-height: 1.2;
  margin: 24px 0 20px;
  letter-spacing: -0.03em;
}
.resource-number > span {
  font-size: 18px;
  margin-left: 4px;
  color: var(--mixm-muted);
}
.resource-detail {
  font-size: 10px;
  color: var(--mixm-muted);
  margin-top: 14px;
}
.system-details {
  display: grid;
  grid-template-columns: 1fr 1.2fr;
  gap: 40px;
  border-top: 1px solid var(--mixm-border);
  border-bottom: 1px solid var(--mixm-border);
  margin-top: 34px;
  padding: 28px 0;
}
.system-details .eyebrow {
  font-size: 9px;
  margin-bottom: 7px;
}
.system-details h2 {
  margin-bottom: 22px;
}
.system-info dl {
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-size: 11px;
}
.system-info dl > div {
  display: flex;
  justify-content: space-between;
  gap: 20px;
}
.system-info dt {
  color: var(--mixm-muted);
  flex-shrink: 0;
}
.system-info dd {
  margin: 0;
  text-align: right;
}
.timestamp {
  overflow-wrap: anywhere;
}
.system-load {
  border-left: 1px solid var(--mixm-border);
  padding-left: 40px;
}
.load-values {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.load-values > div {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.load-values span {
  font-size: 10px;
  color: var(--mixm-muted);
}
.load-values strong {
  font-family: var(--mixm-font-display);
  font-weight: 400;
  font-size: 32px;
}
.dashboard-footer {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-top: 24px;
  color: var(--mixm-muted);
  font-size: 10px;
}
.dashboard-footer a:hover {
  color: var(--mixm-accent);
}
.dashboard-footer a > span {
  margin-left: 8px;
}
@media (max-width: 1000px) {
  .overview-card {
    padding: 22px;
  }
  .resource-grid {
    gap: 12px;
  }
  .resource-card :deep(.n-card__content) {
    padding: 20px;
  }
}
@media (max-width: 680px) {
  .dashboard-edition {
    display: none;
  }
  .dashboard-intro {
    padding-top: 0;
  }
  .dashboard-intro h1 {
    font-size: 30px;
  }
  .overview-grid {
    margin-block: 24px 30px;
  }
  .overview-card {
    padding: 18px 12px;
  }
  .overview-caption {
    font-size: 11px;
  }
  .overview-caption > span:first-child {
    font-size: 8px;
    letter-spacing: 0;
  }
  .overview-number {
    font-size: 40px;
  }
  .overview-link {
    font-size: 10px;
  }
  .resource-grid {
    grid-template-columns: 1fr;
  }
  .resource-number {
    margin: 16px 0;
  }
  .system-details {
    grid-template-columns: 1fr;
    gap: 28px;
    margin-top: 28px;
  }
  .system-load {
    border-left: 0;
    border-top: 1px solid var(--mixm-border);
    padding: 24px 0 0;
  }
  .dashboard-footer {
    flex-wrap: wrap;
  }
}
</style>
