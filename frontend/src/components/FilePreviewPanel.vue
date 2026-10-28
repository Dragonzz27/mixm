<template>
  <div class="preview-container">
    <div class="preview-toolbar">
      <n-space>
        <n-button type="primary" @click="handleDownload">
          <template #icon>
            <n-icon><DownloadOutline /></n-icon>
          </template>
          下载文件
        </n-button>
        <n-button-group v-if="isWordFile">
          <n-button @click="handleZoomOut">缩小</n-button>
          <n-button @click="handleZoomReset">{{ zoomPercent }}%</n-button>
          <n-button @click="handleZoomIn">放大</n-button>
        </n-button-group>
      </n-space>
      <span class="file-info-text">{{ fileData?.file_name }}</span>
    </div>
    <div class="preview-content">
      <div
        v-if="isWordFile && !wordPreviewFailed"
        ref="previewContainerRef"
        :style="docxPreviewStyle"
        class="preview-frame docx-preview"
      />
      <iframe
        v-else-if="fileUrl && isIframePreviewable"
        :src="fileUrl"
        frameborder="0"
        class="preview-frame"
      />
      <video
        v-else-if="fileUrl && isVideoFile"
        :src="fileUrl"
        controls
        class="preview-frame preview-video"
      />
      <div
        v-else-if="isMarkdownFile"
        class="preview-frame preview-text markdown-body"
        v-html="markdownContent"
      />
      <pre
        v-else-if="isTextFile"
        class="preview-frame preview-text plain-text"
      >{{ textContent }}</pre>
      <div v-else class="preview-not-supported">
        <n-empty description="此文件格式暂不支持在线预览">
          <template #icon>
            <n-icon><DocumentTextOutline /></n-icon>
          </template>
          <template #extra>
            <n-button type="primary" @click="handleDownload">下载文件查看</n-button>
          </template>
        </n-empty>
      </div>
      <div v-if="loading" class="preview-loading">
        <n-spin size="large" :description="loadingText" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { marked } from 'marked'
import { NButton, NButtonGroup, NSpace, NIcon, NSpin, NEmpty } from 'naive-ui'
import { DocumentTextOutline, DownloadOutline } from '@vicons/ionicons5'
import { renderAsync } from 'docx-preview'

const props = defineProps({
  fileData: {
    type: Object,
    default: null
  }
})

const fileUrl = ref('')
const loading = ref(false)
const loadingText = ref('正在加载文件...')
const previewContainerRef = ref(null)
const fileExt = ref('')
const wordPreviewFailed = ref(false)
const zoomScale = ref(1)
const markdownContent = ref('')
const textContent = ref('')

const MIN_ZOOM = 0.5
const MAX_ZOOM = 2
const ZOOM_STEP = 0.1
const iframePreviewTypes = ['pdf']
const textPreviewTypes = ['txt']
const markdownPreviewTypes = ['md']
const videoPreviewTypes = ['mp4', 'webm', 'ogg', 'mov', 'm4v']

const isWordFile = computed(() => ['doc', 'docx'].includes(fileExt.value))
const isIframePreviewable = computed(() => iframePreviewTypes.includes(fileExt.value))
const isTextFile = computed(() => textPreviewTypes.includes(fileExt.value))
const isMarkdownFile = computed(() => markdownPreviewTypes.includes(fileExt.value))
const isVideoFile = computed(() => videoPreviewTypes.includes(fileExt.value))
const zoomPercent = computed(() => Math.round(zoomScale.value * 100))
const docxPreviewStyle = computed(() => ({ zoom: zoomScale.value }))

watch(() => props.fileData, async (newVal) => {
  resetPreview()
  if (!newVal) return

  fileUrl.value = newVal.file_url || ''
  const fileName = newVal.file_name || ''
  fileExt.value = fileName.includes('.') ? fileName.split('.').pop().toLowerCase() : ''
  await loadCurrentPreview()
}, { immediate: true })

async function loadCurrentPreview() {
  if (!fileUrl.value) return

  if (isWordFile.value) {
    await loadWordPreview()
    return
  }

  if (isMarkdownFile.value || isTextFile.value) {
    await loadTextPreview()
  }
}

async function loadWordPreview() {
  await nextTick()
  if (!previewContainerRef.value) {
    wordPreviewFailed.value = true
    return
  }

  loading.value = true
  loadingText.value = '正在加载 Word 预览...'

  try {
    const res = await fetch(fileUrl.value)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const blob = await res.blob()
    previewContainerRef.value.innerHTML = ''
    await renderAsync(blob, previewContainerRef.value)
  } catch (err) {
    console.error('Word preview error:', err)
    wordPreviewFailed.value = true
  } finally {
    loading.value = false
  }
}

async function loadTextPreview() {
  loading.value = true
  loadingText.value = isMarkdownFile.value ? '正在加载 Markdown 预览...' : '正在加载文本预览...'

  try {
    const res = await fetch(fileUrl.value)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const text = await res.text()
    if (isMarkdownFile.value) {
      markdownContent.value = marked.parse(text)
    } else {
      textContent.value = text
    }
  } catch (err) {
    console.error('Text preview error:', err)
    markdownContent.value = ''
    textContent.value = '预览加载失败，请下载后查看。'
  } finally {
    loading.value = false
  }
}

function resetPreview() {
  fileUrl.value = ''
  loading.value = false
  loadingText.value = '正在加载文件...'
  fileExt.value = ''
  wordPreviewFailed.value = false
  zoomScale.value = 1
  markdownContent.value = ''
  textContent.value = ''
  if (previewContainerRef.value) {
    previewContainerRef.value.innerHTML = ''
  }
}

function handleZoomIn() {
  zoomScale.value = Math.min(MAX_ZOOM, Number((zoomScale.value + ZOOM_STEP).toFixed(2)))
}

function handleZoomOut() {
  zoomScale.value = Math.max(MIN_ZOOM, Number((zoomScale.value - ZOOM_STEP).toFixed(2)))
}

function handleZoomReset() {
  zoomScale.value = 1
}

function handleDownload() {
  if (!fileUrl.value || !props.fileData?.file_name) return

  const link = document.createElement('a')
  link.href = fileUrl.value
  link.download = props.fileData.file_name
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>

<style scoped>
.preview-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 420px;
}

.preview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--mixm-border);
  margin-bottom: 12px;
  gap: 12px;
}

.file-info-text {
  overflow: hidden;
  color: var(--mixm-muted);
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-content {
  position: relative;
  flex: 1;
  overflow: hidden;
  background: var(--mixm-surface-muted);
  border-radius: 4px;
}

.preview-frame {
  width: 100%;
  height: 100%;
  border: none;
}

.docx-preview {
  overflow: auto;
  background: var(--mixm-surface);
}

.preview-video {
  background: #000;
}

.preview-text {
  overflow: auto;
  box-sizing: border-box;
  padding: 24px;
  background: var(--mixm-surface);
}

.plain-text {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: Consolas, Monaco, monospace;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin-top: 0;
}

.markdown-body :deep(pre) {
  padding: 12px;
  overflow: auto;
  background: var(--mixm-surface-muted);
  border-radius: 6px;
}

.markdown-body :deep(code) {
  font-family: Consolas, Monaco, monospace;
}

.preview-loading {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(247 245 239 / 85%);
  z-index: 1;
}

.preview-not-supported {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}
</style>
