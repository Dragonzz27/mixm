<template>
  <div class="client-assignments">
    <div class="page-header">
      <h1>我的作业</h1>
    </div>

    <n-tabs type="line" animated>
      <n-tab-pane name="available" tab="可提交作业">
        <div class="assignment-list">
          <n-grid v-if="availableAssignments.length > 0" :cols="2" :x-gap="20" :y-gap="20">
            <n-gi v-for="assignment in availableAssignments" :key="assignment.id">
              <n-card class="assignment-card" :class="{ overdue: isOverdue(assignment.deadline) }">
                <template #header>
                  <div class="card-header">
                    <span class="title">{{ assignment.title }}</span>
                    <n-tag v-if="isOverdue(assignment.deadline)" type="error" size="small">已过期</n-tag>
                  </div>
                </template>
                <div class="card-body">
                  <p class="description">{{ assignment.description || '暂无描述' }}</p>
                  <div class="deadline">
                    <span class="label">截止时间：</span>
                    <span :class="{ 'text-error': isOverdue(assignment.deadline) }">{{ formatDate(assignment.deadline) }}</span>
                  </div>
                  <div class="submission-status-placeholder">
                    <div v-if="userSubmissionMap[assignment.id]" class="submission-status">
                      <n-tag type="success">已提交</n-tag>
                      <span class="submit-time">提交于：{{ formatDate(userSubmissionMap[assignment.id].submitted_at) }}</span>
                    </div>
                    <div v-else class="submission-status-empty">
                      <n-tag type="default" size="small">未提交</n-tag>
                    </div>
                  </div>
                </div>
                <template #action>
                  <n-space justify="end">
                    <n-button
                      type="primary"
                      @click="viewDetail(assignment)"
                      :disabled="isOverdue(assignment.deadline) && !userSubmissionMap[assignment.id]"
                    >
                      {{ userSubmissionMap[assignment.id] ? '查看提交' : '查看附件和提交作业' }}
                    </n-button>
                  </n-space>
                </template>
              </n-card>
            </n-gi>
          </n-grid>
          <n-empty v-else description="暂无进行中的作业" />
        </div>
      </n-tab-pane>

      <n-tab-pane name="history" tab="提交历史">
        <n-data-table
          :columns="historyColumns"
          :data="mySubmissions"
          :pagination="pagination"
          :bordered="false"
        />
      </n-tab-pane>
    </n-tabs>

    <FilePreviewModal
      v-model:show="showPreviewModal"
      :file-data="previewFileData"
    />

    <n-modal
      v-model:show="showSubmitModal"
      preset="card"
      :title="currentAssignment?.title || '提交作业'"
      style="width: 96vw; max-width: 1480px;"
    >
      <div v-if="currentAssignment" class="submit-form">
        <div class="detail-top">
          <section class="detail-panel assignment-info-panel">
            <div class="section-title">作业信息</div>
            <div class="assignment-summary">
              <span class="summary-description" :title="currentAssignment.description || '暂无描述'">
                {{ currentAssignment.description || '暂无描述' }}
              </span>
              <span class="summary-divider">|</span>
              <span class="summary-deadline">截止：{{ formatDate(currentAssignment.deadline) }}</span>
            </div>
          </section>

          <section class="detail-panel submission-panel">
            <div class="section-title">提交作业</div>
            <div v-if="userSubmission && allowResubmit" class="current-submission-resubmit">
              <div class="submission-summary" :title="userSubmission.file_name || '文本内容'">
                <n-tag type="success" size="small">已提交</n-tag>
                <span class="submission-file">{{ userSubmission.file_name || '文本内容' }}</span>
                <span class="summary-deadline">{{ formatDate(userSubmission.submitted_at) }}</span>
                <span v-if="userSubmission.grade !== null">{{ userSubmission.grade }} 分</span>
              </div>
              <n-space size="small">
                <n-button v-if="userSubmission.file_path || userSubmission.file_url" @click="previewFile(userSubmission)">预览提交文件</n-button>
                <n-button type="warning" @click="deleteAndResubmit">删除并重新提交</n-button>
              </n-space>
            </div>

            <div v-else class="upload-section">
              <div v-if="uploadFile" class="selected-file">
                <div class="file-info">
                  <n-icon :size="18"><DocumentTextOutline /></n-icon>
                  <span class="file-name">{{ uploadFile.name }}</span>
                  <span class="file-size">{{ formatFileSize(uploadFile.size) }}</span>
                  <n-button size="tiny" @click="viewUploadedFile">预览</n-button>
                  <n-button size="tiny" type="error" @click="removeFile">删除</n-button>
                </div>
              </div>

              <div v-else class="upload-area">
                <n-upload
                  ref="uploadRef"
                  :max="1"
                  :custom-request="customRequest"
                  :show-file-list="false"
                  accept=".doc,.docx,.pdf,.ppt,.pptx,.txt"
                  @before-upload="beforeUpload"
                >
                  <div class="upload-dragger">
                    <n-icon :size="18"><CloudUploadOutline /></n-icon>
                    <span class="upload-text">上传作业文件</span>
                    <span class="upload-hint">.doc / .docx / .pdf / .ppt / .pptx / .txt</span>
                  </div>
                </n-upload>
              </div>

              <n-space justify="end" size="small" class="submission-actions">
                <n-button
                  type="primary"
                  size="small"
                  :loading="submitting"
                  :disabled="!uploadResult"
                  @click="handleSubmit"
                >
                  提交作业
                </n-button>
              </n-space>
            </div>
          </section>
        </div>

        <n-divider v-if="currentAttachments.length" />

        <div v-if="currentAttachments.length" class="attachment-panel">
          <div class="section-title">作业附件</div>
          <n-tabs
            v-model:value="selectedAttachmentTab"
            type="line"
            placement="left"
            class="attachment-tabs"
          >
            <n-tab-pane
              v-for="(attachment, index) in currentAttachments"
              :key="`${attachment.file_name}-${index}`"
              :name="String(index)"
              display-directive="if"
            >
              <template #tab>
                <div class="attachment-tab">
                  <div class="attachment-name">{{ attachment.name }}</div>
                  <!-- <div class="attachment-file">
                    {{ attachment.file_name }}
                    <span v-if="attachment.file_size" class="attachment-size">({{ formatAttachmentSize(attachment.file_size) }})</span>
                  </div> -->
                </div>
              </template>
              <FilePreviewPanel :file-data="attachment" class="attachment-preview" />
            </n-tab-pane>
          </n-tabs>
        </div>
        <n-empty v-else class="attachment-empty" description="此作业暂无附件" />
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import {
  NCard,
  NButton,
  NSpace,
  NTag,
  NGrid,
  NGi,
  NTabs,
  NTabPane,
  NEmpty,
  NDivider,
  NUpload,
  NDataTable,
  NIcon,
  useMessage
} from 'naive-ui'
import { DocumentTextOutline, CloudUploadOutline } from '@vicons/ionicons5'
import FilePreviewModal from '@/components/FilePreviewModal.vue'
import FilePreviewPanel from '@/components/FilePreviewPanel.vue'
import { assignmentApi } from '@/api'
import { formatAttachmentSize, normalizeAssignmentAttachments } from '@/utils/assignmentAttachments'

const message = useMessage()
const availableAssignments = ref([])
const mySubmissions = ref([])
const userSubmissionMap = ref({})
const showSubmitModal = ref(false)
const currentAssignment = ref(null)
const userSubmission = ref(null)
const allowResubmit = ref(false)
const uploadRef = ref(null)
const uploadResult = ref(null)
const uploadFile = ref(null)
const submitting = ref(false)
const pagination = ref({ pageSize: 10 })
const showPreviewModal = ref(false)
const previewFileData = ref(null)
const selectedAttachmentTab = ref('0')

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

function isOverdue(deadline) {
  return new Date(deadline) < new Date()
}

function formatFileSize(bytes) {
  if (!bytes) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}

const currentAttachments = computed(() => normalizeAssignmentAttachments(currentAssignment.value?.attachments, { keepEmpty: false }))

const historyColumns = [
  {
    title: '作业',
    key: 'assignment_id',
    render: (row) => {
      const assignment = availableAssignments.value.find(item => item.id === row.assignment_id)
      return assignment?.title || `作业 #${row.assignment_id}`
    }
  },
  {
    title: '文件名',
    key: 'file_name',
    ellipsis: { tooltip: true },
    render: (row) => row.file_name || (row.content ? '文本内容' : '-')
  },
  { title: '提交时间', key: 'submitted_at', render: (row) => formatDate(row.submitted_at) },
  {
    title: '状态',
    key: 'status',
    render: (row) => {
      const statusMap = {
        submitted: { type: 'warning', text: '待审核' },
        reviewed: { type: 'info', text: '已审核' },
        completed: { type: 'success', text: '已完成' }
      }
      const current = statusMap[row.status] || { type: 'default', text: row.status }
      return h(NTag, { type: current.type }, () => current.text)
    }
  },
  { title: '评分', key: 'grade', render: (row) => row.grade !== null ? `${row.grade} 分` : '-' },
  {
    title: '操作',
    key: 'actions',
    render: (row) => {
      if (row.file_path || row.file_url) {
        return h(NButton, { size: 'small', onClick: () => previewFile(row) }, () => '预览')
      }
      return h(NButton, { size: 'small', disabled: true }, () => '无附件')
    }
  }
]

async function loadAssignments() {
  try {
    const res = await assignmentApi.getActiveList()
    availableAssignments.value = res || []
  } catch (err) {
    console.error(err)
  }
}

async function loadMySubmissions() {
  try {
    const res = await assignmentApi.getMySubmissions({})
    mySubmissions.value = res || []
    userSubmissionMap.value = {}
    for (const item of res || []) {
      userSubmissionMap.value[item.assignment_id] = item
    }
  } catch (err) {
    console.error(err)
  }
}

function viewDetail(assignment) {
  currentAssignment.value = assignment
  userSubmission.value = userSubmissionMap.value[assignment.id] || null
  allowResubmit.value = !!userSubmission.value
  uploadResult.value = null
  uploadFile.value = null
  selectedAttachmentTab.value = '0'
  showSubmitModal.value = true
}

async function deleteAndResubmit() {
  if (!userSubmission.value) return

  try {
    await assignmentApi.deleteSubmission(userSubmission.value.id)
    message.success('已删除原提交，现在可以重新提交')
    userSubmission.value = null
    allowResubmit.value = false
    loadMySubmissions()
    loadAssignments()
  } catch (err) {
    console.error(err)
    message.error('删除失败，请重试')
  }
}

function beforeUpload({ file }) {
  const validTypes = ['.doc', '.docx', '.pdf', '.ppt', '.pptx', '.txt']
  const fileExt = `.${file.name.split('.').pop().toLowerCase()}`
  if (!validTypes.includes(fileExt)) {
    message.error('请上传 .doc、.docx、.pdf、.ppt、.pptx、.txt 文件')
    return false
  }
  return true
}

function customRequest({ file, onFinish, onError }) {
  uploadFile.value = {
    name: file.name,
    size: file.file?.size || 0,
    file: file.file
  }

  assignmentApi.uploadFile(file.file).then(res => {
    uploadResult.value = res
    message.success('文件上传成功')
    onFinish()
  }).catch(err => {
    console.error(err)
    uploadFile.value = null
    message.error(err.response?.data?.detail || '文件上传失败')
    onError()
  })
}

function removeFile() {
  uploadFile.value = null
  uploadResult.value = null
}

function viewUploadedFile() {
  if (!uploadResult.value) return
  previewFile(uploadResult.value)
}

async function handleSubmit() {
  if (!uploadResult.value) {
    message.warning('请先上传文件')
    return
  }

  submitting.value = true
  try {
    await assignmentApi.submit({
      assignment_id: currentAssignment.value.id,
      file_path: uploadResult.value.file_path,
      file_name: uploadResult.value.file_name,
      file_size: uploadResult.value.file_size
    })
    message.success('提交成功')
    showSubmitModal.value = false
    loadMySubmissions()
    loadAssignments()
  } catch (err) {
    console.error(err)
    message.error(err.response?.data?.detail || '提交失败')
  } finally {
    submitting.value = false
  }
}

function previewFile(file) {
  previewFileData.value = {
    file_name: file.file_name,
    file_path: file.file_path,
    file_url: file.file_url
  }
  showPreviewModal.value = true
}

onMounted(() => {
  loadAssignments()
  loadMySubmissions()
})
</script>

<style scoped>
.client-assignments {
  padding-bottom: 24px;
}

.page-header {
  margin-bottom: 16px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.assignment-list {
  padding: 16px 0;
}

.assignment-card {
  transition: all 0.2s;
}

.assignment-card.overdue {
  opacity: 0.8;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header .title {
  font-weight: 600;
}

.card-body .description {
  color: #666;
  margin-bottom: 12px;
  line-height: 1.6;
}

.card-body .deadline {
  font-size: 14px;
  color: #666;
  margin-bottom: 12px;
}

.card-body .deadline .text-error {
  color: #d03050;
}

.card-body .submission-status {
  display: flex;
  align-items: center;
  gap: 12px;
}

.submission-status-placeholder {
  min-height: 32px;
  display: flex;
  align-items: center;
}

.submission-status-empty {
  display: flex;
  align-items: center;
}

.submit-time {
  font-size: 13px;
  color: #999;
}

.submit-form {
  padding: 2px 0;
}

.detail-top {
  display: grid;
  grid-template-columns: minmax(300px, 0.85fr) minmax(430px, 1.15fr);
  gap: 8px;
}

.detail-panel {
  min-width: 0;
  padding: 4px 8px;
  border: 1px solid #e0e0e6;
  border-radius: 8px;
  background: #fff;
}

.section-title {
  margin-bottom: 3px;
  font-weight: 600;
  color: #333;
  font-size: 13px;
  line-height: 18px;
}

.assignment-summary,
.submission-summary {
  display: flex;
  align-items: center;
  min-width: 0;
  height: 26px;
  gap: 8px;
  color: #666;
  font-size: 13px;
}

.summary-description,
.submission-file {
  overflow: hidden;
  min-width: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.summary-description {
  flex: 1;
}

.summary-divider {
  color: #ddd;
}

.summary-deadline {
  flex-shrink: 0;
  color: #888;
}

.attachment-panel {
  min-height: 596px;
}

.attachment-tabs {
  height: 560px;
}

.attachment-tabs :deep(.n-tabs-nav-scroll-wrapper) {
  width: 176px;
  padding-right: 12px;
  border-right: 1px solid #e0e0e6;
}

.attachment-tabs :deep(.n-tabs-pane-wrapper) {
  padding-left: 24px;
}

.attachment-tab {
  position: relative;
  box-sizing: border-box;
  width: 100px;
  padding: 5px 4px 5px;
  text-align: left;
  white-space: normal;
}

.attachment-tabs :deep(.n-tabs-tab--active .attachment-tab)::after {
  position: absolute;
  right: 4px;
  bottom: 2px;
  left: 4px;
  height: 2px;
  border-radius: 2px;
  background: #18a058;
  content: '';
}

.attachment-name {
  color: #333;
  font-weight: 600;
  margin-bottom: 2px;
}

.attachment-file,
.attachment-size {
  color: #666;
  font-size: 12px;
  word-break: break-all;
}

.attachment-preview {
  height: 560px;
}

.attachment-empty {
  padding: 32px 0 16px;
}

.upload-section {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 0;
}

.upload-area {
  flex: 1;
}

.upload-dragger {
  display: flex;
  align-items: center;
  gap: 8px;
  border: 2px dashed #d0d0d5;
  border-radius: 8px;
  padding: 3px 10px;
  cursor: pointer;
  transition: all 0.3s;
}

.upload-dragger:hover {
  border-color: #18a058;
  background: #f0f9f4;
}

.upload-text {
  font-size: 14px;
  color: #333;
}

.upload-hint {
  font-size: 12px;
  color: #999;
}

.selected-file {
  flex: 1;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-name {
  font-weight: 500;
  color: #333;
}

.file-size {
  font-size: 12px;
  color: #999;
  margin-left: auto;
}

.submission-actions {
  flex-shrink: 0;
}

.current-submission-resubmit {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 0;
}

@media (max-width: 960px) {
  .detail-top {
    grid-template-columns: 1fr;
  }

  .attachment-tabs :deep(.n-tabs-nav-scroll-wrapper) {
    width: 154px;
    padding-right: 8px;
  }

  .attachment-tab {
    width: 130px;
  }
}
</style>
