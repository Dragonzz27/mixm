<template>
  <div class="practices">
    <n-card title="练习管理">
      <template #header-extra>
        <n-space>
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索练习标题"
            clearable
            style="width: 220px"
            @keyup.enter="loadPractices"
          />
          <n-button type="primary" @click="handleAddClick">新增练习</n-button>
        </n-space>
      </template>

      <n-data-table
        :columns="columns"
        :data="practices"
        :pagination="pagination"
        remote
        :bordered="false"
      />
    </n-card>

    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑练习' : '新增练习'"
      style="width: 1100px; max-width: 96vw;"
    >
      <n-form ref="formRef" :model="formData" :rules="rules" label-placement="top">
        <n-grid :cols="2" :x-gap="16">
          <n-form-item-gi label="练习标题" path="title">
            <n-input v-model:value="formData.title" placeholder="请输入练习标题" />
          </n-form-item-gi>
          <n-form-item-gi label="可见角色">
            <n-select
              v-model:value="formData.visible_roles"
              :options="roleOptions"
              multiple
              clearable
              placeholder="为空表示全部可见"
            />
          </n-form-item-gi>
        </n-grid>

        <n-form-item label="练习描述">
          <n-input
            v-model:value="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入练习描述"
          />
        </n-form-item>

        <n-form-item label="知识元">
          <n-tree-select
            v-model:value="formData.knowledge_tag_ids"
            :options="knowledgeTreeOptions"
            multiple
            clearable
            filterable
            placeholder="选择用于自动抽题的知识元"
            style="width: 100%"
          />
        </n-form-item>

        <div class="type-counts">
          <div class="section-title">题型分配</div>
          <n-grid :cols="5" :x-gap="12">
            <n-form-item-gi v-for="item in typeItems" :key="item.value" :label="item.label">
              <n-input-number
                v-model:value="formData.type_counts[item.value]"
                :min="0"
                style="width: 100%"
              />
            </n-form-item-gi>
          </n-grid>
          <n-space style="margin-top: 8px">
            <n-button type="primary" :loading="autoGenerating" @click="handleAutoGenerate">
              自动抽题
            </n-button>
            <n-button @click="openQuestionPicker">手动添加题目</n-button>
          </n-space>
        </div>

        <div class="selected-questions">
          <div class="section-title">
            已选题目
            <span class="sub-text">共 {{ formData.selected_questions.length }} 题</span>
          </div>
          <n-data-table
            :columns="selectedQuestionColumns"
            :data="formData.selected_questions"
            :bordered="false"
            size="small"
            :max-height="360"
          />
        </div>
      </n-form>

      <template #action>
        <n-button @click="showModal = false">取消</n-button>
        <n-button type="primary" :loading="saving" @click="handleSubmit">保存</n-button>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showQuestionPicker"
      preset="card"
      title="手动添加题目"
      style="width: 1000px; max-width: 96vw;"
    >
      <n-space style="margin-bottom: 16px" wrap>
        <n-input
          v-model:value="pickerKeyword"
          placeholder="搜索题目"
          clearable
          style="width: 220px"
          @keyup.enter="loadApprovedQuestions"
        />
        <n-select
          v-model:value="pickerType"
          :options="typeFilterOptions"
          clearable
          placeholder="筛选题型"
          style="width: 140px"
          @update:value="loadApprovedQuestions"
        />
        <n-select
          v-model:value="pickerKnowledgeTag"
          :options="knowledgeFlatOptions"
          clearable
          placeholder="筛选知识元"
          style="width: 180px"
          @update:value="loadApprovedQuestions"
        />
        <n-button @click="loadApprovedQuestions">搜索</n-button>
        <n-button type="primary" @click="addSelectedQuestions">加入已选</n-button>
      </n-space>

      <n-data-table
        :columns="questionPickerColumns"
        :data="approvedQuestions"
        :pagination="questionPickerPagination"
        :bordered="false"
        :row-key="(row) => row.id"
        :checked-row-keys="selectedQuestionRowKeys"
        @update:checked-row-keys="(keys) => (selectedQuestionRowKeys = keys)"
      />
    </n-modal>
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItemGi,
  NGrid,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NTag,
  NTreeSelect,
  useMessage,
} from 'naive-ui'
import { categoryApi, practiceApi, questionApi, roleApi } from '@/api'

const message = useMessage()
const formRef = ref(null)

const practices = ref([])
const showModal = ref(false)
const showQuestionPicker = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const autoGenerating = ref(false)
const searchKeyword = ref('')

const roleOptions = ref([])
const roleNameMap = ref({})
const knowledgeTreeData = ref([])
const knowledgeFlatOptions = ref([])

const approvedQuestions = ref([])
const selectedQuestionRowKeys = ref([])
const pickerKeyword = ref('')
const pickerType = ref(null)
const pickerKnowledgeTag = ref(null)

const typeItems = [
  { label: '单选题', value: 'single_choice' },
  { label: '多选题', value: 'multi_choice' },
  { label: '判断题', value: 'true_false' },
  { label: '填空题', value: 'blank' },
  { label: '简答题', value: 'essay' },
]

const typeMap = {
  single_choice: '单选题',
  multi_choice: '多选题',
  true_false: '判断题',
  blank: '填空题',
  essay: '简答题',
}

const typeFilterOptions = [
  { label: '全部', value: null },
  ...typeItems.map((item) => ({ label: item.label, value: item.value })),
]

const rules = {
  title: [{ required: true, message: '请输入练习标题' }],
}

const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    pagination.value.page = page
    loadPractices()
  },
  onUpdatePageSize: (pageSize) => {
    pagination.value.pageSize = pageSize
    pagination.value.page = 1
    loadPractices()
  },
})

const questionPickerPagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    questionPickerPagination.value.page = page
    loadApprovedQuestions()
  },
  onUpdatePageSize: (pageSize) => {
    questionPickerPagination.value.pageSize = pageSize
    questionPickerPagination.value.page = 1
    loadApprovedQuestions()
  },
})

const formData = ref(createEmptyForm())

const knowledgeTreeOptions = computed(() => transformKnowledgeTree(knowledgeTreeData.value))

const columns = computed(() => [
  { title: '序号', key: 'index', width: 80, render: (_, index) => getRowNumber(pagination.value, index) },
  { title: '练习标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '题目数', key: 'question_count', width: 90 },
  {
    title: '可见角色',
    key: 'visible_roles',
    width: 220,
    render: (row) => formatVisibleRoles(row.visible_roles),
  },
  {
    title: '知识元',
    key: 'knowledge_tags',
    width: 220,
    render: (row) => formatKnowledgeTags(row.knowledge_tags),
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render: (row) => formatDate(row.created_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row) =>
      h(NSpace, {}, () => [
        h(NButton, { size: 'small', onClick: () => handleEdit(row.id) }, () => '编辑'),
        h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除'),
      ]),
  },
])

const selectedQuestionColumns = [
  { title: '题目', key: 'title', ellipsis: { tooltip: true } },
  { title: '题型', key: 'type', width: 100, render: (row) => typeMap[row.type] || row.type },
  { title: '难度', key: 'difficulty', width: 80 },
  {
    title: '知识元',
    key: 'knowledge_tags',
    width: 200,
    render: (row) => formatKnowledgeTags(row.knowledge_tags),
  },
  {
    title: '操作',
    key: 'actions',
    width: 90,
    render: (row) =>
      h(NButton, { size: 'small', type: 'error', onClick: () => removeSelectedQuestion(row.id) }, () => '移除'),
  },
]

const questionPickerColumns = [
  { type: 'selection', width: 48 },
  { title: '题目', key: 'title', ellipsis: { tooltip: true } },
  { title: '题型', key: 'type', width: 100, render: (row) => typeMap[row.type] || row.type },
  { title: '难度', key: 'difficulty', width: 80 },
  {
    title: '知识元',
    key: 'knowledge_tags',
    width: 220,
    render: (row) => formatKnowledgeTags(row.knowledge_tags),
  },
]

function createEmptyForm() {
  return {
    id: '',
    title: '',
    description: '',
    visible_roles: null,
    knowledge_tag_ids: [],
    type_counts: {
      single_choice: 0,
      multi_choice: 0,
      true_false: 0,
      blank: 0,
      essay: 0,
    },
    selected_questions: [],
  }
}

function getRowNumber(targetPagination, index) {
  const page = targetPagination?.page || 1
  const pageSize = targetPagination?.pageSize || 10
  return (page - 1) * pageSize + index + 1
}

function formatDate(value) {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

function formatVisibleRoles(roleIds) {
  if (!roleIds || roleIds.length === 0) return '全部可见'
  return roleIds.map((id) => roleNameMap.value[id] || id).join(', ')
}

function formatKnowledgeTags(tags) {
  if (!tags || tags.length === 0) return '-'
  return tags.map((item) => item.name).join(', ')
}

function transformKnowledgeTree(nodes) {
  return (nodes || [])
    .filter((node) => node.name)
    .map((node) => ({
      key: node.id,
      value: node.id,
      label: node.name,
      children: node.children?.length ? transformKnowledgeTree(node.children) : undefined,
    }))
}

function flattenKnowledgeTree(nodes, result = []) {
  for (const node of nodes || []) {
    if (node.id && node.name) {
      result.push({ label: node.name, value: node.id })
    }
    if (node.children?.length) {
      flattenKnowledgeTree(node.children, result)
    }
  }
  return result
}

function normalizeTypeCounts() {
  const output = {}
  for (const item of typeItems) {
    output[item.value] = Number(formData.value.type_counts[item.value] || 0)
  }
  return output
}

function mergeQuestions(existing, incoming) {
  const map = new Map()
  for (const item of existing) {
    map.set(item.id, item)
  }
  for (const item of incoming) {
    if (!map.has(item.id)) {
      map.set(item.id, item)
    }
  }
  return Array.from(map.values())
}

async function loadRoles() {
  try {
    const res = await roleApi.getList({ skip: 0, limit: 200 })
    roleOptions.value = (res.roles || []).map((item) => ({
      label: item.name,
      value: item.id,
    }))
    roleNameMap.value = roleOptions.value.reduce((acc, item) => {
      acc[item.value] = item.label
      return acc
    }, {})
  } catch (err) {
    console.error(err)
  }
}

async function loadKnowledgeTags() {
  try {
    const res = await categoryApi.getTree({ max_depth: 100 })
    knowledgeTreeData.value = res || []
    knowledgeFlatOptions.value = flattenKnowledgeTree(knowledgeTreeData.value)
  } catch (err) {
    console.error(err)
  }
}

async function loadPractices() {
  try {
    const page = pagination.value.page || 1
    const pageSize = pagination.value.pageSize || 10
    const params = {
      skip: (page - 1) * pageSize,
      limit: pageSize,
    }
    if (searchKeyword.value) params.keyword = searchKeyword.value

    const res = await practiceApi.getList(params)
    practices.value = res.practices || []
    pagination.value.itemCount = res.total || 0
  } catch (err) {
    console.error(err)
    message.error('加载练习列表失败')
  }
}

async function loadApprovedQuestions() {
  try {
    const page = questionPickerPagination.value.page || 1
    const pageSize = questionPickerPagination.value.pageSize || 10
    const params = {
      skip: (page - 1) * pageSize,
      limit: pageSize,
      review_status: 'approved',
    }
    if (pickerKeyword.value) params.keyword = pickerKeyword.value
    if (pickerType.value) params.type = pickerType.value
    if (pickerKnowledgeTag.value) params.knowledge_tag_id = pickerKnowledgeTag.value

    const res = await questionApi.getList(params)
    approvedQuestions.value = res.questions || []
    questionPickerPagination.value.itemCount = res.total || 0
  } catch (err) {
    console.error(err)
    message.error('加载题库失败')
  }
}

function handleAddClick() {
  isEdit.value = false
  formData.value = createEmptyForm()
  showModal.value = true
}

async function handleEdit(id) {
  try {
    const detail = await practiceApi.getDetail(id)
    isEdit.value = true
    formData.value = {
      id: detail.id,
      title: detail.title,
      description: detail.description || '',
      visible_roles: detail.visible_roles?.length ? detail.visible_roles : null,
      knowledge_tag_ids: detail.knowledge_tag_ids || [],
      type_counts: {
        ...createEmptyForm().type_counts,
        ...(detail.type_counts || {}),
      },
      selected_questions: detail.questions || [],
    }
    showModal.value = true
  } catch (err) {
    console.error(err)
    message.error('加载练习详情失败')
  }
}

async function handleDelete(id) {
  try {
    await practiceApi.delete(id)
    message.success('删除成功')
    loadPractices()
  } catch (err) {
    console.error(err)
    message.error('删除失败')
  }
}

async function handleAutoGenerate() {
  autoGenerating.value = true
  try {
    const res = await practiceApi.autoGenerate({
      knowledge_tag_ids: formData.value.knowledge_tag_ids || [],
      type_counts: normalizeTypeCounts(),
    })
    formData.value.selected_questions = res.questions || []

    const shortages = res.shortages || {}
    const shortageText = Object.entries(shortages)
      .filter(([, value]) => Number(value) > 0)
      .map(([key, value]) => `${typeMap[key] || key}缺少${value}题`)
      .join('，')

    if (shortageText) {
      message.warning(`自动抽题完成，${shortageText}`)
    } else {
      message.success('自动抽题完成')
    }
  } catch (err) {
    console.error(err)
    message.error('自动抽题失败')
  } finally {
    autoGenerating.value = false
  }
}

function openQuestionPicker() {
  selectedQuestionRowKeys.value = []
  questionPickerPagination.value.page = 1
  loadApprovedQuestions()
  showQuestionPicker.value = true
}

function addSelectedQuestions() {
  const incoming = approvedQuestions.value.filter((item) => selectedQuestionRowKeys.value.includes(item.id))
  formData.value.selected_questions = mergeQuestions(formData.value.selected_questions, incoming)
  selectedQuestionRowKeys.value = []
  showQuestionPicker.value = false
  message.success('已加入题目')
}

function removeSelectedQuestion(id) {
  formData.value.selected_questions = formData.value.selected_questions.filter((item) => item.id !== id)
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    const payload = {
      title: formData.value.title,
      description: formData.value.description || null,
      visible_roles: formData.value.visible_roles?.length ? formData.value.visible_roles : null,
      knowledge_tag_ids: formData.value.knowledge_tag_ids || [],
      type_counts: normalizeTypeCounts(),
      question_ids: formData.value.selected_questions.map((item) => item.id),
    }

    if (isEdit.value) {
      await practiceApi.update(formData.value.id, payload)
      message.success('练习更新成功')
    } else {
      await practiceApi.create(payload)
      message.success('练习创建成功')
    }

    showModal.value = false
    loadPractices()
  } catch (err) {
    console.error(err)
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadRoles()
  loadKnowledgeTags()
  loadPractices()
})
</script>

<style scoped>
.practices {
  padding: 20px 0;
}

.type-counts,
.selected-questions {
  margin-top: 12px;
}

.section-title {
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
}

.sub-text {
  margin-left: 8px;
  font-weight: 400;
  color: #888;
}
</style>
