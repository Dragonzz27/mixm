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
        <!-- 新增：侧边栏折叠/展开按钮 -->
        <n-button secondary @click="toggleSidebar">
          <template #icon>
            <n-icon>
              <EyeOffOutline v-if="isSidebarOpen" />
              <EyeOutline v-else />
            </n-icon>
          </template>
          {{ isSidebarOpen ? '隐藏信息' : '显示信息' }}
        </n-button>
        
        <!-- <n-tag :type="statusType" size="medium">{{ instance.status }}</n-tag> -->
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
      <!-- 修改：动态绑定 class，添加 is-collapsed 状态 -->
      <aside class="info-panel" :class="{ 'is-collapsed': !isSidebarOpen }">
        <!-- 新增：内部包裹层，防止动画过程中的文字换行挤压 -->
        <div class="info-panel-inner">
          <n-card title="实例信息" :bordered="false">
            <n-descriptions :column="1" label-placement="left">
              <n-descriptions-item label="ID">{{ instance.id }}</n-descriptions-item>
              <n-descriptions-item label="名称">{{ instance.name }}</n-descriptions-item>
              <n-descriptions-item label="状态">
                <n-tag :type="statusType" size="small">{{ instance.status }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="模板">{{ instance.template_id || 'N/A' }}</n-descriptions-item>
              <n-descriptions-item label="创建时间">{{ formatTime(instance.created_at) }}</n-descriptions-item>
              <n-descriptions-item label="更新时间">{{ formatTime(instance.updated_at) }}</n-descriptions-item>
            </n-descriptions>
          </n-card>

          <n-divider style="margin: 8px 0;" />

          <n-card title="资源" :bordered="false">
            <n-descriptions :column="1" label-placement="left">
              <n-descriptions-item label="CPU">{{ instance.config?.cpu_limit }}</n-descriptions-item>
              <n-descriptions-item label="内存">{{ instance.config?.memory_limit  }}</n-descriptions-item>
              <!-- <n-descriptions-item label="副本数">{{ instance.replicas || 1 }}</n-descriptions-item> -->
              <!-- <n-descriptions-item label="端口">{{ instance.port || 18800 }}</n-descriptions-item> -->
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage, useDialog } from 'naive-ui'
import { instanceApi } from '@/api'
// 导入新的图标
import { 
  ArrowBackOutline, 
  ExpandOutline, 
  ContractOutline,
  EyeOutline,
  EyeOffOutline
} from '@vicons/ionicons5'

// ... (保留原有的路由和 naive-ui 钩子) ...
const router = useRouter()
const route = useRoute()
const message = useMessage()
const dialog = useDialog()

const previewPanel = ref(null)
const isFullscreen = ref(false)
const isIframeLoading = ref(true)

// 新增：控制侧边栏折叠状态的变量
const isSidebarOpen = ref(true)

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
  endpoint: null
})

// 新增：切换侧边栏状态的方法
function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}

// ... (完全保留原有的 computed 和方法: statusType, codeServerUrl, toggleFullscreen 等) ...
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
  const hostIp = instance.value.hostIp
  const endpoint = instance.value.endpoint || `/${instance.value.id}`
  const port = instance.value.port || 80
  let config = {}
  try {
    config = instance.value.config
  } catch (e) {
    console.error('Failed to parse config:', e)
  }
  console.log('Parsed config:', config)
  const password = config.password || ''
  show_code_server_password.value = password
  return `http://${hostIp}:${port}${endpoint}/?password=${encodeURIComponent(password)}`
})

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
/* 原有的基础布局保持不变 */
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

/* ================= 核心修改区域 ================= */
.detail-content {
  display: flex;
  flex: 1;
  /* 移除原本的 gap: 12px; 改用 info-panel 的 margin-right 实现间距，方便折叠动画 */
  overflow: hidden; 
}

.info-panel {
  width: 300px;
  margin-right: 12px; /* 替代原本的 gap */
  flex-shrink: 0;
  overflow-x: hidden; /* 折叠时隐藏溢出内容 */
  overflow-y: auto;
  /* 添加平滑动画效果 */
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  opacity: 1;
}

/* 折叠状态的样式 */
.info-panel.is-collapsed {
  width: 0;
  margin-right: 0; /* 折叠后让出间距空间 */
  opacity: 0; /* 淡出效果 */
}

/* 内部固定宽度的容器：确保动画收缩时，内部文字不会因为父级变窄而疯狂换行抖动 */
.info-panel-inner {
  width: 300px; 
  display: flex;
  flex-direction: column;
}
/* ============================================== */

/* 右侧面板原有样式保持不变 */
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