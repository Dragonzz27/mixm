<template>
  <div class="instance-detail fullscreen-page">
    <header class="detail-header">
      <div class="detail-heading">
        <router-link to="/app/my-instances" class="back-link">
          <n-icon><ArrowBackOutline /></n-icon>
          我的实例
        </router-link>
        <h1>{{ instance.name }}</h1>
      </div>
      <div class="header-actions">
        <n-button secondary @click="showConfigModal = true">
          <template #icon>
            <n-icon><SettingsOutline /></n-icon>
          </template>
          配置
        </n-button>
        <n-button
          secondary
          :aria-expanded="isInfoVisible"
          aria-controls="instance-overview"
          @click="isInfoVisible = !isInfoVisible"
        >
          <template #icon>
            <n-icon>
              <EyeOffOutline v-if="isInfoVisible" />
              <EyeOutline v-else />
            </n-icon>
          </template>
          {{ isInfoVisible ? '隐藏信息' : '显示信息' }}
        </n-button>
        <n-button v-if="instance.status !== 'running'" type="primary" @click="handleStart">
          启动
        </n-button>
        <n-button v-else type="warning" @click="handleStop">暂停</n-button>
        <n-button type="error" secondary @click="handleDelete">删除</n-button>
      </div>
    </header>

    <div v-show="isInfoVisible" id="instance-overview" class="overview-grid">
      <section class="detail-panel instance-panel" aria-labelledby="instance-info-title">
        <div class="panel-heading">
          <h2 id="instance-info-title">实例信息</h2>
          <n-tag :type="statusType" size="small" round>{{ instance.status }}</n-tag>
        </div>
        <dl class="instance-fields">
          <div>
            <dt>实例 ID</dt>
            <dd class="mono-text">{{ instance.id }}</dd>
          </div>
          <div>
            <dt>镜像</dt>
            <dd class="mono-text">{{ instance.image || '—' }}</dd>
          </div>
        </dl>
      </section>

      <section class="detail-panel access-panel" aria-labelledby="instance-access-title">
        <div class="panel-heading">
          <h2 id="instance-access-title">访问信息</h2>
        </div>
        <dl class="access-fields">
          <div>
            <dt>访问地址</dt>
            <dd>
              <a
                v-if="codeServerUrl"
                :href="codeServerUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="access-link mono-text"
              >
                {{ codeServerAddress }}
                <span aria-hidden="true">↗</span>
              </a>
              <span v-else class="muted-text">实例启动后可访问</span>
            </dd>
          </div>
          <div>
            <dt>SSH访问</dt>
            <dd class="mono-text">{{ instance.ssh_connection || '—' }}</dd>
          </div>
          <div>
            <dt>访问密码</dt>
            <dd class="mono-text">{{ codeServerPassword || '—' }}</dd>
          </div>
        </dl>
      </section>

      <section class="detail-panel resource-panel" aria-labelledby="instance-resources-title">
        <div class="panel-heading">
          <h2 id="instance-resources-title">实例资源</h2>
        </div>
        <dl class="resource-fields">
          <div>
            <dt>CPU</dt>
            <dd>
              {{ resources.cpuLimit }}
              <span>核</span>
            </dd>
          </div>
          <div>
            <dt>内存</dt>
            <dd>{{ resources.memoryLimit }}</dd>
          </div>
          <div>
            <dt>临时存储</dt>
            <dd>
              {{ resources.storageLimit }}
              <span>Gi</span>
            </dd>
          </div>
          <div>
            <dt>GPU</dt>
            <dd>
              {{ resources.gpuLimit }}
              <span>张</span>
            </dd>
          </div>
          <div>
            <dt>单卡显存</dt>
            <dd>{{ formatGpuMemory(resources.gpuMemoryLimit) }}</dd>
          </div>
        </dl>
      </section>
    </div>

    <section ref="previewPanel" class="preview-panel" aria-labelledby="code-server-title">
      <header class="preview-header">
        <div class="preview-title">
          <h2 id="code-server-title">Code-Server</h2>
          <span class="muted-text">开发工作区</span>
        </div>
        <div class="preview-actions">
          <n-button
            size="small"
            quaternary
            :aria-label="isFullscreen ? '退出全屏' : '全屏显示'"
            :title="isFullscreen ? '退出全屏' : '全屏显示'"
            @click="toggleFullscreen"
          >
            <template #icon>
              <n-icon>
                <ContractOutline v-if="isFullscreen" />
                <ExpandOutline v-else />
              </n-icon>
            </template>
          </n-button>
        </div>
      </header>
      <div class="preview-content">
        <div v-if="instance.status !== 'running'" class="preview-placeholder">
          <n-empty description="实例未运行">
            <template #extra>
              <p class="placeholder-text">启动实例后，即可进入开发工作区。</p>
            </template>
          </n-empty>
        </div>
        <div v-else class="preview-runtime">
          <iframe
            :src="codeServerUrl"
            title="Code-Server 开发工作区"
            class="code-server-frame"
            allow="clipboard-read; clipboard-write"
          ></iframe>
          <button class="iframe-mask" @click="openCodeServerInBrowser">
            <span class="preview-entry">
              <strong>
                进入开发工作区
                <span aria-hidden="true">↗</span>
              </strong>
              <span>在新窗口中打开 Code-Server</span>
            </span>
          </button>
        </div>
      </div>
    </section>

    <InstanceConfigModal v-model:show="showConfigModal" :config="instanceConfig" />
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
import { getInstanceResources, formatGpuMemory } from '@/utils/instanceResources'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const dialog = useDialog()

const previewPanel = ref(null)
const isFullscreen = ref(false)
const isInfoVisible = ref(true)
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

const resources = computed(() => getInstanceResources(instance.value))

const instanceConfig = computed(() => {
  const config = instance.value.config || {}
  const sshHost = instance.value.hostIp || ''
  const sshPort = instance.value.ssh_port || ''
  const sshUsername = instance.value.ssh_username || ''
  const sshConnection = instance.value.ssh_connection || ''
  const endpoint = instance.value.endpoint || `/${instance.value.id}`
  const accessPort = instance.value.port || 80
  const codeServerPassword = config.password || ''

  return {
    id: instance.value.id,
    name: instance.value.name,
    image: instance.value.image || '-',
    ...resources.value,
    namespace: instance.value.namespace || 'default',
    sshHost: sshHost,
    sshPort: sshPort,
    sshUsername: sshUsername,
    sshPassword: instance.value.ssh_password || '',
    sshConnection: sshConnection,
    endpoint: endpoint,
    accessPort: accessPort,
    codeServerPassword: codeServerPassword
  }
})

const statusType = computed(() => {
  switch (instance.value.status) {
    case 'running':
      return 'success'
    case 'stopped':
      return 'warning'
    case 'failed':
      return 'error'
    default:
      return 'default'
  }
})

const codeServerPassword = computed(() => instance.value.config?.password || '')
const codeServerUrl = computed(() => {
  if (instance.value.status !== 'running') return ''
  const hostIp = instance.value.hostIp
  const endpoint = instance.value.endpoint || `/${instance.value.id}`
  const port = instance.value.port || 80
  return `http://${hostIp}:${port}${endpoint}/?password=${encodeURIComponent(codeServerPassword.value)}`
})
const codeServerAddress = computed(() => codeServerUrl.value.split('?')[0])

async function toggleFullscreen() {
  try {
    if (document.fullscreenElement === previewPanel.value) {
      await document.exitFullscreen()
    } else {
      await previewPanel.value?.requestFullscreen()
    }
  } catch (err) {
    message.warning('当前浏览器无法切换全屏')
  }
}

function handleFullscreenChange() {
  isFullscreen.value = document.fullscreenElement === previewPanel.value
}

function openCodeServerInBrowser() {
  if (!codeServerUrl.value) {
    message.warning('实例未运行，无法访问')
    return
  }
  window.open(codeServerUrl.value, '_blank', 'noopener,noreferrer')
}

async function fetchInstance() {
  try {
    const res = await instanceApi.getDetail(route.params.id)
    instance.value = { ...instance.value, ...res }
    instance.value.config = instance.value.config ? JSON.parse(instance.value.config) : {}
  } catch (err) {
    console.error('获取实例失败:', err)
    message.error('加载实例详情失败')
  }
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
  gap: 24px;
  min-height: 100%;
  padding: 24px;
}

.detail-header,
.header-actions,
.panel-heading,
.preview-header,
.preview-title,
.preview-actions {
  display: flex;
  align-items: center;
}

.detail-header {
  justify-content: space-between;
  gap: 20px;
}

.detail-heading {
  min-width: 0;
}
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  color: var(--mixm-muted);
  font-size: 12px;
  text-decoration: none;
}
.back-link:hover {
  color: var(--mixm-accent);
}
.detail-heading h1 {
  margin: 0;
  color: var(--mixm-ink);
  font-family: var(--mixm-font-display);
  font-size: 26px;
  font-weight: 500;
  overflow-wrap: anywhere;
}
.header-actions {
  flex-wrap: wrap;
  gap: 8px;
}

/* 左侧纵向排列信息和访问卡片，资源卡片跨越两行。 */
.overview-grid {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(260px, 1fr);
  grid-template-areas: 'info resources' 'access resources';
  gap: 18px;
}
.instance-panel {
  grid-area: info;
}
.access-panel {
  grid-area: access;
}
.resource-panel {
  grid-area: resources;
}
.detail-panel {
  min-width: 0;
  padding: 20px 24px;
  border: 1px solid var(--mixm-border);
  border-radius: 12px;
  background: var(--mixm-surface);
}
.panel-heading {
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}
.panel-heading h2,
.preview-title h2 {
  margin: 0;
  color: var(--mixm-ink);
  font-size: 16px;
  font-weight: 600;
}
.detail-panel dl {
  margin: 0;
}
.detail-panel dt {
  color: var(--mixm-muted);
  font-size: 12px;
}
.detail-panel dd {
  margin: 0;
  color: var(--mixm-ink);
  font-size: 14px;
  overflow-wrap: anywhere;
}
.instance-fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.3fr);
  gap: 16px 24px;
}
.instance-fields dt {
  margin-bottom: 5px;
}
.mono-text {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', monospace;
}
.muted-text {
  color: var(--mixm-muted);
}
.access-fields {
  display: grid;
  gap: 12px;
}
.access-fields > div {
  display: grid;
  grid-template-columns: 68px minmax(0, 1fr);
  gap: 16px;
  align-items: baseline;
}
.access-link {
  color: var(--mixm-accent);
  text-decoration: none;
}
.access-link:hover {
  text-decoration: underline;
  text-underline-offset: 3px;
}
.resource-fields {
  display: grid;
}
.resource-fields > div {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 0;
  border-bottom: 1px solid var(--mixm-border);
}
.resource-fields > div:first-child {
  padding-top: 0;
}
.resource-fields > div:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}
.resource-fields dd {
  font-size: 20px;
  font-variant-numeric: tabular-nums;
}
.resource-fields dd span {
  color: var(--mixm-muted);
  font-size: 12px;
}

.preview-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 440px;
  overflow: hidden;
  border: 1px solid var(--mixm-border);
  border-radius: 12px;
  background: var(--mixm-surface);
}
.preview-header {
  flex-shrink: 0;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 20px;
  border-bottom: 1px solid var(--mixm-border);
}
.preview-title {
  flex-wrap: wrap;
  gap: 12px;
}
.preview-title > span {
  font-size: 12px;
}
.preview-actions {
  gap: 8px;
  flex-shrink: 0;
}
.preview-content {
  position: relative;
  flex: 1;
  display: flex;
  min-height: 360px;
}
.preview-runtime {
  position: relative;
  flex: 1;
  min-width: 0;
}
.preview-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--mixm-paper);
}
.placeholder-text {
  margin: 8px 0 0;
  color: var(--mixm-muted);
  font-size: 13px;
}
.code-server-frame {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  border: none;
  background: var(--mixm-surface);
}
.iframe-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 24px;
  border: 0;
  background: rgb(41 41 35 / 45%);
  backdrop-filter: blur(2px);
  cursor: pointer;
  font: inherit;
}
.iframe-mask:hover {
  background: rgb(41 41 35 / 35%);
}
.iframe-mask:focus-visible {
  outline: 3px solid var(--mixm-accent);
  outline-offset: -4px;
}
.preview-entry {
  display: grid;
  gap: 8px;
  padding: 24px 32px;
  border: 1px solid var(--mixm-border);
  border-radius: 10px;
  background: var(--mixm-surface);
  text-align: left;
}
.preview-entry strong {
  color: var(--mixm-ink);
  font-size: 18px;
  font-weight: 500;
}
.preview-entry > span {
  color: var(--mixm-muted);
  font-size: 12px;
}
.preview-panel:fullscreen {
  width: 100%;
  height: 100dvh;
  border: 0;
  border-radius: 0;
}
.preview-panel:fullscreen .preview-content {
  min-height: 0;
}

@media (max-width: 1100px) {
  .detail-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 16px;
  }
  .overview-grid {
    grid-template-columns: minmax(0, 1.7fr) minmax(230px, 1fr);
  }
  .instance-fields {
    gap: 12px;
  }
  .detail-panel {
    padding: 20px;
  }
}

@media (max-width: 760px) {
  .instance-fields {
    grid-template-columns: minmax(0, 1fr);
  }
  .instance-detail {
    padding: 20px 16px;
    gap: 18px;
  }
  .detail-heading h1 {
    font-size: 23px;
  }
  .overview-grid {
    grid-template-columns: minmax(0, 1fr);
    grid-template-areas: 'info' 'access' 'resources';
    gap: 14px;
  }
  .header-actions {
    gap: 6px;
  }
  .detail-panel {
    padding: 18px;
  }
  .access-fields > div {
    grid-template-columns: minmax(0, 1fr);
    gap: 5px;
  }
  .resource-fields > div {
    padding: 12px 0;
  }
  .preview-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
  }
  .preview-title {
    justify-content: space-between;
    width: 100%;
  }
  .preview-actions {
    align-self: flex-end;
  }
  .preview-panel {
    min-height: 420px;
  }
  .preview-entry {
    padding: 20px;
  }
}
</style>
