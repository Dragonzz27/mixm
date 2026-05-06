<template>
  <div class="questions">
    <n-card title="题库管理">
      <template #header-extra>
        <n-space>
          <n-select
            v-model:value="filterKnowledgeTag"
            :options="knowledgeTagOptions"
            placeholder="筛选知识元"
            clearable
            style="width: 180px"
            @update:value="loadQuestions"
          />
          <n-select
            v-model:value="filterType"
            :options="typeOptions"
            placeholder="筛选题型"
            clearable
            style="width: 140px"
            @update:value="loadQuestions"
          />
          <n-select
            v-model:value="filterDifficulty"
            :options="difficultyOptions"
            placeholder="筛选难度"
            clearable
            style="width: 120px"
            @update:value="loadQuestions"
          />
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索题目..."
            style="width: 200px"
            clearable
            @keyup.enter="loadQuestions"
            @clear="loadQuestions"
          />
          <n-button type="primary" @click="handleAddClick">添加题目</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="questions"
        :pagination="questionPagination"
        remote
        :bordered="false"
      />
    </n-card>

    <n-modal v-model:show="showAddModal" preset="dialog" :title="isEdit ? '编辑题目' : '添加题目'" style="width: 800px; max-width: 95vw;">
      <n-form :model="formData" :rules="rules" ref="formRef" label-placement="top">
        <n-form-item label="题目内容" path="title">
          <n-input v-model:value="formData.title" placeholder="请输入题目内容" />
        </n-form-item>
        <n-row :gutter="16">
          <n-col :span="8">
            <n-form-item label="题型" path="type">
              <n-select v-model:value="formData.type" :options="typeOptions.filter(t => t.value)" />
            </n-form-item>
          </n-col>
          <n-col :span="8">
            <n-form-item label="难度" path="difficulty">
              <n-select v-model:value="formData.difficulty" :options="difficultyOptions.filter(d => d.value)" />
            </n-form-item>
          </n-col>
          <n-col :span="8">
            <n-form-item label="来源" path="source">
              <n-input v-model:value="formData.source" placeholder="请输入来源" />
            </n-form-item>
          </n-col>
        </n-row>
        <n-form-item label="知识元" path="knowledge_tag_ids">
          <n-tree-select
            v-model:value="formData.knowledge_tag_ids"
            :options="cascaderOptions"
            placeholder="选择知识元/知识点"
            multiple
            clearable
            filterable
            :max-height="200"
            style="width: 100%"
          />
        </n-form-item>

        <n-card size="small" title="智能体生成题目" v-if="!isEdit" style="margin-bottom: 16px;">
          <n-space align="center">
            <n-button type="primary" :loading="aiLoading" :disabled="!formData.knowledge_tag_ids?.length" @click="handleAiGenerate">
              生成题目（{{ typeOptions.find(t => t.value === formData.type)?.label }} / 难度{{ formData.difficulty }}）
            </n-button>
            <div v-if="aiStreaming" class="ai-streaming">
              <n-spin size="small" />
              <span>AI生成中: {{ aiContent }}</span>
            </div>
          </n-space>
        </n-card>
        <div v-if="['single_choice', 'multi_choice'].includes(formData.type)">
          <n-form-item label="选项" path="options">
            <div class="options-container">
              <n-space vertical :size="12">
                <n-space v-for="(opt, idx) in formData.options" :key="idx" align="center">
                  <n-tag :bordered="false" style="width: 30px; justify-content: center;">{{ opt.option_key }}</n-tag>
                  <n-input v-model:value="opt.content" placeholder="选项内容" style="width: 350px" />
                  <n-radio
                    v-if="formData.type === 'single_choice'"
                    :checked="opt.is_correct"
                    @update:checked="(val) => setSingleCorrect(idx, val)"
                  >
                    正确答案
                  </n-radio>
                  <n-checkbox v-if="formData.type === 'multi_choice'" v-model:checked="opt.is_correct">正确答案</n-checkbox>
                  <n-button size="small" type="error" quaternary @click="removeOption(idx)">删除</n-button>
                </n-space>
              </n-space>
              <n-button type="primary" text @click="addOption" style="margin-top: 8px">+ 添加选项</n-button>
            </div>
          </n-form-item>
        </div>
        <n-form-item v-if="formData.type === 'true_false'" label="答案" path="answer">
          <n-select
            v-model:value="formData.true_false_answer"
            :options="[
              { label: '正确', value: true },
              { label: '错误', value: false }
            ]"
            placeholder="请选择答案"
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item v-if="!['single_choice', 'multi_choice', 'true_false'].includes(formData.type)" label="答案" path="answer">
          <n-input v-model:value="formData.answer" type="textarea" placeholder="请输入答案" :rows="2" />
        </n-form-item>
        <n-form-item label="解析" path="explanation">
          <n-input v-model:value="formData.explanation" type="textarea" placeholder="请输入解析" :rows="2" />
        </n-form-item>
        <n-form-item label="简单标签" path="tags">
          <n-input v-model:value="tagsInput" placeholder="请输入标签，用逗号分隔（非知识元）" />
        </n-form-item>
        <n-form-item label="可见角色" path="visible_roles">
          <n-select
            v-model:value="formData.visible_roles"
            :options="roleOptions"
            placeholder="选择可见角色 (留空表示所有角色可见)"
            multiple
            clearable
            style="width: 100%"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showAddModal = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleSubmit">确定</n-button>
      </template>
    </n-modal>

    <n-modal v-model:show="showDetailModal" preset="card" title="题目详情" style="width: 700px; max-width: 95vw;">
      <n-descriptions :column="2" label-placement="left" v-if="currentQuestion">
        <n-descriptions-item label="题型">{{ typeMap[currentQuestion.type] || currentQuestion.type }}</n-descriptions-item>
        <n-descriptions-item label="难度">{{ currentQuestion.difficulty }}</n-descriptions-item>
        <n-descriptions-item label="来源">{{ currentQuestion.source || '-' }}</n-descriptions-item>
        <n-descriptions-item label="AI生成">{{ currentQuestion.is_ai_generated ? '是' : '否' }}</n-descriptions-item>
        <n-descriptions-item label="创建时间">{{ formatDate(currentQuestion.created_at) }}</n-descriptions-item>
      </n-descriptions>
      <div v-if="currentQuestion?.knowledge_tags?.length > 0" style="margin-top: 16px">
        <n-divider>知识元/知识点</n-divider>
        <n-space>
          <n-tag v-for="tag in currentQuestion.knowledge_tags" :key="tag.id" type="success">{{ tag.name }}</n-tag>
        </n-space>
      </div>
      <n-divider>题目内容</n-divider>
      <div class="question-content">{{ currentQuestion?.title }}</div>
      <div v-if="currentQuestion?.options?.length > 0" style="margin-top: 16px">
        <n-divider>选项</n-divider>
        <n-space vertical>
          <n-space v-for="opt in currentQuestion.options" :key="opt.id" align="center">
            <n-tag :type="opt.is_correct ? 'success' : 'default'" :bordered="false">{{ opt.option_key }}</n-tag>
            <span>{{ opt.content }}</span>
          </n-space>
        </n-space>
      </div>
      <div v-if="currentQuestion?.answer" style="margin-top: 16px">
        <n-divider>答案</n-divider>
        <div>{{ currentQuestion.answer }}</div>
      </div>
      <div v-if="currentQuestion?.explanation" style="margin-top: 16px">
        <n-divider>解析</n-divider>
        <div>{{ currentQuestion.explanation }}</div>
      </div>
      <div v-if="currentQuestion?.tags?.length > 0" style="margin-top: 16px">
        <n-divider>简单标签</n-divider>
        <n-space>
          <n-tag v-for="tag in currentQuestion.tags" :key="tag" type="info">{{ tag }}</n-tag>
        </n-space>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { NCard, NButton, NInput, NDataTable, NSpace, NModal, NForm, NFormItem, NSelect, NTag, NDivider, NDescriptions, NDescriptionsItem, NRow, NCol, NCheckbox, NRadio, NTreeSelect, useMessage } from 'naive-ui'
import { questionApi } from '@/api'
import { roleApi, categoryApi } from '@/api'

const message = useMessage()
const formRef = ref(null)
const questions = ref([])
const showAddModal = ref(false)
const showDetailModal = ref(false)
const isEdit = ref(false)
const loading = ref(false)
const currentQuestion = ref(null)
const searchKeyword = ref('')
const filterKnowledgeTag = ref(null)
const filterType = ref(null)
const filterDifficulty = ref(null)
const knowledgeTagOptions = ref([])
const knowledgeTagTreeData = ref([])

const cascaderOptions = computed(() => {
  const transform = (nodes) => nodes
    .filter(node => node.name)
    .map(node => ({
      key: node.id,
      value: node.id,
      label: node.name,
      description: node.description || '',
      children: node.children?.length > 0 ? transform(node.children) : undefined
    }))
  return transform(knowledgeTagTreeData.value)
})

const roleOptions = ref([])
const tagsInput = ref('')

const typeOptions = [
  { label: '全部', value: null },
  { label: '单选题', value: 'single_choice' },
  { label: '多选题', value: 'multi_choice' },
  { label: '判断题', value: 'true_false' },
  { label: '填空题', value: 'blank' },
  { label: '简答题', value: 'essay' }
]

const difficultyOptions = [
  { label: '全部', value: null },
  { label: '1', value: 1 },
  { label: '2', value: 2 },
  { label: '3', value: 3 },
  { label: '4', value: 4 },
  { label: '5', value: 5 }
]

const typeMap = {
  single_choice: '单选题',
  multi_choice: '多选题',
  true_false: '判断题',
  blank: '填空题',
  essay: '简答题'
}

const questionPagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    questionPagination.value.page = page
    loadQuestions()
  },
  onUpdatePageSize: (pageSize) => {
    questionPagination.value.pageSize = pageSize
    questionPagination.value.page = 1
    loadQuestions()
  }
})

const formData = ref({
  title: '',
  type: 'single_choice',
  difficulty: 3,
  knowledge_tag_ids: null,
  answer: '',
  explanation: '',
  source: '',
  visible_roles: null,
  options: [
    { option_key: 'A', content: '', is_correct: false },
    { option_key: 'B', content: '', is_correct: false },
    { option_key: 'C', content: '', is_correct: false },
    { option_key: 'D', content: '', is_correct: false }
  ]
})

const aiLoading = ref(false)
const aiStreaming = ref(false)
const aiContent = ref('')

const rules = {
  title: [{ required: true, message: '请输入题目内容' }],
  type: [{ required: true, message: '请选择题型' }],
  difficulty: [{ required: true, message: '请选择难度' }]
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

const columns = computed(() => [
  { title: 'ID', key: 'id', width: 340 },
  { title: '题目', key: 'title', ellipsis: { tooltip: true } },
  { title: '题型', key: 'type', width: 100, render: (row) => typeMap[row.type] || row.type },
  { title: '难度', key: 'difficulty', width: 80, render: (row) => h(NTag, { type: getDifficultyType(row.difficulty), size: 'small' }, () => row.difficulty) },
  { title: '知识元', key: 'knowledge_tags', width: 200, render: (row) => {
    if (!row.knowledge_tags || row.knowledge_tags.length === 0) return '-'
    return h('div', {}, row.knowledge_tags.slice(0, 2).map(t => t.name).join(', ') + (row.knowledge_tags.length > 2 ? ` +${row.knowledge_tags.length - 2}` : ''))
  }},
  { title: 'AI生成', key: 'is_ai_generated', width: 80, render: (row) => row.is_ai_generated ? h(NTag, { type: 'warning', size: 'small' }, () => '是') : '-' },
  { title: '创建时间', key: 'created_at', width: 170, render: (row) => formatDate(row.created_at) },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    render: (row) => h(NSpace, {}, () => [
      h(NButton, { size: 'small', onClick: () => viewDetail(row) }, () => '查看'),
      h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => '编辑'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
])

function getDifficultyType(difficulty) {
  if (difficulty <= 2) return 'success'
  if (difficulty <= 3) return 'warning'
  return 'error'
}

async function loadKnowledgeTags() {
  try {
    const res = await categoryApi.getTree({ max_depth: 100 })
    knowledgeTagTreeData.value = res || []
  } catch (err) {
    console.error(err)
  }
}

async function loadRoles() {
  try {
    const res = await roleApi.getList({ skip: 0, limit: 100 })
    roleOptions.value = (res.roles || []).map(item => ({
      label: item.name,
      value: item.id
    }))
  } catch (err) {
    console.error(err)
  }
}

async function loadQuestions() {
  try {
    const page = questionPagination.value.page || 1
    const pageSize = questionPagination.value.pageSize || 10
    const params = {
      skip: (page - 1) * pageSize,
      limit: pageSize
    }
    if (filterKnowledgeTag.value) params.knowledge_tag_id = filterKnowledgeTag.value
    if (filterType.value) params.type = filterType.value
    if (filterDifficulty.value) params.difficulty = filterDifficulty.value
    if (searchKeyword.value) params.keyword = searchKeyword.value

    const res = await questionApi.getList(params)
    questions.value = res.questions || []
    questionPagination.value.itemCount = res.total || 0

    const maxPage = Math.max(1, Math.ceil((questionPagination.value.itemCount || 0) / pageSize))
    if (questionPagination.value.page > maxPage) {
      questionPagination.value.page = maxPage
      return loadQuestions()
    }
  } catch (err) {
    console.error(err)
    message.error('加载题库失败')
  }
}

function getKnowledgeItems(ids) {
  const items = []
  function findInTree(nodes) {
    for (const node of nodes) {
      if (ids.includes(node.id)) {
        items.push({
          name: node.name || node.label || '',
          description: node.description || ''
        })
      }
      if (node.children) findInTree(node.children)
    }
  }
  findInTree(knowledgeTagTreeData.value)
  return items
}

function handleAiGenerate() {
  if (!formData.value.knowledge_tag_ids?.length) {
    message.warning('请先选择知识元')
    return
  }
  aiLoading.value = true
  aiStreaming.value = true
  aiContent.value = ''
  formData.value.title = ''
  formData.value.answer = ''
  formData.value.explanation = ''

  const knowledgeItems = getKnowledgeItems(formData.value.knowledge_tag_ids)

  questionApi.generateQuestion(
    {
      knowledge_items: knowledgeItems,
      question_type: formData.value.type,
      difficulty: formData.value.difficulty
    },
    (chunk) => {
      aiContent.value += chunk
    },
    (data) => {
      formData.value.title = data.title || ''
      formData.value.type = data.type || formData.value.type
      formData.value.difficulty = data.difficulty || formData.value.difficulty
      formData.value.answer = data.answer || ''
      formData.value.explanation = data.explanation || ''

      if (data.options?.length) {
        formData.value.options = data.options.map(opt => ({
          option_key: opt.option_key,
          content: opt.content,
          is_correct: opt.is_correct
        }))
      }

      aiStreaming.value = false
      aiLoading.value = false
      message.success('题目生成成功')
    },
    (err) => {
      console.error(err)
      message.error(err.message || '生成失败')
      aiStreaming.value = false
      aiLoading.value = false
    }
  )
}

function handleAddClick() {
  isEdit.value = false
  formData.value = {
    title: '',
    type: 'single_choice',
    difficulty: 3,
    knowledge_tag_ids: null,
    answer: '',
    true_false_answer: null,
    explanation: '',
    source: '',
    visible_roles: null,
    options: [
      { option_key: 'A', content: '', is_correct: false },
      { option_key: 'B', content: '', is_correct: false },
      { option_key: 'C', content: '', is_correct: false },
      { option_key: 'D', content: '', is_correct: false }
    ]
  }
  tagsInput.value = ''
  showAddModal.value = true
}

async function handleEdit(row) {
  isEdit.value = true
  currentQuestion.value = row
  formData.value = {
    id: row.id,
    title: row.title,
    type: row.type,
    difficulty: row.difficulty,
    knowledge_tag_ids: getKnowledgeTagPaths(row.knowledge_tags),
    answer: row.answer || '',
    true_false_answer: row.type === 'true_false' ? (row.answer === 'true' || row.answer === true) : null,
    explanation: row.explanation || '',
    source: row.source || '',
    visible_roles: row.visible_roles && row.visible_roles.length > 0 ? row.visible_roles : null,
    options: row.options && row.options.length > 0 ? row.options.map(o => ({
      option_key: o.option_key,
      content: o.content,
      is_correct: o.is_correct
    })) : [
      { option_key: 'A', content: '', is_correct: false },
      { option_key: 'B', content: '', is_correct: false },
      { option_key: 'C', content: '', is_correct: false },
      { option_key: 'D', content: '', is_correct: false }
    ]
  }
  tagsInput.value = (row.tags || []).join(', ')
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
      tags: tagsInput.value ? tagsInput.value.split(',').map(t => t.trim()).filter(t => t) : null,
      visible_roles: formData.value.visible_roles && formData.value.visible_roles.length > 0 ? formData.value.visible_roles : null,
      options: ['single_choice', 'multi_choice'].includes(formData.value.type) ? formData.value.options.filter(o => o.content) : null
    }

    // 处理判断题答案
    if (formData.value.type === 'true_false') {
      data.answer = formData.value.true_false_answer ? 'true' : 'false'
    }

    if (isEdit.value) {
      await questionApi.update(formData.value.id, data)
      message.success('更新成功')
    } else {
      await questionApi.create(data)
      message.success('创建成功')
    }
    showAddModal.value = false
    loadQuestions()
  } catch (err) {
    console.error(err)
    message.error('操作失败')
  } finally {
    loading.value = false
  }
}

async function handleDelete(id) {
  try {
    await questionApi.delete(id)
    message.success('删除成功')
    loadQuestions()
  } catch (err) {
    console.error(err)
    message.error('删除失败')
  }
}

async function viewDetail(row) {
  currentQuestion.value = row
  showDetailModal.value = true
}

function addOption() {
  const keys = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H']
  const nextKey = keys[formData.value.options.length]
  if (nextKey) {
    formData.value.options.push({ option_key: nextKey, content: '', is_correct: false })
  }
}

function removeOption(idx) {
  if (formData.value.options.length > 2) {
    formData.value.options.splice(idx, 1)
    const keys = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H']
    formData.value.options.forEach((opt, i) => {
      opt.option_key = keys[i]
    })
  }
}

function setSingleCorrect(idx, val) {
  if (val) {
    formData.value.options.forEach((opt, i) => {
      opt.is_correct = i === idx
    })
  }
}

onMounted(() => {
  loadKnowledgeTags()
  loadRoles()
  loadQuestions()
})
</script>

<style scoped>
.questions {
  padding: 20px 0;
}
.question-content {
  font-size: 16px;
  line-height: 1.8;
  padding: 12px;
  background: #f9f9f9;
  border-radius: 4px;
}
.ai-streaming {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f5f5f5;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
}
</style>