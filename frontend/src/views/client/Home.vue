<template>
  <div class="client-home">
    <section class="stats-bar">
      <div class="stat-item">
        <n-icon :size="24" color="#18a058"><RocketOutline /></n-icon>
        <div class="stat-info">
          <span class="stat-value">{{ stats.total }}</span>
          <span class="stat-label">Total Instances</span>
        </div>
      </div>
      <div class="stat-item">
        <n-icon :size="24" color="#2080f0"><PlayCircleOutline /></n-icon>
        <div class="stat-info">
          <span class="stat-value">{{ stats.running }}</span>
          <span class="stat-label">Running</span>
        </div>
      </div>
      <div class="stat-item">
        <n-icon :size="24" color="#f0a020"><TimeOutline /></n-icon>
        <div class="stat-info">
          <span class="stat-value">{{ stats.stopped }}</span>
          <span class="stat-label">Stopped</span>
        </div>
      </div>
    </section>

    <section class="section">
      <div class="section-header">
        <h2>我启动的实例</h2>
        <router-link to="/app/my-instances" class="more-link">More</router-link>
      </div>
      <n-spin :show="loadingInstances">
        <n-grid :cols="4" :x-gap="20" :y-gap="20" :item-responsive="true" v-if="instances.length > 0">
          <n-gi :span="2" :xs="24" :sm="12" :md="12" :lg="6" v-for="instance in instances" :key="instance.id">
            <n-card :bordered="false" class="instance-card" @click="goToInstance(instance.id)">
              <div class="instance-header">
                <n-tag :type="instance.status === 'running' ? 'success' : 'warning'" size="small">
                  {{ instance.status === 'running' ? 'Running' : 'Stopped' }}
                </n-tag>
                <span class="instance-time">{{ formatTime(instance.created_at) }}</span>
              </div>
              <h3>{{ instance.name }}</h3>
              <p class="instance-desc">Template: {{ instance.template_id || 'N/A' }}</p>
              <div class="instance-actions">
                <n-button size="small" type="primary" @click.stop="startInstance(instance.id)" v-if="instance.status !== 'running'">
                  Start
                </n-button>
                <n-button size="small" type="warning" @click.stop="stopInstance(instance.id)" v-if="instance.status === 'running'">
                  Stop
                </n-button>
                <n-button size="small" type="error" @click.stop="deleteInstance(instance.id)">
                  Delete
                </n-button>
              </div>
            </n-card>
          </n-gi>
        </n-grid>
        <n-empty v-else description="No instances started yet" style="padding: 40px 0;">
          <template #extra>
            <n-button type="primary" @click="$router.push('/app/discovery')">Browse Templates</n-button>
          </template>
        </n-empty>
      </n-spin>
    </section>

    <section class="section recent-section">
      <div class="section-header">
        <h2>最近活动</h2>
      </div>
      <n-spin :show="loadingActivities">
        <div class="activity-list" v-if="recentActivities.length > 0">
          <div class="activity-item" v-for="activity in recentActivities" :key="activity.id" @click="goToActivity(activity)">
            <div class="activity-icon" :style="{ background: getActivityColor(activity.type) }">
              <n-icon :size="18" color="#fff">
                <component :is="getActivityIcon(activity.type)" />
              </n-icon>
            </div>
            <div class="activity-content">
              <div class="activity-title">{{ activity.title }}</div>
              <div class="activity-desc">{{ activity.description }}</div>
            </div>
            <div class="activity-time">{{ formatRelativeTime(activity.time) }}</div>
          </div>
        </div>
        <n-empty v-else description="暂无最近活动" style="padding: 40px 0;" />
      </n-spin>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import { instanceApi } from '@/api'
import { RocketOutline, PlayCircleOutline, TimeOutline, StopCircleOutline, TrashOutline, EyeOutline, CreateOutline } from '@vicons/ionicons5'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const loadingInstances = ref(false)
const loadingActivities = ref(false)
const instances = ref([])
const recentActivities = ref([])

const stats = computed(() => ({
  total: instances.value.length,
  running: instances.value.filter(i => i.status === 'running').length,
  stopped: instances.value.filter(i => i.status !== 'running').length
}))

const activityTypeConfig = {
  instance_start: { color: '#18a058', icon: PlayCircleOutline },
  instance_stop: { color: '#f0a020', icon: StopCircleOutline },
  instance_delete: { color: '#d03050', icon: TrashOutline },
  instance_create: { color: '#2080f0', icon: CreateOutline },
  content_view: { color: '#2080f0', icon: EyeOutline }
}

function getActivityColor(type) {
  return activityTypeConfig[type]?.color || '#666'
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
    instances.value = res.instances || []
    buildRecentActivities()
  } catch (err) {
    console.error('Failed to fetch instances:', err)
    message.error('Failed to load instances')
  } finally {
    loadingInstances.value = false
  }
}

function buildRecentActivities() {
  const activities = []
  
  instances.value.forEach(instance => {
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
    message.success('Instance started')
    fetchMyInstances()
  } catch (err) {
    message.error('Failed to start instance')
  }
}

async function stopInstance(id) {
  try {
    await instanceApi.stop(id)
    message.success('Instance stopped')
    fetchMyInstances()
  } catch (err) {
    message.error('Failed to stop instance')
  }
}

async function deleteInstance(id) {
  try {
    await instanceApi.delete(id)
    message.success('Instance deleted')
    fetchMyInstances()
  } catch (err) {
    message.error('Failed to delete instance')
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
  padding-bottom: 24px;
}

.stats-bar {
  display: flex;
  gap: 32px;
  margin-bottom: 24px;
  padding: 20px 24px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e0e0e6;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: #666;
  margin-top: 4px;
}

.section {
  margin-bottom: 32px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h2 {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.more-link {
  font-size: 14px;
  color: #18a058;
  text-decoration: none;
}

.more-link:hover {
  text-decoration: underline;
}

.instance-card {
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #e0e0e6;
}

.instance-card:hover {
  border-color: #18a058;
}

.instance-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.instance-time {
  font-size: 12px;
  color: #999;
}

.instance-card h3 {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.instance-desc {
  font-size: 13px;
  color: #666;
  margin-bottom: 16px;
}

.instance-actions {
  display: flex;
  gap: 8px;
}

.activity-list {
  background: white;
  border-radius: 8px;
  border: 1px solid #e0e0e6;
  overflow: hidden;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  cursor: pointer;
  transition: background 0.2s;
  border-bottom: 1px solid #f0f0f5;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-item:hover {
  background: #f9fafb;
}

.activity-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.activity-desc {
  font-size: 13px;
  color: #666;
  margin-top: 2px;
}

.activity-time {
  font-size: 12px;
  color: #999;
  white-space: nowrap;
}
</style>
