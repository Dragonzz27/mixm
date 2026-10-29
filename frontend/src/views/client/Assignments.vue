<template>
  <div class="client-assignments">
    <n-tabs type="line" animated>
      <n-tab-pane name="available" tab="可提交作业">
        <div class="assignment-list">
          <n-grid v-if="availableAssignments.length > 0" :cols="2" :x-gap="20" :y-gap="20">
            <n-gi v-for="assignment in availableAssignments" :key="assignment.id">
              <n-card class="assignment-card" :class="{ overdue: isOverdue(assignment.deadline) }">
                <template #header>
                  <div class="card-header">
                    <span class="title">{{ assignment.title }}</span>
                    <n-tag v-if="isOverdue(assignment.deadline)" type="error" size="small">
                      已过期
                    </n-tag>
                  </div>
                </template>
                <div class="card-body">
                  <p class="description">{{ assignment.description || '暂无描述' }}</p>
                  <div class="deadline">
                    <span class="label">截止时间：</span>
                    <span :class="{ 'text-error': isOverdue(assignment.deadline) }">
                      {{ formatDate(assignment.deadline) }}
                    </span>
                  </div>
                  <div class="submission-status-placeholder">
                    <div v-if="userSubmissionMap[assignment.id]" class="submission-status">
                      <n-tag type="success">已提交</n-tag>
                      <span class="submit-time">
                        提交于：{{ formatDate(userSubmissionMap[assignment.id].submitted_at) }}
                      </span>
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
                      :disabled="
                        isOverdue(assignment.deadline) && !userSubmissionMap[assignment.id]
                      "
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

    <FilePreviewModal v-model:show="showPreviewModal" :file-data="previewFileData" />

    <AssignmentDetailModal
      v-model:show="showSubmitModal"
      :assignment="currentAssignment"
      :submission="currentSubmission"
      @change="refreshAssignments"
    />
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
  NDataTable
} from 'naive-ui'
import FilePreviewModal from '@/components/FilePreviewModal.vue'
import AssignmentDetailModal from '@/components/AssignmentDetailModal.vue'
import { assignmentApi } from '@/api'

const availableAssignments = ref([])
const mySubmissions = ref([])
const userSubmissionMap = ref({})
const showSubmitModal = ref(false)
const currentAssignment = ref(null)
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

const currentSubmission = computed(
  () => userSubmissionMap.value[currentAssignment.value?.id] || null
)

const historyColumns = [
  {
    title: '作业',
    key: 'assignment_id',
    render: (row) => {
      const assignment = availableAssignments.value.find((item) => item.id === row.assignment_id)
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
  { title: '评分', key: 'grade', render: (row) => (row.grade !== null ? `${row.grade} 分` : '-') },
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
    availableAssignments.value = res.items || []
  } catch (err) {
    console.error(err)
  }
}

async function loadMySubmissions() {
  try {
    const res = await assignmentApi.getMySubmissions({})
    mySubmissions.value = res.items || []
    userSubmissionMap.value = {}
    for (const item of res.items || []) {
      userSubmissionMap.value[item.assignment_id] = item
    }
  } catch (err) {
    console.error(err)
  }
}

function viewDetail(assignment) {
  currentAssignment.value = assignment
  showSubmitModal.value = true
}

function refreshAssignments() {
  loadMySubmissions()
  loadAssignments()
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
  color: var(--mixm-muted);
  margin-bottom: 12px;
  line-height: 1.6;
}

.card-body .deadline {
  font-size: 14px;
  color: var(--mixm-muted);
  margin-bottom: 12px;
}

.card-body .deadline .text-error {
  color: var(--mixm-error);
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
  color: var(--mixm-faint);
}
</style>
