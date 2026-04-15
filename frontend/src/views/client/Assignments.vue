<template>
  <div class="client-assignments">
    <div class="page-header">
      <h1>我的作业</h1>
    </div>

    <n-tabs type="line" animated>
      <n-tab-pane name="available" tab="可提交作业">
        <div class="assignment-list">
          <n-grid :cols="2" :x-gap="20" :y-gap="20" v-if="availableAssignments.length > 0">
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
                    <n-button type="primary" @click="viewDetail(assignment)" :disabled="isOverdue(assignment.deadline) && !userSubmissionMap[assignment.id]">
                      {{ userSubmissionMap[assignment.id] ? '查看提交' : '提交作业' }}
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

    <n-modal v-model:show="showSubmitModal" preset="card" :title="currentAssignment?.title || '提交作业'" style="width: 600px;">
      <div v-if="currentAssignment" class="submit-form">
        <n-descriptions :column="1" label-placement="left" bordered>
          <n-descriptions-item label="作业描述">{{ currentAssignment.description || '暂无描述' }}</n-descriptions-item>
          <n-descriptions-item label="截止时间">{{ formatDate(currentAssignment.deadline) }}</n-descriptions-item>
        </n-descriptions>

        <n-divider />

        <div v-if="userSubmission && allowResubmit" class="current-submission-resubmit">
          <n-alert type="info" title="您已提交过此作业">
            <template #default>
              <div><strong>文件名：</strong>{{ userSubmission.file_name || '文本内容' }}</div>
              <div><strong>提交时间：</strong>{{ formatDate(userSubmission.submitted_at) }}</div>
              <div v-if="userSubmission.grade !== null"><strong>评分：</strong>{{ userSubmission.grade }}分</div>
              <div v-if="userSubmission.feedback"><strong>反馈：</strong>{{ userSubmission.feedback }}</div>
            </template>
          </n-alert>
          
          <n-space style="margin-top: 16px;">
            <n-button v-if="userSubmission.file_path || userSubmission.file_url" @click="previewFile(userSubmission)">预览文件</n-button>
            <n-button type="warning" @click="deleteAndResubmit">删除并重新提交</n-button>
          </n-space>
        </div>

        <div v-else class="upload-section">
          <div v-if="uploadFile" class="selected-file">
            <n-card size="small">
              <div class="file-info">
                <n-icon :size="20"><DocumentTextOutline /></n-icon>
                <span class="file-name">{{ uploadFile.name }}</span>
                <span class="file-size">{{ formatFileSize(uploadFile.size) }}</span>
              </div>
              <n-space>
                <n-button size="small" @click="viewUploadedFile">预览</n-button>
                <n-button size="small" type="error" @click="removeFile">删除</n-button>
              </n-space>
            </n-card>
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
                <n-icon :size="40"><CloudUploadOutline /></n-icon>
                <div class="upload-text">点击或拖拽文件到此处上传</div>
                <div class="upload-hint">支持 .doc, .docx, .pdf, .ppt, .pptx, .txt 格式，限 1 个文件</div>
              </div>
            </n-upload>
          </div>

          <n-space justify="end" style="margin-top: 16px;">
            <n-button @click="showSubmitModal = false">取消</n-button>
            <n-button 
              type="primary" 
              :loading="submitting" 
              :disabled="!uploadResult" 
              @click="handleSubmit"
            >
              提交作业
            </n-button>
          </n-space>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { NCard, NButton, NSpace, NTag, NGrid, NGi, NTabs, NTabPane, NEmpty, NDescriptions, NDescriptionsItem, NDivider, NAlert, NUpload, NDataTable, NIcon, useMessage } from 'naive-ui'
import { DocumentTextOutline, CloudUploadOutline } from '@vicons/ionicons5'
import FilePreviewModal from '@/components/FilePreviewModal.vue'
import { assignmentApi } from '@/api'

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

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

function isOverdue(deadline) {
  return new Date(deadline) < new Date()
}

function formatFileSize(bytes) {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
}

const historyColumns = [
  { title: '作业', key: 'assignment_id', render: (row) => {
    const assignment = availableAssignments.value.find(a => a.id === row.assignment_id)
    return assignment?.title || `作业 #${row.assignment_id}`
  }},
  { title: '文件名', key: 'file_name', ellipsis: { tooltip: true }, render: (row) => row.file_name || (row.content ? '文本内容' : '-') },
  { title: '提交时间', key: 'submitted_at', render: (row) => formatDate(row.submitted_at) },
  { 
    title: '状态', 
    key: 'status',
    render: (row) => {
      const statusMap = { submitted: { type: 'warning', text: '待审核' }, reviewed: { type: 'info', text: '已审核' }, completed: { type: 'success', text: '已完成' } }
      const s = statusMap[row.status] || { type: 'default', text: row.status }
      return h(NTag, { type: s.type }, () => s.text)
    }
  },
  { title: '评分', key: 'grade', render: (row) => row.grade !== null ? `${row.grade}分` : '-' },
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
    for (const sub of res || []) {
      userSubmissionMap.value[sub.assignment_id] = sub
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
  const fileExt = '.' + file.name.split('.').pop().toLowerCase()
  if (!validTypes.includes(fileExt)) {
    message.error('不支持的文件格式，请上传 .doc, .docx, .pdf, .ppt, .pptx, .txt 文件')
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
    message.error('文件上传失败')
    onError()
  })
}

function removeFile() {
  uploadFile.value = null
  uploadResult.value = null
}

function viewUploadedFile() {
  if (uploadResult.value?.file_url) {
    window.open(uploadResult.value.file_url, '_blank')
  } else if (uploadFile.value?.file) {
    const url = URL.createObjectURL(uploadFile.value.file)
    window.open(url, '_blank')
  }
}

async function handleSubmit() {
  if (!uploadResult.value) {
    message.warning('请上传文件')
    return
  }

  submitting.value = true
  try {
    const payload = {
      assignment_id: currentAssignment.value.id,
      file_path: uploadResult.value.file_path,
      file_name: uploadResult.value.file_name,
      file_size: uploadResult.value.file_size
    }

    await assignmentApi.submit(payload)
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

function previewFile(submission) {
  const fileData = {
    file_name: submission.file_name,
    file_path: submission.file_path,
    file_url: submission.file_url
  }
  previewFileData.value = fileData
  showPreviewModal.value = true
}

function viewFile(submission) {
  previewFile(submission)
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

.upload-section {
  padding: 8px 0;
}

.upload-dragger {
  border: 2px dashed #d0d0d5;
  border-radius: 8px;
  padding: 32px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
}

.upload-dragger:hover {
  border-color: #18a058;
  background: #f0f9f4;
}

.upload-text {
  margin-top: 12px;
  font-size: 14px;
  color: #333;
}

.upload-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #999;
}

.selected-file {
  margin-bottom: 16px;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
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

.current-submission-resubmit {
  margin-bottom: 16px;
}

.submit-form {
  padding: 8px 0;
}

.preview-card {
  transition: all 0.2s;
}

.preview-card.overdue {
  opacity: 0.8;
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
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.preview-not-supported {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}
</style>
