<template>
  <div class="assignments">
    <n-card title="作业管理">
      <template #header-extra>
        <n-space>
          <n-select
            v-model:value="filterStatus"
            :options="statusOptions"
            placeholder="筛选状态"
            clearable
            style="width: 120px"
            @update:value="handleStatusChange"
          />
          <n-button type="primary" @click="handleAddClick">发布作业</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="assignments"
        :pagination="assignmentPagination"
        remote
        :bordered="false"
      />
    </n-card>

    <n-modal
      v-model:show="showAddModal"
      preset="dialog"
      :title="isEdit ? '编辑作业' : '发布作业'"
      style="width: 760px;"
    >
      <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="88">
        <n-form-item label="标题" path="title">
          <n-input v-model:value="formData.title" placeholder="作业标题" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formData.description" type="textarea" placeholder="作业描述" :rows="3" />
        </n-form-item>
        <n-form-item label="截止时间" path="deadline">
          <n-date-picker
            v-model:value="formData.deadline"
            type="datetime"
            style="width: 100%"
            actions="confirm"
          />
        </n-form-item>
        <n-form-item label="可见角色" path="visible_roles">
          <n-select
            v-model:value="formData.visible_roles"
            :options="roleOptions"
            placeholder="留空表示所有角色可见"
            multiple
            clearable
          />
        </n-form-item>
        <n-form-item label="状态" path="status">
          <n-select v-model:value="formData.status" :options="statusOptions.filter(item => item.value)" />
        </n-form-item>
        <n-form-item label="附件列表">
          <div class="attachment-editor">
            <div class="attachment-toolbar">
              <span class="attachment-tip">默认包含实验指导、实验报告模板、学习视频，可编辑名称、删除或继续新增。</span>
              <n-button size="small" @click="handleAddAttachment">新增附件项</n-button>
            </div>
            <div v-if="formData.attachments.length" class="attachment-list">
              <n-card
                v-for="(attachment, index) in formData.attachments"
                :key="index"
                size="small"
                class="attachment-card"
              >
                <div class="attachment-card-header">
                  <n-input
                    v-model:value="attachment.name"
                    placeholder="附件名称"
                    maxlength="100"
                  />
                  <n-button quaternary type="error" @click="handleRemoveAttachment(index)">删除</n-button>
                </div>
                <div v-if="attachment.file_name" class="attachment-file-row">
                  <div class="attachment-file-meta">
                    <span class="attachment-file-name">{{ attachment.file_name }}</span>
                    <span class="attachment-file-size">{{ formatAttachmentSize(attachment.file_size) }}</span>
                  </div>
                  <n-space>
                    <n-button size="small" @click="previewAttachment(attachment)">预览</n-button>
                    <n-button size="small" @click="clearAttachmentFile(index)">移除文件</n-button>
                  </n-space>
                </div>
                <n-upload
                  :max="1"
                  :show-file-list="false"
                  :accept="ASSIGNMENT_ATTACHMENT_ACCEPT"
                  :custom-request="createAttachmentUploadHandler(index)"
                  @before-upload="beforeAttachmentUpload"
                >
                  <n-button dashed block>上传文件</n-button>
                </n-upload>
                <div class="attachment-hint">支持视频、PDF、Word、文本、Markdown。每项仅 1 个文件。</div>
              </n-card>
            </div>
          </div>
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showAddModal = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleSubmit">确定</n-button>
      </template>
    </n-modal>

    <n-modal v-model:show="showSubmissionsModal" preset="card" title="提交列表" style="width: 900px; max-width: 90vw;">
      <template #header-extra>
        <n-button size="small" @click="exportSubmissions">导出 CSV</n-button>
      </template>
      <n-data-table
        :columns="submissionColumns"
        :data="submissions"
        :pagination="submissionPagination"
        remote
        :bordered="false"
      />
    </n-modal>

    <FilePreviewModal
      v-model:show="showPreviewModal"
      :file-data="previewFileData"
    />

    <n-modal v-model:show="gradeModalVisible" preset="dialog" title="评分" style="width: 400px;">
      <n-form :model="gradeData" label-placement="left" label-width="60">
        <n-form-item label="分数">
          <n-input-number v-model:value="gradeData.grade" :min="0" :max="100" style="width: 100%" />
        </n-form-item>
        <n-form-item label="反馈">
          <n-input v-model:value="gradeData.feedback" type="textarea" :rows="3" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="gradeModalVisible = false">取消</n-button>
        <n-button type="primary" @click="submitGrade">确定</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import {
  NCard,
  NButton,
  NInput,
  NDataTable,
  NSpace,
  NModal,
  NForm,
  NFormItem,
  NSelect,
  NDatePicker,
  useMessage,
  NTag,
  NInputNumber,
  NUpload
} from 'naive-ui'
import { assignmentApi, roleApi } from '@/api'
import FilePreviewModal from '@/components/FilePreviewModal.vue'
import {
  ASSIGNMENT_ATTACHMENT_ACCEPT,
  createDefaultAssignmentAttachments,
  normalizeAssignmentAttachments,
  isAssignmentAttachmentFileValid,
  formatAttachmentSize
} from '@/utils/assignmentAttachments'

const message = useMessage()
const formRef = ref(null)
const assignments = ref([])
const submissions = ref([])
const roleOptions = ref([])
const showAddModal = ref(false)
const showSubmissionsModal = ref(false)
const showPreviewModal = ref(false)
const previewFileData = ref(null)
const isEdit = ref(false)
const loading = ref(false)
const filterStatus = ref(null)
const currentAssignmentId = ref(null)
const gradeModalVisible = ref(false)
const gradeData = ref({ id: null, grade: null, feedback: '' })

const assignmentPagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    assignmentPagination.value.page = page
    loadAssignments()
  },
  onUpdatePageSize: (pageSize) => {
    assignmentPagination.value.pageSize = pageSize
    assignmentPagination.value.page = 1
    loadAssignments()
  }
})

const submissionPagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    submissionPagination.value.page = page
    loadSubmissions()
  },
  onUpdatePageSize: (pageSize) => {
    submissionPagination.value.pageSize = pageSize
    submissionPagination.value.page = 1
    loadSubmissions()
  }
})

const formData = ref(createEmptyForm())

const rules = {
  title: [{ required: true, message: '请输入标题' }],
  deadline: [{ required: true, message: '请选择截止时间' }]
}

const statusOptions = [
  { label: '全部', value: null },
  { label: '进行中', value: 'active' },
  { label: '已完成', value: 'completed' },
  { label: '已归档', value: 'archived' }
]

function createEmptyForm() {
  return {
    id: '',
    title: '',
    description: '',
    deadline: null,
    visible_roles: null,
    status: 'active',
    attachments: createDefaultAssignmentAttachments()
  }
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

function isOverdue(deadline) {
  return new Date(deadline) < new Date()
}

function getRowNumber(pagination, index) {
  const page = pagination?.page || 1
  const pageSize = pagination?.pageSize || 10
  return (page - 1) * pageSize + index + 1
}

const columns = computed(() => [
  { title: '序号', key: 'index', width: 70, render: (_, index) => getRowNumber(assignmentPagination.value, index) },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  {
    title: '附件数',
    key: 'attachments',
    width: 90,
    render: (row) => normalizeAssignmentAttachments(row.attachments, { keepEmpty: false }).length
  },
  {
    title: '截止时间',
    key: 'deadline',
    width: 180,
    render: (row) => {
      const overdue = isOverdue(row.deadline) && row.status === 'active'
      return h('span', { style: { color: overdue ? 'var(--mixm-error)' : 'inherit' } }, formatDate(row.deadline))
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const statusMap = {
        active: { type: 'success', text: '进行中' },
        completed: { type: 'info', text: '已完成' },
        archived: { type: 'default', text: '已归档' }
      }
      const current = statusMap[row.status] || { type: 'default', text: row.status }
      return h(NTag, { type: current.type }, () => current.text)
    }
  },
  { title: '提交数', key: 'submission_count', width: 80 },
  {
    title: '角色',
    key: 'visible_roles',
    width: 200,
    render: (row) => {
      if (!row.visible_roles?.length) {
        return h(NTag, { size: 'small', type: 'default' }, () => '全部角色')
      }

      const roleMap = new Map(roleOptions.value.map(role => [role.value, role.label]))
      return h(
        NSpace,
        { size: 4 },
        () => row.visible_roles.map(roleId => h(
          NTag,
          { key: roleId, size: 'small', type: 'info' },
          () => roleMap.get(roleId) || roleId
        ))
      )
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 260,
    render: (row) => h(NSpace, {}, () => [
      h(NButton, { size: 'small', onClick: () => viewSubmissions(row) }, () => '查看提交'),
      h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => '编辑'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
])

const submissionColumns = computed(() => [
  { title: '用户', key: 'username', width: 120 },
  { title: '文件名', key: 'file_name', ellipsis: { tooltip: true } },
  { title: '提交时间', key: 'submitted_at', width: 180, render: (row) => formatDate(row.submitted_at) },
  {
    title: '状态',
    key: 'status',
    width: 100,
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
  {
    title: '评分',
    key: 'grade',
    width: 80,
    render: (row) => row.grade != null
      ? h('span', { style: { color: row.grade >= 60 ? 'var(--mixm-success)' : 'var(--mixm-error)', fontWeight: 'bold' } }, row.grade)
      : '-'
  },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render: (row) => h(NSpace, {}, () => [
      row.file_path ? h(NButton, { size: 'small', onClick: () => previewFile(row) }, () => '预览') : null,
      h(NButton, { size: 'small', onClick: () => handleGrade(row) }, () => '评分'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDeleteSubmission(row.id) }, () => '删除')
    ])
  }
])

async function loadRoleOptions() {
  try {
    const res = await roleApi.getList({ offset: 0, limit: 100 })
    roleOptions.value = (res.items || []).map(item => ({
      label: item.name,
      value: item.id
    }))
  } catch (err) {
    console.error(err)
  }
}

async function loadAssignments() {
  try {
    const page = assignmentPagination.value.page || 1
    const pageSize = assignmentPagination.value.pageSize || 10
    const params = {
      offset: (page - 1) * pageSize,
      limit: pageSize
    }
    if (filterStatus.value) params.status = filterStatus.value

    const res = await assignmentApi.getList(params)
    assignments.value = res.items || []
    assignmentPagination.value.itemCount = res.total || 0
  } catch (err) {
    console.error(err)
  }
}

function handleStatusChange() {
  assignmentPagination.value.page = 1
  loadAssignments()
}

async function loadSubmissions(assignmentId = currentAssignmentId.value) {
  if (!assignmentId) return

  const page = submissionPagination.value.page || 1
  const pageSize = submissionPagination.value.pageSize || 10

  try {
    const res = await assignmentApi.getSubmissions(assignmentId, {
      offset: (page - 1) * pageSize,
      limit: pageSize
    })
    submissions.value = res.items || []
    submissionPagination.value.itemCount = res.total || 0
  } catch (err) {
    console.error(err)
    message.error('加载提交列表失败')
  }
}

function handleAddClick() {
  isEdit.value = false
  formData.value = createEmptyForm()
  showAddModal.value = true
}

function handleEdit(row) {
  isEdit.value = true
  formData.value = {
    id: row.id,
    title: row.title || '',
    description: row.description || '',
    deadline: row.deadline ? new Date(row.deadline).getTime() : null,
    visible_roles: row.visible_roles?.length ? [...row.visible_roles] : null,
    status: row.status || 'active',
    attachments: normalizeAssignmentAttachments(
      row.attachments?.length ? row.attachments : createDefaultAssignmentAttachments()
    )
  }
  showAddModal.value = true
}

function buildAssignmentPayload() {
  return {
    ...formData.value,
    deadline: new Date(formData.value.deadline).toISOString(),
    visible_roles: formData.value.visible_roles?.length ? formData.value.visible_roles : null,
    attachments: normalizeAssignmentAttachments(formData.value.attachments)
  }
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  if (formData.value.attachments.some(item => !item.name.trim())) {
    message.warning('附件名称不能为空')
    return
  }

  loading.value = true
  try {
    const payload = buildAssignmentPayload()
    if (isEdit.value) {
      await assignmentApi.update(formData.value.id, payload)
      message.success('更新成功')
    } else {
      await assignmentApi.create(payload)
      message.success('发布成功')
    }
    showAddModal.value = false
    loadAssignments()
  } catch (err) {
    console.error(err)
    message.error('操作失败')
  } finally {
    loading.value = false
  }
}

async function handleDelete(id) {
  try {
    await assignmentApi.delete(id)
    message.success('删除成功')
    loadAssignments()
  } catch (err) {
    console.error(err)
    message.error('删除失败')
  }
}

async function viewSubmissions(row) {
  currentAssignmentId.value = row.id
  submissionPagination.value.page = 1
  showSubmissionsModal.value = true
  await loadSubmissions(row.id)
}

async function exportSubmissions() {
  if (!currentAssignmentId.value) return

  try {
    const res = await assignmentApi.exportSubmissions(currentAssignmentId.value)
    const fileUrl = `${import.meta.env.VITE_API_BASE_URL.replace('/api', '')}${res.file_url}`
    const link = document.createElement('a')
    link.href = fileUrl
    link.download = res.filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    message.success('导出成功')
  } catch (err) {
    console.error(err)
    message.error('导出失败')
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

function previewAttachment(attachment) {
  if (!attachment.file_url && !attachment.file_path) {
    message.warning('当前附件还未上传文件')
    return
  }
  previewFile(attachment)
}

function handleAddAttachment() {
  formData.value.attachments.push({
    name: '',
    file_path: '',
    file_name: '',
    file_size: null,
    file_url: ''
  })
}

function handleRemoveAttachment(index) {
  formData.value.attachments.splice(index, 1)
}

function clearAttachmentFile(index) {
  formData.value.attachments[index] = {
    ...formData.value.attachments[index],
    file_path: '',
    file_name: '',
    file_size: null,
    file_url: ''
  }
}

function beforeAttachmentUpload({ file }) {
  if (!isAssignmentAttachmentFileValid(file.name)) {
    message.error('仅支持视频、PDF、Word、文本和 Markdown 文件')
    return false
  }
  return true
}

function createAttachmentUploadHandler(index) {
  return async ({ file, onFinish, onError }) => {
    try {
      const res = await assignmentApi.uploadFile(file.file)
      formData.value.attachments[index] = {
        ...formData.value.attachments[index],
        file_path: res.file_path,
        file_name: res.file_name,
        file_size: res.file_size,
        file_url: res.file_url
      }
      message.success('附件上传成功')
      onFinish()
    } catch (err) {
      console.error(err)
      message.error('附件上传失败')
      onError()
    }
  }
}

function handleGrade(row) {
  gradeData.value = { id: row.id, grade: row.grade ?? null, feedback: row.feedback || '' }
  gradeModalVisible.value = true
}

async function submitGrade() {
  try {
    await assignmentApi.updateSubmission(gradeData.value.id, {
      grade: gradeData.value.grade,
      feedback: gradeData.value.feedback,
      status: gradeData.value.grade !== null ? 'reviewed' : 'submitted'
    })
    message.success('评分成功')
    gradeModalVisible.value = false
    await loadSubmissions(currentAssignmentId.value)
  } catch (err) {
    console.error(err)
    message.error('评分失败')
  }
}

async function handleDeleteSubmission(id) {
  try {
    await assignmentApi.deleteSubmission(id)
    message.success('删除成功')
    await loadSubmissions(currentAssignmentId.value)
  } catch (err) {
    console.error(err)
    message.error('删除失败')
  }
}

onMounted(() => {
  loadRoleOptions()
  loadAssignments()
})
</script>

<style scoped>
.assignments {
  padding: 20px 0;
}

.attachment-editor {
  width: 100%;
}

.attachment-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.attachment-tip {
  color: var(--mixm-muted);
  font-size: 12px;
}

.attachment-list {
  display: grid;
  gap: 12px;
}

.attachment-card {
  background: var(--mixm-paper);
}

.attachment-card-header {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 12px;
  margin-bottom: 12px;
}

.attachment-file-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.attachment-file-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.attachment-file-name {
  color: var(--mixm-ink);
  font-weight: 500;
  word-break: break-all;
}

.attachment-file-size,
.attachment-hint {
  color: var(--mixm-faint);
  font-size: 12px;
}
</style>
