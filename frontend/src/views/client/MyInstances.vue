<template>
  <div class="my-instances">
    <div class="page-header">
      <div>
        <h1>我启动的实例</h1>
        <p>查看实验状态，随时继续你的实践。</p>
      </div>
      <n-button type="primary" @click="showMessage">创建实例</n-button>
      <!-- <n-button type="primary" @click="showStartModal = true">创建实例</n-button> -->
    </div>

    <n-spin :show="loading">
      <n-grid
        cols="1 s:2 l:3"
        responsive="screen"
        :x-gap="20"
        :y-gap="20"
        v-if="instances.length > 0"
      >
        <n-gi v-for="instance in instances" :key="instance.id">
          <n-card :bordered="false" class="instance-card" @click="goToInstance(instance.id)">
            <div class="instance-header">
              <n-tag :type="instance.status === 'running' ? 'success' : 'warning'" size="small">
                {{ instance.status === 'running' ? '运行中' : '已停止' }}
              </n-tag>
              <span class="instance-time">{{ formatTime(instance.created_at) }}</span>
            </div>
            <h3>{{ instance.name }}</h3>
            <p class="instance-desc">镜像: {{ instance.image || '-' }}</p>
            <div class="instance-actions">
              <n-button
                size="small"
                type="primary"
                @click.stop="startInstance(instance.id)"
                v-if="instance.status !== 'running'"
              >
                启动
              </n-button>
              <n-button
                size="small"
                type="warning"
                @click.stop="stopInstance(instance.id)"
                v-if="instance.status === 'running'"
              >
                停止
              </n-button>
              <n-button size="small" type="error" @click.stop="deleteInstance(instance.id)">
                删除
              </n-button>
            </div>
          </n-card>
        </n-gi>
      </n-grid>
      <n-empty v-else description="还没有启动的实例" style="padding: 80px 0">
        <template #extra>
          <n-button type="primary" @click="showStartModal = true">创建第一个实例</n-button>
        </template>
      </n-empty>
    </n-spin>

    <StartInstanceModal v-model:show="showStartModal" @success="fetchMyInstances" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { instanceApi } from '@/api'
import StartInstanceModal from '@/components/StartInstanceModal.vue'

const router = useRouter()
const message = useMessage()

const loading = ref(false)
const instances = ref([])
const showStartModal = ref(false)

async function fetchMyInstances() {
  try {
    loading.value = true
    const res = await instanceApi.getMyInstances()
    instances.value = res.items || []
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

const showMessage = () => {
  message.info(
    'The online instance creation feature is not available yet. Please contact the administrator.'
  )
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: clamp(26px, 3vw, 34px);
  font-weight: 400;
  color: var(--mixm-ink);
  margin-bottom: 8px;
}

.page-header p {
  color: var(--mixm-muted);
  font-size: 14px;
}

.instance-card {
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid var(--mixm-border);
}

.instance-card:hover {
  border-color: var(--mixm-accent);
}

.instance-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.instance-time {
  font-size: 12px;
  color: var(--mixm-faint);
}

.instance-card h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--mixm-ink);
  margin-bottom: 4px;
}

.instance-desc {
  font-size: 13px;
  color: var(--mixm-muted);
  margin-bottom: 16px;
}

.instance-actions {
  display: flex;
  gap: 8px;
}
</style>
