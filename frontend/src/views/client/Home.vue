<template>
  <div class="client-home">
    <section class="home-hero" aria-labelledby="home-title">
      <div class="hero-copy">
        <p class="eyebrow">
          YOUR LEARNING SPACE <span class="hero-date">{{ today }}</span>
        </p>
        <p class="hero-greeting">你好，{{ userStore.userInfo?.username || '探索者' }}。</p>
        <h1 id="home-title" class="editorial-title">把想法，变成下一次探索。</h1>
        <p class="hero-description">读一点新知，动手做一次实验。让好奇心，带你再向前一步。</p>
        <div class="hero-actions">
          <n-button type="primary" @click="router.push('/app/contents')"
            >探索学习内容 <span class="link-arrow" aria-hidden="true">↗</span></n-button
          >
          <router-link to="/app/my-instances" class="text-link"
            >我的实验实例 <span aria-hidden="true">→</span></router-link
          >
        </div>
      </div>
      <ExplorationArt class="home-art" />
    </section>

    <section class="stats-bar" aria-label="实例概览">
      <div class="stat-item">
        <span class="stat-value">{{ stats.total.toString().padStart(2, '0') }}</span>
        <div>
          <span class="stat-label">全部实例</span><span class="stat-caption">每一个想法的起点</span>
        </div>
      </div>
      <div class="stat-item">
        <span class="stat-value">{{ stats.running.toString().padStart(2, '0') }}</span>
        <div>
          <span class="stat-label"><i class="status-dot running"></i>正在运行</span
          ><span class="stat-caption">继续正在进行的探索</span>
        </div>
      </div>
      <div class="stat-item">
        <span class="stat-value">{{ stats.stopped.toString().padStart(2, '0') }}</span>
        <div>
          <span class="stat-label"><i class="status-dot"></i>已停止</span
          ><span class="stat-caption">准备好时，随时再出发</span>
        </div>
      </div>
    </section>

    <div class="home-columns">
      <section class="instances-section">
        <div class="section-header">
          <div>
            <p class="eyebrow">EXPERIMENTS</p>
            <h2>你的实践空间</h2>
          </div>
          <router-link to="/app/my-instances" class="text-link"
            >查看全部 <span aria-hidden="true">↗</span></router-link
          >
        </div>
        <n-spin :show="loadingInstances">
          <div v-if="instances.length > 0" class="instance-grid">
            <article v-for="instance in instances" :key="instance.id" class="instance-card">
              <div class="instance-header">
                <n-tag
                  :type="instance.status === 'running' ? 'success' : 'warning'"
                  size="small"
                  :bordered="false"
                  >{{ instance.status === 'running' ? '运行中' : '已停止' }}</n-tag
                ><span class="instance-time">{{ formatTime(instance.created_at) }}</span>
              </div>
              <h3>
                <router-link :to="`/app/instances/${instance.id}`"
                  >{{ instance.name }} <span aria-hidden="true">↗</span></router-link
                >
              </h3>
              <p class="instance-desc">实验模板 / {{ instance.template_id || '未指定' }}</p>
              <div class="instance-actions">
                <n-button
                  v-if="instance.status !== 'running'"
                  size="small"
                  type="primary"
                  @click="startInstance(instance.id)"
                  >启动实例</n-button
                >
                <n-button v-else size="small" @click="stopInstance(instance.id)">停止实例</n-button>
                <n-button size="small" text type="error" @click="deleteInstance(instance.id)"
                  >删除</n-button
                >
              </div>
            </article>
          </div>
          <div v-else class="empty-experiments">
            <span class="empty-number" aria-hidden="true">↗</span>
            <h3>第一次探索，从这里开始。</h3>
            <p>从学习内容中选择实验，开启你的实践空间。</p>
            <n-button @click="router.push('/app/contents')">浏览学习内容</n-button>
          </div>
        </n-spin>
      </section>

      <section class="recent-section">
        <div class="section-header">
          <div>
            <p class="eyebrow">RECENT ACTIVITY</p>
            <h2>最近的足迹</h2>
          </div>
        </div>
        <n-spin :show="loadingActivities">
          <div v-if="recentActivities.length > 0" class="activity-list">
            <button
              v-for="activity in recentActivities"
              :key="activity.id"
              type="button"
              class="activity-item"
              @click="goToActivity(activity)"
            >
              <span class="activity-icon" :style="{ color: getActivityColor(activity.type) }"
                ><n-icon :size="17"><component :is="getActivityIcon(activity.type)" /></n-icon
              ></span>
              <span class="activity-content"
                ><span class="activity-title">{{ activity.title }}</span
                ><span class="activity-desc">{{ formatRelativeTime(activity.time) }}</span></span
              >
              <span class="activity-arrow" aria-hidden="true">↗</span>
            </button>
          </div>
          <div v-else class="empty-activity">
            <span class="eyebrow">A FRESH START</span>
            <p>每一次尝试，<br />都值得被记录。</p>
            <span>开始实验后，你的足迹会出现在这里。</span>
          </div>
        </n-spin>
      </section>
    </div>

    <section class="explore-section" aria-label="学习入口">
      <router-link
        v-for="entry in entryLinks"
        :key="entry.path"
        :to="entry.path"
        class="explore-entry"
      >
        <span class="entry-index">{{ entry.number }} / {{ entry.label }}</span>
        <h3>{{ entry.title }} <span aria-hidden="true">↗</span></h3>
        <p>{{ entry.description }}</p>
      </router-link>
    </section>
    <footer class="home-footer">
      <span>为每一次探索，留出空间。</span><span class="eyebrow">MIXM / LEARN BY DOING</span>
    </footer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import { instanceApi } from '@/api'
import ExplorationArt from '@/components/ExplorationArt.vue'
import { palette } from '@/styles/theme'
import {
  PlayCircleOutline,
  StopCircleOutline,
  TrashOutline,
  EyeOutline,
  CreateOutline
} from '@vicons/ionicons5'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const loadingInstances = ref(false)
const loadingActivities = ref(false)
const instances = ref([])
const recentActivities = ref([])

const today = new Intl.DateTimeFormat('zh-CN', {
  month: 'long',
  day: 'numeric',
  weekday: 'long'
}).format(new Date())
const entryLinks = [
  {
    number: '01',
    title: '学习内容',
    description: '找到值得深入的下一个问题。',
    path: '/app/contents',
    label: 'EXPLORE'
  },
  {
    number: '02',
    title: '我的作业',
    description: '让每一次思考，都有所回应。',
    path: '/app/assignments',
    label: 'CREATE'
  },
  {
    number: '03',
    title: '模拟测试',
    description: '回顾所学，发现新的可能。',
    path: '/app/simulation-test',
    label: 'REFLECT'
  }
]

const stats = computed(() => ({
  total: instances.value.length,
  running: instances.value.filter((i) => i.status === 'running').length,
  stopped: instances.value.filter((i) => i.status !== 'running').length
}))

const activityTypeConfig = {
  instance_start: { color: palette.success, icon: PlayCircleOutline },
  instance_stop: { color: palette.warning, icon: StopCircleOutline },
  instance_delete: { color: palette.error, icon: TrashOutline },
  instance_create: { color: palette.info, icon: CreateOutline },
  content_view: { color: palette.info, icon: EyeOutline }
}

function getActivityColor(type) {
  return activityTypeConfig[type]?.color || palette.muted
}

function getActivityIcon(type) {
  return activityTypeConfig[type]?.icon || EyeOutline
}

function formatRelativeTime(timestamp) {
  if (!timestamp) return ''
  const now = new Date()
  const time = new Date(timestamp)
  const diff = now - time
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  return time.toLocaleDateString()
}

function goToActivity(activity) {
  if (activity.targetRoute) {
    router.push(activity.targetRoute)
  }
}

async function fetchMyInstances() {
  try {
    loadingInstances.value = true
    const res = await instanceApi.getMyInstances()
    instances.value = res.items || []
    buildRecentActivities()
  } catch (err) {
    console.error('Failed to fetch instances:', err)
    message.error('实例加载失败')
  } finally {
    loadingInstances.value = false
  }
}

function buildRecentActivities() {
  const activities = []

  instances.value.forEach((instance) => {
    activities.push({
      id: `create-${instance.id}`,
      type: 'instance_create',
      title: `创建实例: ${instance.name}`,
      description: `模板: ${instance.template_id || 'N/A'}`,
      time: instance.created_at,
      targetRoute: `/app/instances/${instance.id}`
    })

    if (instance.status === 'running' && instance.started_at) {
      activities.push({
        id: `start-${instance.id}`,
        type: 'instance_start',
        title: `启动实例: ${instance.name}`,
        description: '实例正在运行中',
        time: instance.started_at,
        targetRoute: `/app/instances/${instance.id}`
      })
    }
  })

  activities.sort((a, b) => new Date(b.time) - new Date(a.time))
  recentActivities.value = activities.slice(0, 10)
}

function formatTime(timestamp) {
  if (!timestamp) return 'N/A'
  return new Date(timestamp).toLocaleDateString()
}

function goToInstance(id) {
  router.push(`/app/instances/${id}`)
}

async function startInstance(id) {
  try {
    await instanceApi.start(id)
    message.success('实例已启动')
    fetchMyInstances()
  } catch (err) {
    message.error('实例启动失败')
  }
}

async function stopInstance(id) {
  try {
    await instanceApi.stop(id)
    message.success('实例已停止')
    fetchMyInstances()
  } catch (err) {
    message.error('实例停止失败')
  }
}

async function deleteInstance(id) {
  try {
    await instanceApi.delete(id)
    message.success('实例已删除')
    fetchMyInstances()
  } catch (err) {
    message.error('实例删除失败')
  }
}

onMounted(() => {
  loadingActivities.value = true
  fetchMyInstances().finally(() => {
    loadingActivities.value = false
  })
})
</script>

<style scoped>
.client-home {
  padding-bottom: 8px;
}
.home-hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 10px 0 36px;
}
.hero-copy {
  flex: 1;
  min-width: 0;
}
.hero-copy > .eyebrow {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 9px;
  margin-bottom: 28px;
}
.hero-date {
  font-weight: 400;
  letter-spacing: 0.03em;
  color: var(--mixm-faint);
}
.hero-greeting {
  font-size: 13px;
  color: var(--mixm-muted);
  margin-bottom: 10px;
}
.hero-copy h1 {
  font-size: clamp(30px, 3.1vw, 46px);
  line-height: 1.35;
  margin-bottom: 18px;
}
.hero-description {
  max-width: 470px;
  font-size: 13px;
  line-height: 1.9;
  color: var(--mixm-muted);
}
.hero-actions {
  display: flex;
  align-items: center;
  gap: 24px;
  flex-wrap: wrap;
  margin-top: 26px;
}
.link-arrow {
  margin-left: 16px;
  font-size: 18px;
}
.text-link {
  font-size: 12px;
  color: var(--mixm-muted);
  white-space: nowrap;
}
.text-link span {
  margin-left: 8px;
}
.text-link:hover {
  color: var(--mixm-accent);
}
.home-art {
  width: 29%;
  max-width: 310px;
  min-width: 200px;
  flex-shrink: 0;
}
.stats-bar {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-top: 1px solid var(--mixm-border);
  border-bottom: 1px solid var(--mixm-border);
  margin-bottom: 40px;
}
.stat-item {
  display: flex;
  gap: 22px;
  align-items: center;
  padding: 24px;
}
.stat-item:first-child {
  padding-left: 0;
}
.stat-item + .stat-item {
  border-left: 1px solid var(--mixm-border);
}
.stat-value {
  font-family: var(--mixm-font-display);
  font-size: 42px;
  line-height: 1;
  letter-spacing: -0.04em;
}
.stat-label {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
  margin-bottom: 4px;
}
.stat-caption {
  font-size: 10px;
  color: var(--mixm-muted);
}
.status-dot {
  display: inline-block;
  width: 5px;
  height: 5px;
  background: var(--mixm-warning);
  border-radius: 50%;
}
.status-dot.running {
  background: var(--mixm-success);
}
.home-columns {
  display: grid;
  grid-template-columns: minmax(0, 1.8fr) minmax(220px, 1fr);
  gap: 36px;
  margin-bottom: 40px;
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}
.section-header .eyebrow {
  font-size: 9px;
  margin-bottom: 6px;
  color: var(--mixm-faint);
}
.section-header h2 {
  font-size: 24px;
  font-weight: 400;
  letter-spacing: -0.025em;
}
.instance-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}
.instance-card {
  background: var(--mixm-surface);
  border: 1px solid var(--mixm-border);
  border-radius: 5px;
  padding: 20px;
  min-width: 0;
  transition: border-color 0.2s;
}
.instance-card:hover {
  border-color: var(--mixm-accent);
}
.instance-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 20px;
}
.instance-time {
  font-size: 9px;
  color: var(--mixm-muted);
}
.instance-card h3 {
  font-family: var(--mixm-font-display);
  font-size: 19px;
  font-weight: 400;
  overflow-wrap: anywhere;
}
.instance-card h3 a {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}
.instance-card h3 a:hover {
  color: var(--mixm-accent);
}
.instance-card h3 span {
  font-family: var(--mixm-font-body);
  color: var(--mixm-faint);
}
.instance-desc {
  font-size: 11px;
  color: var(--mixm-muted);
  margin: 8px 0 22px;
}
.instance-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border-top: 1px solid var(--mixm-border);
  padding-top: 14px;
}
.activity-list {
  border-top: 1px solid var(--mixm-border);
}
.activity-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  border: 0;
  border-bottom: 1px solid var(--mixm-border);
  padding: 17px 0;
  background: transparent;
  text-align: left;
  cursor: pointer;
  color: var(--mixm-ink);
}
.activity-item:hover .activity-title {
  color: var(--mixm-accent);
}
.activity-icon {
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  background: var(--mixm-surface-muted);
  border-radius: 50%;
  flex-shrink: 0;
}
.activity-content {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}
.activity-title {
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.activity-desc {
  font-size: 10px;
  color: var(--mixm-muted);
  margin-top: 4px;
}
.activity-arrow {
  font-size: 16px;
  color: var(--mixm-faint);
}
.empty-experiments {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  min-height: 230px;
  border: 1px solid var(--mixm-border);
  padding: 30px;
  background: var(--mixm-surface);
  border-radius: 5px;
}
.empty-number {
  font-family: var(--mixm-font-display);
  font-size: 40px;
  line-height: 1;
  color: var(--mixm-accent);
  margin-bottom: 18px;
}
.empty-experiments h3 {
  font-family: var(--mixm-font-display);
  font-size: 21px;
  font-weight: 400;
  margin-bottom: 10px;
}
.empty-experiments p {
  font-size: 12px;
  color: var(--mixm-muted);
  margin-bottom: 22px;
}
.empty-activity {
  padding: 26px 24px;
  background: var(--mixm-surface-muted);
  min-height: 230px;
  border-radius: 5px;
}
.empty-activity .eyebrow {
  font-size: 9px;
}
.empty-activity p {
  font-family: var(--mixm-font-display);
  font-size: 26px;
  margin: 24px 0;
  line-height: 1.5;
}
.empty-activity > span:last-child {
  font-size: 11px;
  color: var(--mixm-muted);
}
.explore-section {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-top: 1px solid var(--mixm-border);
  border-bottom: 1px solid var(--mixm-border);
}
.explore-entry {
  padding: 24px;
}
.explore-entry:first-child {
  padding-left: 0;
}
.explore-entry + .explore-entry {
  border-left: 1px solid var(--mixm-border);
}
.entry-index {
  font-size: 9px;
  letter-spacing: 0.1em;
  color: var(--mixm-faint);
}
.explore-entry h3 {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-family: var(--mixm-font-display);
  font-size: 21px;
  font-weight: 400;
  margin: 14px 0 8px;
  transition: color 0.2s;
}
.explore-entry h3 span {
  font-family: var(--mixm-font-body);
  color: var(--mixm-faint);
}
.explore-entry:hover h3,
.explore-entry:hover h3 span {
  color: var(--mixm-accent);
}
.explore-entry p {
  font-size: 11px;
  color: var(--mixm-muted);
}
.home-footer {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-top: 25px;
  font-size: 10px;
  color: var(--mixm-muted);
}
.home-footer .eyebrow {
  font-size: 8px;
}
@media (max-width: 1100px) {
  .home-art {
    min-width: 150px;
    width: 24%;
  }
  .home-columns {
    gap: 24px;
    grid-template-columns: minmax(0, 1.5fr) minmax(210px, 1fr);
  }
  .instance-grid {
    grid-template-columns: 1fr;
  }
  .stat-item {
    padding: 22px 16px;
    gap: 14px;
  }
  .stat-caption {
    display: none;
  }
}
@media (max-width: 680px) {
  .home-art {
    display: none;
  }
  .home-hero {
    padding: 0 0 28px;
  }
  .hero-copy > .eyebrow {
    margin-bottom: 22px;
  }
  .hero-copy h1 {
    font-size: 32px;
  }
  .hero-actions {
    gap: 16px;
  }
  .stats-bar {
    margin-bottom: 30px;
  }
  .stat-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 20px 14px;
  }
  .stat-value {
    font-size: 34px;
  }
  .stat-label {
    font-size: 11px;
  }
  .home-columns {
    grid-template-columns: 1fr;
    gap: 30px;
    margin-bottom: 30px;
  }
  .section-header h2 {
    font-size: 23px;
  }
  .empty-experiments,
  .empty-activity {
    min-height: 200px;
  }
  .explore-section {
    grid-template-columns: 1fr;
  }
  .explore-entry,
  .explore-entry:first-child {
    padding: 22px 0;
  }
  .explore-entry + .explore-entry {
    border-left: 0;
    border-top: 1px solid var(--mixm-border);
  }
  .explore-entry h3 {
    margin-top: 10px;
  }
}
</style>
