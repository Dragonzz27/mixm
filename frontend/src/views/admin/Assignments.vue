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
            @update:value="loadAssignments"
          />
          <n-button type="primary" @click="handleAddClick">发布作业</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="assignments"
        :pagination="pagination"
        :bordered="false"
      />
    </n-card>

    <n-modal v-model:show="showAddModal" preset="dialog" :title="isEdit ? '编辑作业' : '发布作业'" style="width: 600px;">
      <n-form :model="formData" :rules="rules" ref="formRef" label-placement="left" label-width="80">
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
        <n-form-item label="状态" path="status">
          <n-select v-model:value="formData.status" :options="statusOptions.filter(s => s.value)" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showAddModal = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleSubmit">确定</n-button>
      </template>
    </n-modal>

    <n-modal v-model:show="showSubmissionsModal" preset="card" title="提交列表" style="width: 900px; max-width: 90vw;">
      <n-data-table
        :columns="submissionColumns"
        :data="submissions"
        :pagination="pagination"
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
import { NCard, NButton, NInput, NDataTable, NSpace, NModal, NForm, NFormItem, NSelect, NDatePicker, useMessage, NTag, NInputNumber } from 'naive-ui'
import { assignmentApi } from '@/api'
import FilePreviewModal from '@/components/FilePreviewModal.vue'

const message = useMessage()
const formRef = ref(null)
const assignments = ref([])
const submissions = ref([])
const showAddModal = ref(false)
const showSubmissionsModal = ref(false)
const showPreviewModal = ref(false)
const previewFileData = ref(null)
const isEdit = ref(false)
const loading = ref(false)
const pagination = ref({ pageSize: 10 })
const filterStatus = ref(null)
const currentAssignmentId = ref(null)

const formData = ref({
  title: '',
  description: '',
  deadline: null,
  status: 'active'
})

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

function formatDate(dateStr) {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

function isOverdue(deadline) {
  return new Date(deadline) < new Date()
}

const columns = computed(() => [
  { title: 'ID', key: 'id', width: 60 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { 
    title: '截止时间', 
    key: 'deadline', 
    width: 180,
    render: (row) => {
      const overdue = isOverdue(row.deadline) && row.status === 'active'
      return h('span', { style: { color: overdue ? '#d03050' : 'inherit' } }, formatDate(row.deadline))
    }
  },
  { 
    title: '状态', 
    key: 'status', 
    width: 100,
    render: (row) => {
      const statusMap = { active: { type: 'success', text: '进行中' }, completed: { type: 'info', text: '已完成' }, archived: { type: 'default', text: '已归档' } }
      const s = statusMap[row.status] || { type: 'default', text: row.status }
      return h(NTag, { type: s.type }, () => s.text)
    }
  },
  { title: '提交数', key: 'submission_count', width: 80 },
  { title: '创建时间', key: 'created_at', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 400,
    render: (row) => h(NSpace, {}, () => [
      h(NButton, { size: 'small', onClick: () => viewSubmissions(row) }, () => '查看提交'),
      h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => '编辑'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
])

const submissionColumns = computed(() => [
  { title: 'ID', key: 'id', width: 60 },
  { title: '用户', key: 'username', width: 120 },
  { title: '文件名', key: 'file_name', ellipsis: { tooltip: true } },
  { title: '提交时间', key: 'submitted_at', width: 180, render: (row) => formatDate(row.submitted_at) },
  { 
    title: '状态', 
    key: 'status', 
    width: 100,
    render: (row) => {
      const statusMap = { submitted: { type: 'warning', text: '待审核' }, reviewed: { type: 'info', text: '已审核' }, completed: { type: 'success', text: '已完成' } }
      const s = statusMap[row.status] || { type: 'default', text: row.status }
      return h(NTag, { type: s.type }, () => s.text)
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 300,
    render: (row) => h(NSpace, {}, () => [
      row.file_path ? h(NButton, { size: 'small', onClick: () => previewFile(row) }, () => '预览') : null,
      h(NButton, { size: 'small', onClick: () => handleGrade(row) }, () => '评分'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDeleteSubmission(row.id) }, () => '删除')
    ])
  }
])

async function loadAssignments() {
  try {
    const params = { skip: 0, limit: 100 }
    if (filterStatus.value) params.status = filterStatus.value
    const res = await assignmentApi.getList(params)
    assignments.value = res.assignments || []
  } catch (err) {
    console.error(err)
  }
}

function handleAddClick() {
  isEdit.value = false
  formData.value = { title: '', description: '', deadline: null, status: 'active' }
  showAddModal.value = true
}

async function handleEdit(row) {
  isEdit.value = true
  formData.value = {
    ...row,
    deadline: row.deadline ? new Date(row.deadline).getTime() : null
  }
  showAddModal.value = true
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  
  loading.value = true
  try {
    const data = {
      ...formData.value,
      deadline: new Date(formData.value.deadline).toISOString()
    }
    if (isEdit.value) {
      await assignmentApi.update(formData.value.id, data)
      message.success('更新成功')
    } else {
      await assignmentApi.create(data)
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
  try {
    const res = await assignmentApi.getSubmissions(row.id, { skip: 0, limit: 100 })
    submissions.value = res.submissions || []
    showSubmissionsModal.value = true
  } catch (err) {
    console.error(err)
    message.error('加载提交列表失败')
  }
}

function previewFile(row) {
  const fileData = {
    file_name: row.file_name,
    file_path: row.file_path,
    file_url: row.file_url
  }
  previewFileData.value = fileData
  showPreviewModal.value = true
}

const gradeModalVisible = ref(false)
const gradeData = ref({ id: null, grade: null, feedback: '' })

function handleGrade(row) {
  gradeData.value = { id: row.id, grade: row.grade || null, feedback: row.feedback || '' }
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
    viewSubmissions({ id: currentAssignmentId.value })
  } catch (err) {
    console.error(err)
    message.error('评分失败')
  }
}

async function handleDeleteSubmission(id) {
  try {
    await assignmentApi.deleteSubmission(id)
    message.success('删除成功')
    viewSubmissions({ id: currentAssignmentId.value })
  } catch (err) {
    console.error(err)
    message.error('删除失败')
  }
}

onMounted(() => {
  loadAssignments()
})
</script>

<style scoped>
.assignments {
  padding: 20px 0;
}
</style>