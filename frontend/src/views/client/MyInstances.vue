<template>
  <div class="my-instances">
    <div class="page-header">
      <h1>我启动的实例</h1>
      <p>Manage and monitor your instances</p>
    </div>

    <n-spin :show="loading">
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
            <p class="instance-desc">模板: {{ instance.template_name || 'N/A' }}</p>
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
      <n-empty v-else description="No instances started yet" style="padding: 80px 0;">
        <template #extra>
          <n-button type="primary" @click="$router.push('/app')">Back to Home</n-button>
        </template>
      </n-empty>
    </n-spin>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { instanceApi } from '@/api'

const router = useRouter()
const message = useMessage()

const loading = ref(false)
const instances = ref([])

async function fetchMyInstances() {
  try {
    loading.value = true
    const res = await instanceApi.getMyInstances()
    instances.value = res.instances || []
  } catch (err) {
    console.error('Failed to fetch instances:', err)
    message.error('Failed to load instances')
  } finally {
    loading.value = false
  }
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
  fetchMyInstances()
})
</script>

<style scoped>
.my-instances {
  padding-bottom: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.page-header p {
  color: #666;
  font-size: 14px;
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
</style>
