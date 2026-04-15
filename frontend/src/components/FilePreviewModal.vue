<template>
  <n-modal
    v-model:show="showModal"
    preset="card"
    :title="fileData?.file_name || '文件预览'"
    style="width: 90vw; max-width: 1400px; height: 85vh;"
    :closable="true"
    @close="handleClose"
  >
    <div class="preview-container">
      <div class="preview-toolbar">
        <n-space>
          <n-button type="primary" @click="handleDownload">
            <template #icon>
              <n-icon><DownloadOutline /></n-icon>
            </template>
            下载文件 {{ fileData?.file_name }}
          </n-button>
          <!-- <n-button @click="handleOpenInNewTab">
            <template #icon>
              <n-icon><OpenOutline /></n-icon>
            </template>
            新窗口打开 
          </n-button> -->
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
        <div v-else class="preview-not-supported">
          <n-empty description="此文件格式不支持在线预览">
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
  </n-modal>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { NModal, NButton, NButtonGroup, NSpace, NIcon, NSpin, NEmpty } from 'naive-ui'
import { DocumentTextOutline, DownloadOutline, OpenOutline } from '@vicons/ionicons5'
import { renderAsync } from 'docx-preview'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  fileData: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:show', 'close'])

const showModal = ref(false)
const fileUrl = ref('')
const fileData = ref(null)
const loading = ref(false)
const loadingText = ref('正在加载文档...')
const previewContainerRef = ref(null)
const fileExt = ref('')
const wordPreviewFailed = ref(false)
const zoomScale = ref(1)

const MIN_ZOOM = 0.5
const MAX_ZOOM = 2
const ZOOM_STEP = 0.1

const isWordFile = computed(() => {
  return fileExt.value === 'docx' || fileExt.value === 'doc'
})

const iframePreviewTypes = ['pdf', 'txt', 'jpg', 'jpeg', 'png', 'gif']

const isIframePreviewable = computed(() => {
  return iframePreviewTypes.includes(fileExt.value)
})

const zoomPercent = computed(() => Math.round(zoomScale.value * 100))

const docxPreviewStyle = computed(() => ({
  zoom: zoomScale.value
}))

watch(() => props.show, (newVal) => {
  showModal.value = newVal

  if (newVal && isWordFile.value) {
    loadWordPreview()
  }

  // 当弹窗关闭时，重置状态
  if (!newVal) {
    resetPreview()
  }
})

watch(() => props.fileData, (newVal) => {
  if (newVal) {
    initPreview(newVal)
  }
}, { immediate: true })

function initPreview(data) {
  fileData.value = data
  wordPreviewFailed.value = false
  zoomScale.value = 1
  
  // 处理文件 URL
  let url = data.file_url 
  
  fileUrl.value = url
  
  if (!url || !data.file_name) {
    return
  }
  
  const fileName = data.file_name || ''
  fileExt.value = fileName.includes('.') ? fileName.split('.').pop().toLowerCase() : ''

  // 如果是 Word 文件且弹窗已打开，立即加载预览
  if (isWordFile.value && showModal.value) {
    loadWordPreview()
  }
}

async function loadWordPreview() {
  if (!fileUrl.value || !isWordFile.value) return

  wordPreviewFailed.value = false

  await nextTick()

  if (!previewContainerRef.value) {
    wordPreviewFailed.value = true
    return
  }

  loading.value = true

  try {
    const res = await fetch(fileUrl.value)
    if (!res.ok) {
      throw new Error('HTTP ' + res.status)
    }

    const blob = await res.blob()
    if (!blob || blob.size === 0) {
      throw new Error('Empty file')
    }

    previewContainerRef.value.innerHTML = ''
    await renderAsync(blob, previewContainerRef.value)
  } catch (err) {
    console.error('Word preview error:', err)
    wordPreviewFailed.value = true
  } finally {
    loading.value = false
  }
}

function handleClose() {
  emit('update:show', false)
  emit('close')
  resetPreview()
}

function resetPreview() {
  fileData.value = null
  fileUrl.value = ''
  loading.value = false
  fileExt.value = ''
  wordPreviewFailed.value = false
  zoomScale.value = 1
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
  if (!fileUrl.value || !fileData.value?.file_name) return
  
  // 获取完整 URL 进行下载
  let downloadUrl = fileUrl.value
  if (!downloadUrl.startsWith('http')) {
    downloadUrl = window.location.origin + downloadUrl
  }
  
  const link = document.createElement('a')
  link.href = downloadUrl
  link.download = fileData.value.file_name
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

function handleOpenInNewTab() {
  if (!fileUrl.value) return
  let openUrl = fileUrl.value
  if (!openUrl.startsWith('http')) {
    openUrl = window.location.origin + openUrl
  }
  window.open(openUrl, '_blank')
}
</script>

<style scoped>
.preview-container {
  display: flex;
  flex-direction: column;
  height: 75vh;
}

.preview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid #e0e0e6;
  margin-bottom: 12px;
}

.file-info-text {
  font-size: 13px;
  color: #666;
}

.preview-content {
  position: relative;
  flex: 1;
  overflow: hidden;
  background: #f5f5f5;
  border-radius: 4px;
}

.preview-frame {
  width: 100%;
  height: 100%;
  border: none;
}

.docx-preview {
  overflow: auto;
  background: white;
}

.docx-preview-container {
  padding: 20px;
}

.docx-preview-container table {
  border-collapse: collapse;
  width: 100%;
}

.docx-preview-container p {
  margin: 10px 0;
}

.preview-loading {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(245, 245, 245, 0.85);
  z-index: 1;
}

.preview-not-supported {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}
</style>
