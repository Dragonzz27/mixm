<template>
  <n-modal :show="show" :mask-closable="!busy" :close-on-esc="!busy" @update:show="updateShow">
    <section
      v-if="assignment"
      class="assignment-dialog"
      role="dialog"
      aria-modal="true"
      aria-labelledby="assignment-dialog-title"
    >
      <header class="dialog-header">
        <div class="dialog-title">
          <span class="dialog-caption">作业详情</span>
          <h2 id="assignment-dialog-title">{{ assignment.title }}</h2>
          <div class="assignment-meta">
            <span class="deadline" :class="{ overdue }">
              <n-icon><CalendarOutline /></n-icon>
              截止时间 {{ formatDate(assignment.deadline) }}
            </span>
            <n-tag size="small" :bordered="false" :type="submissionStatus.type">
              {{ submissionStatus.label }}
            </n-tag>
            <n-tag v-if="overdue" size="small" :bordered="false" type="error">已截止</n-tag>
          </div>
        </div>
        <n-button
          class="dialog-close"
          quaternary
          circle
          :disabled="busy"
          aria-label="关闭作业详情"
          @click="updateShow(false)"
        >
          <template #icon>
            <n-icon><CloseOutline /></n-icon>
          </template>
        </n-button>
      </header>

      <div class="dialog-body">
        <section class="assignment-description" aria-labelledby="assignment-description-title">
          <h3 id="assignment-description-title">作业说明</h3>
          <p>{{ assignment.description || '此作业暂无补充说明，请查看下方附件。' }}</p>
        </section>

        <section class="attachment-section" aria-labelledby="assignment-attachments-title">
          <div class="section-heading">
            <h3 id="assignment-attachments-title">
              作业附件
              <span>{{ attachments.length }}</span>
            </h3>
            <span class="section-hint">选择附件阅读或下载</span>
          </div>
          <div v-if="attachments.length" class="attachment-workspace">
            <nav class="attachment-list" aria-label="作业附件列表">
              <button
                v-for="(attachment, index) in attachments"
                :key="index"
                class="attachment-item"
                :class="{ active: selectedAttachmentIndex === index }"
                :aria-pressed="selectedAttachmentIndex === index"
                aria-controls="assignment-attachment-preview"
                @click="selectedAttachmentIndex = index"
              >
                <n-icon :size="18"><DocumentTextOutline /></n-icon>
                <span class="attachment-label">
                  <strong>{{ attachment.name }}</strong>
                  <span class="attachment-filename">{{ attachment.file_name || '附件' }}</span>
                  <span v-if="attachment.file_size != null" class="attachment-size">
                    {{ formatAttachmentSize(attachment.file_size) }}
                  </span>
                </span>
              </button>
            </nav>
            <div id="assignment-attachment-preview" class="attachment-preview">
              <FilePreviewPanel
                :key="`${assignment.id}:${selectedAttachmentIndex}`"
                :file-data="selectedAttachment"
                class="assignment-preview-panel"
              />
            </div>
          </div>
          <div v-else class="attachment-empty">
            <n-icon :size="28"><DocumentTextOutline /></n-icon>
            <div>
              <strong>此作业暂无附件</strong>
              <p>请根据上方作业说明完成作业。</p>
            </div>
          </div>
        </section>

        <section v-if="localSubmission?.content" class="submitted-text">
          <h3>已提交内容</h3>
          <p>{{ localSubmission.content }}</p>
        </section>
        <section
          v-if="localSubmission && (localSubmission.grade != null || localSubmission.feedback)"
          class="review-panel"
          aria-label="批改结果"
        >
          <div v-if="localSubmission.grade != null" class="grade">
            <span>作业评分</span>
            <strong>
              {{ localSubmission.grade }}
              <small>分</small>
            </strong>
          </div>
          <div v-if="localSubmission.feedback" class="review-feedback">
            <h3>教师反馈</h3>
            <p>{{ localSubmission.feedback }}</p>
          </div>
        </section>
      </div>

      <footer class="submission-footer">
        <div class="section-heading">
          <h3>我的提交</h3>
          <span class="section-hint">
            {{ overdue ? '作业已截止，仅可查看' : '支持 Word、PDF、PPT、TXT 文件' }}
          </span>
        </div>
        <n-alert v-if="formError" type="error" :bordered="false" class="submission-error">
          {{ formError }}
        </n-alert>
        <div class="submission-footer-content">
          <template v-if="localSubmission">
            <div class="submission-file-card submitted-file">
              <n-icon :size="24"><CheckmarkCircleOutline /></n-icon>
              <div class="file-details">
                <strong>{{ localSubmission.file_name || '文本内容' }}</strong>
                <span>提交于 {{ formatDate(localSubmission.submitted_at) }}</span>
              </div>
            </div>
            <div class="submission-actions">
              <n-button
                v-if="localSubmission.file_path || localSubmission.file_url"
                secondary
                @click="previewFile(localSubmission)"
              >
                预览提交文件
              </n-button>
              <n-button
                v-if="!overdue"
                quaternary
                type="error"
                :loading="deleting"
                :disabled="busy"
                @click="deleteAndResubmit"
              >
                删除并重新提交
              </n-button>
            </div>
          </template>
          <template v-else>
            <div class="upload-area">
              <n-upload
                v-show="!uploadFile"
                v-model:file-list="uploadList"
                :max="1"
                :disabled="busy || overdue"
                :custom-request="uploadSubmission"
                :show-file-list="false"
                accept=".doc,.docx,.pdf,.ppt,.pptx,.txt"
                @before-upload="beforeUpload"
              >
                <n-upload-dragger class="submission-dropzone">
                  <n-icon :size="24"><CloudUploadOutline /></n-icon>
                  <div>
                    <strong>{{ overdue ? '已停止接收作业' : '点击或拖拽文件到此处上传' }}</strong>
                    <span>每次提交一个文件</span>
                  </div>
                </n-upload-dragger>
              </n-upload>
              <div v-if="uploadFile" class="submission-file-card">
                <n-icon :size="24"><DocumentTextOutline /></n-icon>
                <div class="file-details">
                  <strong>{{ uploadFile.name }}</strong>
                  <span>
                    {{ formatAttachmentSize(uploadFile.size) }} ·
                    {{ uploading ? '正在上传…' : '已上传，待提交' }}
                  </span>
                </div>
                <div class="file-actions">
                  <n-button
                    v-if="uploadResult"
                    size="small"
                    quaternary
                    @click="previewFile(uploadResult)"
                  >
                    预览
                  </n-button>
                  <n-button size="small" quaternary :disabled="busy" @click="removeFile">
                    移除
                  </n-button>
                </div>
              </div>
            </div>
            <n-button
              type="primary"
              class="submit-button"
              :loading="submitting"
              :disabled="!uploadResult || busy || overdue"
              @click="submitAssignment"
            >
              {{ uploading ? '文件上传中' : '提交作业' }}
            </n-button>
          </template>
        </div>
      </footer>
    </section>
  </n-modal>
  <FilePreviewModal v-model:show="showPreview" :file-data="previewData" />
</template>

<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { NAlert, NButton, NIcon, NModal, NTag, NUpload, NUploadDragger, useMessage } from 'naive-ui'
import {
  CalendarOutline,
  CheckmarkCircleOutline,
  CloseOutline,
  CloudUploadOutline,
  DocumentTextOutline
} from '@vicons/ionicons5'
import { assignmentApi } from '@/api'
import FilePreviewModal from '@/components/FilePreviewModal.vue'
import FilePreviewPanel from '@/components/FilePreviewPanel.vue'
import { formatAttachmentSize, normalizeAssignmentAttachments } from '@/utils/assignmentAttachments'

const props = defineProps({
  show: Boolean,
  assignment: { type: Object, default: null },
  submission: { type: Object, default: null }
})
const emit = defineEmits(['update:show', 'change'])
const message = useMessage()
const localSubmission = ref(null)
const selectedAttachmentIndex = ref(0)
const uploadList = ref([])
const uploadFile = ref(null)
const uploadResult = ref(null)
const uploading = ref(false)
const submitting = ref(false)
const deleting = ref(false)
const formError = ref('')
const showPreview = ref(false)
const previewData = ref(null)
let uploadVersion = 0

const attachments = computed(() =>
  normalizeAssignmentAttachments(props.assignment?.attachments, { keepEmpty: false })
)
const selectedAttachment = computed(() => attachments.value[selectedAttachmentIndex.value])
const overdue = computed(
  () =>
    !!props.assignment &&
    (props.assignment.status !== 'active' || new Date(props.assignment.deadline) < new Date())
)
const busy = computed(() => uploading.value || submitting.value || deleting.value)
const submissionStatus = computed(() => {
  if (!localSubmission.value) return { label: '未提交', type: 'default' }
  const status = localSubmission.value.status
  if (status === 'completed') return { label: '已完成', type: 'success' }
  if (status === 'reviewed' || localSubmission.value.grade != null)
    return { label: '已批改', type: 'success' }
  return { label: '已提交 · 待批改', type: 'info' }
})

function formatDate(value) {
  return value ? new Date(value).toLocaleString('zh-CN') : '未设置'
}

function updateShow(value) {
  if (!busy.value) emit('update:show', value)
}

// 每次打开作业重置文件状态，避免上一份作业的上传结果被带入。
watch(
  [() => props.show, () => props.assignment?.id],
  ([show]) => {
    if (!show) return
    localSubmission.value = props.submission
    selectedAttachmentIndex.value = 0
    removeFile()
    showPreview.value = false
    previewData.value = null
  },
  { immediate: true }
)

watch(
  () => props.submission,
  (submission) => {
    if (props.show) localSubmission.value = submission
  }
)

function removeFile() {
  uploadVersion++
  uploadFile.value = null
  uploadResult.value = null
  uploadList.value = []
  uploading.value = false
  formError.value = ''
}

function beforeUpload({ file }) {
  if (busy.value || overdue.value) return false
  const extension = `.${file.name.split('.').pop().toLowerCase()}`
  if (!['.doc', '.docx', '.pdf', '.ppt', '.pptx', '.txt'].includes(extension)) {
    formError.value = '请上传 Word、PDF、PPT 或 TXT 格式的文件。'
    return false
  }
  return true
}

// 上传成功后才允许提交；失败或移除文件时同步清空上传结果。
async function uploadSubmission({ file, onFinish, onError }) {
  const version = ++uploadVersion
  uploadFile.value = { name: file.name, size: file.file?.size || 0 }
  uploadResult.value = null
  formError.value = ''
  uploading.value = true
  try {
    const result = await assignmentApi.uploadFile(file.file)
    if (version !== uploadVersion) return
    uploadResult.value = result
    onFinish()
  } catch (error) {
    if (version !== uploadVersion) return
    formError.value = error.response?.data?.message || '文件上传失败，请重新选择文件。'
    uploadFile.value = null
    onError()
    uploadList.value = []
  } finally {
    if (version === uploadVersion) uploading.value = false
  }
}

async function submitAssignment() {
  if (!uploadResult.value || busy.value || overdue.value) return
  submitting.value = true
  formError.value = ''
  try {
    await assignmentApi.submit({
      assignment_id: props.assignment.id,
      file_path: uploadResult.value.file_path,
      file_name: uploadResult.value.file_name,
      file_size: uploadResult.value.file_size
    })
    message.success('作业提交成功')
    emit('change')
    emit('update:show', false)
  } catch (error) {
    formError.value = error.response?.data?.message || '提交失败，请重试。'
  } finally {
    submitting.value = false
  }
}

async function deleteAndResubmit() {
  if (!localSubmission.value || busy.value || overdue.value) return
  deleting.value = true
  formError.value = ''
  try {
    await assignmentApi.deleteSubmission(localSubmission.value.id)
    localSubmission.value = null
    removeFile()
    emit('change')
    message.success('已删除原提交，请上传新文件后提交')
  } catch (error) {
    formError.value = error.response?.data?.message || '删除失败，请重试。'
  } finally {
    deleting.value = false
  }
}

function previewFile(file) {
  previewData.value = file
  showPreview.value = true
}

onBeforeUnmount(() => {
  uploadVersion++
})
</script>

<style scoped>
.assignment-dialog {
  display: flex;
  flex-direction: column;
  width: min(1240px, calc(100vw - 48px));
  max-height: calc(100dvh - 64px);
  margin: 32px auto;
  overflow: hidden;
  color: var(--mixm-ink);
  background: var(--mixm-surface);
  border: 1px solid var(--mixm-border);
  border-radius: 12px;
  box-shadow: 0 24px 80px rgb(41 41 35 / 18%);
}
.dialog-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  flex-shrink: 0;
  padding: 24px 28px 20px;
  border-bottom: 1px solid var(--mixm-border);
}
.dialog-title {
  min-width: 0;
}
.dialog-caption,
.section-hint {
  color: var(--mixm-muted);
  font-size: 12px;
}
.dialog-title h2 {
  margin: 4px 0 12px;
  font-family: var(--mixm-font-display);
  font-size: 26px;
  font-weight: 400;
  line-height: 1.4;
  overflow-wrap: anywhere;
}
.dialog-close {
  flex-shrink: 0;
}
.assignment-meta,
.deadline {
  display: flex;
  align-items: center;
  gap: 8px;
}
.assignment-meta {
  flex-wrap: wrap;
  gap: 10px 16px;
}
.deadline {
  color: var(--mixm-muted);
  font-size: 12px;
}
.deadline.overdue {
  color: var(--mixm-error);
}
.dialog-body {
  display: flex;
  flex-direction: column;
  gap: 24px;
  min-height: 0;
  padding: 24px 28px;
  overflow-y: auto;
  overscroll-behavior: contain;
}
.dialog-body > section {
  flex-shrink: 0;
}
.assignment-dialog h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}
.assignment-description p,
.submitted-text p,
.review-feedback p {
  margin: 10px 0 0;
  color: var(--mixm-muted);
  line-height: 1.8;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.section-heading h3 span {
  margin-left: 6px;
  font-size: 12px;
  font-weight: 400;
  color: var(--mixm-muted);
}
.attachment-workspace {
  display: grid;
  grid-template-columns: 210px minmax(0, 1fr);
  height: clamp(320px, 44vh, 520px);
  overflow: hidden;
  border: 1px solid var(--mixm-border);
  border-radius: 8px;
}
.attachment-list {
  overflow-y: auto;
  padding: 8px;
  border-right: 1px solid var(--mixm-border);
  background: var(--mixm-paper);
}
.attachment-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
  padding: 12px 10px;
  text-align: left;
  font: inherit;
  color: var(--mixm-muted);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 5px;
  cursor: pointer;
  transition:
    background-color 0.2s,
    color 0.2s;
}
.attachment-item + .attachment-item {
  margin-top: 4px;
}
.attachment-item:hover {
  background: var(--mixm-surface-muted);
}
.attachment-item.active {
  background: var(--mixm-accent-soft);
  color: var(--mixm-accent);
}
.attachment-item:focus-visible {
  outline: 2px solid var(--mixm-accent);
  outline-offset: -2px;
}
.attachment-item > .n-icon {
  flex-shrink: 0;
  margin-top: 2px;
}
.attachment-label {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}
.attachment-label strong {
  font-size: 13px;
  font-weight: 600;
  overflow-wrap: anywhere;
}
.attachment-filename,
.attachment-size {
  font-size: 11px;
  overflow-wrap: anywhere;
  color: var(--mixm-muted);
}
.attachment-preview {
  min-width: 0;
  min-height: 0;
  padding: 16px;
}
.assignment-preview-panel {
  min-height: 0;
  height: 100%;
}
.attachment-preview :deep(.preview-toolbar) {
  flex-wrap: wrap;
  gap: 8px;
}
.attachment-preview :deep(.preview-toolbar .n-button) {
  height: 28px;
  padding: 0 10px;
  font-size: 12px;
}
.attachment-preview :deep(.file-info-text) {
  font-size: 12px;
}
.attachment-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 32px 20px;
  color: var(--mixm-muted);
  border: 1px dashed var(--mixm-border);
  border-radius: 8px;
  background: var(--mixm-paper);
}
.attachment-empty strong {
  font-size: 13px;
  font-weight: 500;
}
.attachment-empty p {
  margin: 4px 0 0;
  font-size: 12px;
}
.review-panel {
  display: flex;
  gap: 24px;
  padding: 20px;
  background: var(--mixm-success-soft);
  border-radius: 8px;
}
.grade {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  gap: 4px;
  color: var(--mixm-success);
}
.grade > span {
  font-size: 12px;
}
.grade strong {
  font-size: 30px;
  line-height: 1.2;
  font-weight: 500;
}
.grade small {
  margin-left: 4px;
  font-size: 12px;
  font-weight: 400;
}
.review-feedback {
  min-width: 0;
}
.review-feedback p {
  margin-top: 4px;
}
.submission-footer {
  flex-shrink: 0;
  padding: 18px 28px 22px;
  border-top: 1px solid var(--mixm-border);
  background: var(--mixm-paper);
}
.submission-error {
  margin-bottom: 12px;
}
.submission-footer-content {
  display: flex;
  align-items: center;
  gap: 20px;
}
.upload-area {
  flex: 1;
  min-width: 0;
}
.submission-dropzone {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 14px 18px;
  border-radius: 6px;
  background: var(--mixm-surface);
}
.submission-dropzone > div {
  display: flex;
  flex-direction: column;
  gap: 2px;
  text-align: left;
}
.submission-dropzone strong {
  font-size: 13px;
  font-weight: 500;
}
.submission-dropzone span {
  font-size: 12px;
  color: var(--mixm-muted);
}
.submission-file-card {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
  padding: 14px 16px;
  border: 1px solid var(--mixm-border);
  border-radius: 6px;
  background: var(--mixm-surface);
}
.submission-file-card > .n-icon {
  flex-shrink: 0;
  color: var(--mixm-muted);
}
.submitted-file > .n-icon {
  color: var(--mixm-success);
}
.file-details {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  gap: 4px;
}
.file-details strong {
  font-size: 13px;
  font-weight: 500;
  overflow-wrap: anywhere;
}
.file-details > span {
  font-size: 12px;
  color: var(--mixm-muted);
}
.file-actions,
.submission-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.submit-button {
  min-width: 116px;
}
@media (max-width: 760px) {
  .assignment-dialog {
    width: calc(100vw - 24px);
    max-height: calc(100dvh - 24px);
    margin: 12px auto;
    border-radius: 10px;
  }
  .dialog-header {
    padding: 18px;
    gap: 12px;
  }
  .dialog-title h2 {
    font-size: 22px;
  }
  .dialog-body {
    padding: 20px 18px;
    gap: 20px;
  }
  .attachment-workspace {
    display: flex;
    flex-direction: column;
    height: 430px;
  }
  .attachment-list {
    display: flex;
    flex-shrink: 0;
    gap: 6px;
    overflow-x: auto;
    border-right: 0;
    border-bottom: 1px solid var(--mixm-border);
  }
  .attachment-item {
    flex: 0 0 160px;
    width: 160px;
    padding: 10px 8px;
  }
  .attachment-item + .attachment-item {
    margin-top: 0;
  }
  .attachment-filename {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .attachment-preview {
    flex: 1;
    padding: 12px;
  }
  .submission-footer {
    padding: 16px 18px;
  }
  .submission-footer-content {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  .submission-actions {
    justify-content: flex-end;
  }
  .file-actions {
    flex-shrink: 0;
  }
  .review-panel {
    gap: 16px;
    padding: 16px;
  }
}
@media (max-height: 600px) {
  .assignment-dialog {
    max-height: calc(100dvh - 24px);
    margin-block: 12px;
  }
  .dialog-header {
    padding-block: 12px;
  }
  .dialog-title h2 {
    margin: 2px 0 6px;
    font-size: 20px;
  }
  .submission-footer {
    padding-block: 12px;
  }
}
</style>
