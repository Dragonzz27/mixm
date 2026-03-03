<template>
  <div class="instance-detail fullscreen-page">
    <div class="detail-header">
      <router-link to="/app/my-instances" class="back-link">
        <n-icon>
          <ArrowBackOutline />
        </n-icon>
        返回
      </router-link>
      <h1>{{ instance.name }}</h1>
      <div class="header-actions">
        <n-button secondary @click="showConfigModal = true">
          <template #icon>
            <n-icon><SettingsOutline /></n-icon>
          </template>
          配置
        </n-button>
        <n-button secondary @click="toggleSidebar">
          <template #icon>
            <n-icon>
              <EyeOffOutline v-if="isSidebarOpen" />
              <EyeOutline v-else />
            </n-icon>
          </template>
          {{ isSidebarOpen ? '隐藏信息' : '显示信息' }}
        </n-button>
        
        <n-button type="primary" @click="handleStart" v-if="instance.status !== 'running'">
          启动
        </n-button>
        <n-button type="warning" @click="handleStop" v-if="instance.status === 'running'">
          暂停
        </n-button>
        <n-button type="error" @click="handleDelete">删除</n-button>
      </div>
    </div>

    <div class="detail-content">
      <aside class="info-panel" :class="{ 'is-collapsed': !isSidebarOpen }">
        <div class="info-panel-inner">
          <n-card title="实例信息" :bordered="false">
            <n-descriptions :column="1" label-placement="left">
              <n-descriptions-item label="ID">{{ instance.id }}</n-descriptions-item>
              <n-descriptions-item label="名称">{{ instance.name }}</n-descriptions-item>
              <n-descriptions-item label="状态">
                <n-tag :type="statusType" size="small">{{ instance.status }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="镜像">{{ instance.image || 'N/A' }}</n-descriptions-item>
              <n-descriptions-item label="创建时间">{{ formatTime(instance.created_at) }}</n-descriptions-item>
              <n-descriptions-item label="更新时间">{{ formatTime(instance.updated_at) }}</n-descriptions-item>
            </n-descriptions>
          </n-card>

          <n-divider style="margin: 8px 0;" />

          <n-card title="资源" :bordered="false">
            <n-descriptions :column="1" label-placement="left">
              <n-descriptions-item label="CPU">{{ instance.config?.cpu_limit || instance.config?.cpuLimit }}</n-descriptions-item>
              <n-descriptions-item label="内存">{{ instance.config?.memory_limit || instance.config?.memoryLimit }}</n-descriptions-item>
              <n-descriptions-item label="存储">{{ instance.config?.storage_limit || instance.config?.storageLimit }} Gi</n-descriptions-item>
              <n-descriptions-item label="GPU">{{ instance.config?.gpu_limit || instance.config?.gpuLimit || 0 }} 张</n-descriptions-item>
              <n-descriptions-item label="单卡显存">{{ formatGpuMemory(instance.config?.gpu_memory_limit || instance.config?.gpuMemoryLimit) }}</n-descriptions-item>
            </n-descriptions>
          </n-card>

          <n-divider style="margin: 8px 0;" />

          <n-card title="访问信息" :bordered="false">
            <n-descriptions :column="1" label-placement="left">
              <n-descriptions-item label="访问地址">
                <n-a :href="codeServerUrl" target="_blank">{{ codeServerUrl }}</n-a>
              </n-descriptions-item>
              <n-descriptions-item label="密码">{{ show_code_server_password }}</n-descriptions-item>
            </n-descriptions>
          </n-card>
        </div>
      </aside>

      <main class="preview-panel" :class="{ 'is-fullscreen': isFullscreen }" ref="previewPanel">
        <n-button 
          class="fullscreen-toggle" 
          circle 
          secondary
          strong
          @click="toggleFullscreen"
        >
          <template #icon>
            <n-icon>
              <ExpandOutline v-if="!isFullscreen" />
              <ContractOutline v-else />
            </n-icon>
          </template>
        </n-button>

        <div class="preview-content">
          <div v-if="instance.status !== 'running'" class="preview-placeholder">
            <n-empty description="实例未运行">
              <template #extra>
                <p class="placeholder-text">请启动实例以访问 code-server</p>
              </template>
            </n-empty>
          </div>

          <template v-else>
            <div v-if="isIframeLoading" class="preview-loading">
              <n-spin size="large">
                <template #description>
                  <span style="color: #64748b">正在加载 code-server...</span>
                </template>
              </n-spin>
            </div>
            <iframe 
              v-show="!isIframeLoading" 
              :src="codeServerUrl" 
              class="code-server-frame"
              @load="handleIframeLoad"
              allow="clipboard-read; clipboard-write"
            ></iframe>
          </template>
        </div>
      </main>
    </div>

    <InstanceConfigModal
      v-model:show="showConfigModal"
      :config="instanceConfig"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage, useDialog } from 'naive-ui'
import { instanceApi } from '@/api'
import { 
  ArrowBackOutline, 
  ExpandOutline, 
  ContractOutline,
  EyeOutline,
  EyeOffOutline,
  SettingsOutline
} from '@vicons/ionicons5'
import InstanceConfigModal from '@/components/InstanceConfigModal.vue'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const dialog = useDialog()

const previewPanel = ref(null)
const isFullscreen = ref(false)
const isIframeLoading = ref(true)
const isSidebarOpen = ref(true)
const showConfigModal = ref(false)

const instance = ref({
  id: route.params.id,
  name: '加载中...',
  status: 'unknown',
  template_id: null,
  created_at: null,
  updated_at: null,
  cpu: null,
  memory: null,
  replicas: null,
  port: null,
  ip: null,
  endpoint: null,
  image: null
})

const instanceConfig = computed(() => {
  const config = instance.value.config || {}
  return {
    id: instance.value.id,
    name: instance.value.name,
    image: instance.value.image || '-',
    cpuLimit: config.cpu_limit || config.cpuLimit || '-',
    memoryLimit: config.memory_limit || config.memoryLimit || '-',
    storageLimit: config.storage_limit || config.storageLimit || '-',
    gpuLimit: config.gpu_limit || config.gpuLimit || 0,
    gpuMemoryLimit: config.gpu_memory_limit || config.gpuMemoryLimit || 0,
    namespace: instance.value.namespace || 'default',
    sshConnection: `ssh ${config.sshUsername || 'root'}@${instance.value.ip || 'localhost'} -p ${config.sshPort || 22}`,
    sshUsername: config.sshUsername || 'root',
    sshPort: config.sshPort || 22,
    sshPassword: config.sshPassword || '-',
    sshHost: instance.value.ip || '',
    endpoint: instance.value.endpoint || '',
    accessPort: instance.value.port || 80,
    codeServerPassword: config.password || ''
  }
})

function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}

const statusType = computed(() => {
  switch (instance.value.status) {
    case 'running': return 'success'
    case 'stopped': return 'warning'
    case 'failed': return 'error'
    default: return 'default'
  }
})

const show_code_server_password = ref("")

const codeServerUrl = computed(() => {
  if (instance.value.status !== 'running') return ''
  const hostIp = instance.value.ip
  const endpoint = instance.value.endpoint || `/${instance.value.id}`
  const port = instance.value.port || 80
  let config = {}
  try {
    config = instance.value.config || {}
  } catch (e) {
    console.error('Failed to parse config:', e)
  }
  console.log('Parsed config:', config)
  const password = config.password || ''
  show_code_server_password.value = password
  return `http://${hostIp}:${port}${endpoint}/?password=${encodeURIComponent(password)}`
})

function formatGpuMemory(mb) {
  if (!mb) return '-'
  const memoryMb = Number(mb) || 0
  if (memoryMb <= 0) return '-'
  if (memoryMb % 1024 === 0) {
    return `${memoryMb / 1024}Gi`
  }
  return `${memoryMb}Mi`
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    previewPanel.value?.requestFullscreen()
    isFullscreen.value = true
  } else {
    document.exitFullscreen()
    isFullscreen.value = false
  }
}

function handleFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

function handleIframeLoad() {
  isIframeLoading.value = false
}

async function fetchInstance() {
  try {
    const res = await instanceApi.getDetail(route.params.id)
    const oldStatus = instance.value.status
    instance.value = { ...instance.value, ...res }
    instance.value.config = instance.value.config ? JSON.parse(instance.value.config) : {}
    if (oldStatus !== 'running' && instance.value.status === 'running') {
      isIframeLoading.value = true
    }
  } catch (err) {
    console.error('获取实例失败:', err)
    message.error('加载实例详情失败')
  }
}

function formatTime(timestamp) {
  if (!timestamp) return 'N/A'
  return new Date(timestamp).toLocaleString()
}

async function handleStart() {
  try {
    await instanceApi.start(route.params.id)
    message.success('实例已启动')
    fetchInstance()
  } catch (err) {
    message.error('启动实例失败')
  }
}

async function handleStop() {
  try {
    await instanceApi.stop(route.params.id)
    message.success('实例已停止')
    fetchInstance()
  } catch (err) {
    message.error('停止实例失败')
  }
}

async function handleDelete() {
  dialog.warning({
    title: '删除实例',
    content: `确定要删除实例 "${instance.value.name}" 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await instanceApi.delete(route.params.id)
        message.success('实例已删除')
        router.push('/app/my-instances')
      } catch (err) {
        message.error('删除实例失败')
      }
    }
  })
}

onMounted(() => {
  fetchInstance()
  document.addEventListener('fullscreenchange', handleFullscreenChange)
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
})
</script>

<style scoped>
.instance-detail {
  display: flex;
  flex-direction: column;
  padding: 20px 24px 0;
  height: calc(100vh - 112px);
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e0e0e6;
  margin-bottom: 16px;
}

.back-link {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #666;
  text-decoration: none;
}
.back-link:hover { color: #18a058; }

.detail-header h1 {
  flex: 1;
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-content {
  display: flex;
  flex: 1;
  overflow: hidden; 
}

.info-panel {
  width: 300px;
  margin-right: 12px;
  flex-shrink: 0;
  overflow-x: hidden;
  overflow-y: auto;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  opacity: 1;
}

.info-panel.is-collapsed {
  width: 0;
  margin-right: 0;
  opacity: 0;
}

.info-panel-inner {
  width: 300px; 
  display: flex;
  flex-direction: column;
}

.preview-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #0f172a;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.preview-panel.is-fullscreen {
  border-radius: 0;
}

.fullscreen-toggle {
  position: absolute;
  top: 1px;
  left: 1px;
  z-index: 100;
  background-color: rgba(255, 255, 255, 1);
  color: #000000;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.fullscreen-toggle:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

.preview-content {
  flex: 1;
  overflow: hidden;
}

.preview-placeholder {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-loading {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
}

.placeholder-text {
  color: #64748b;
  font-size: 14px;
  margin-top: 8px;
}

.code-server-frame {
  width: 100%;
  height: 100%;
  border: none;
  background: #fff;
}
</style>
