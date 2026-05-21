<template>
  <div class="simulation-page">
    <div class="hero">
      <div class="hero-main">
        <div class="eyebrow">Student</div>
        <h1>模拟测试</h1>
  <p>支持题库随机出题，也可尝试智能体生成。</p>
      </div>
      <div class="hero-stats">
        <div class="stat-item">
          <span class="stat-label">题型</span>
          <strong>{{ effectiveTypeCount }}</strong>
        </div>
        <div class="stat-item">
          <span class="stat-label">难度</span>
          <strong>{{ form.difficulty || '-' }}</strong>
        </div>
      </div>
      <n-button v-if="!started" type="primary" @click="startTest">开始测试</n-button>
      <n-button v-else secondary @click="resetAll">重新开始</n-button>
    </div>

    <n-card v-if="!started" class="config-card" title="测试配置">
      <n-form label-placement="top" :model="form">
        <n-grid :cols="2" :x-gap="16" :y-gap="12">
          <n-gi>
            <n-form-item label="课程 / 分类ID">
              <n-select v-model:value="form.course_id" :options="courseOptions" filterable clearable placeholder="可选：选择课程范围" />
            </n-form-item>
            <n-alert
              v-if="showKnowledgeHint"
              type="warning"
              :bordered="false"
              style="margin-top: -8px"
            >
              当前分类下未检测到子知识点，Q矩阵会显示学科级名称。建议先在“分类管理”中添加子分类并绑定内容。
            </n-alert>
          </n-gi>
          <n-gi>
            <n-form-item label="题目数量">
              <n-input-number v-model:value="form.question_count" :min="1" :max="50" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="难度">
              <n-select v-model:value="form.difficulty" :options="difficultyOptions" clearable />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="时间限制（分钟）">
              <n-input-number v-model:value="form.time_limit_minutes" :min="1" :max="180" style="width: 100%" clearable />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :cols="2" :x-gap="16" :y-gap="12">
          <n-gi>
            <n-form-item label="出题方式">
              <n-select v-model:value="form.generation_mode" :options="generationOptions" />
            </n-form-item>
          </n-gi>
          <n-gi v-if="form.generation_mode === 'practice'">
            <n-form-item label="练习卷">
              <n-select
                v-model:value="form.practice_id"
                :options="practiceOptions"
                filterable
                clearable
                placeholder="请选择练习卷"
              />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-form-item label="题型">
          <n-checkbox-group v-model:value="form.question_types">
            <n-space wrap class="type-pills">
              <n-checkbox value="single_choice">单选题</n-checkbox>
              <n-checkbox value="multi_choice">多选题</n-checkbox>
              <n-checkbox value="true_false">判断题</n-checkbox>
              <n-checkbox value="fill_blank">填空题</n-checkbox>
              <n-checkbox value="short_answer">简答题</n-checkbox>
            </n-space>
          </n-checkbox-group>
        </n-form-item>

        <n-card
          v-if="form.generation_mode === 'random'"
          size="small"
          title="按题型数量出题（可选）"
          style="margin-top: 8px"
        >
          <n-grid :cols="5" :x-gap="12" :y-gap="12">
            <n-gi>
              <n-form-item label="单选题">
                <n-input-number v-model:value="form.type_counts.single_choice" :min="0" :max="50" style="width: 100%" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="多选题">
                <n-input-number v-model:value="form.type_counts.multi_choice" :min="0" :max="50" style="width: 100%" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="判断题">
                <n-input-number v-model:value="form.type_counts.true_false" :min="0" :max="50" style="width: 100%" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="填空题">
                <n-input-number v-model:value="form.type_counts.fill_blank" :min="0" :max="50" style="width: 100%" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="简答题">
                <n-input-number v-model:value="form.type_counts.short_answer" :min="0" :max="50" style="width: 100%" />
              </n-form-item>
            </n-gi>
          </n-grid>
          <div class="count-tip">已配置总题数：{{ typeCountTotal }}</div>
        </n-card>
      </n-form>
    </n-card>

    <n-spin :show="loading">
      <div v-if="started && questions.length > 0" class="test-area">
        <div class="test-main">
          <n-alert v-if="timeLeftText" type="info" class="timer">剩余时间：{{ timeLeftText }}</n-alert>

          <n-card
            v-for="(q, index) in questions"
            :key="q.id"
            class="question-card"
            :id="`question-${q.id}`"
          >
            <template #header>
              <div class="question-head">
                <span>第 {{ index + 1 }} 题</span>
                <n-tag size="small">{{ questionTypeLabelMap[q.question_type] || q.question_type }}</n-tag>
              </div>
            </template>
            <div class="stem">{{ q.stem }}</div>
            <template v-if="q.question_type === 'multi_choice'">
              <n-checkbox-group v-model:value="multiAnswers[q.id]">
                <n-space vertical>
                  <n-checkbox v-for="(opt, optIndex) in q.options" :key="optIndex" :value="optIndex">
                    {{ String.fromCharCode(65 + optIndex) }}. {{ opt }}
                  </n-checkbox>
                </n-space>
              </n-checkbox-group>
            </template>
            <template v-else-if="q.question_type === 'true_false'">
              <n-button-group>
                <n-button
                  :type="judgeAnswers[q.id] === 0 ? 'primary' : 'default'"
                  @click="judgeAnswers[q.id] = 0"
                >
                  对
                </n-button>
                <n-button
                  :type="judgeAnswers[q.id] === 1 ? 'primary' : 'default'"
                  @click="judgeAnswers[q.id] = 1"
                >
                  错
                </n-button>
              </n-button-group>
            </template>
            <template v-else-if="['fill_blank', 'short_answer'].includes(q.question_type)">
              <n-input
                v-model:value="textAnswers[q.id]"
                type="textarea"
                :autosize="{ minRows: 2, maxRows: 4 }"
                placeholder="请输入你的答案"
              />
            </template>
            <n-radio-group v-else v-model:value="answers[q.id]">
              <n-space vertical>
                <n-radio v-for="(opt, optIndex) in q.options" :key="optIndex" :value="optIndex">
                  {{ String.fromCharCode(65 + optIndex) }}. {{ opt }}
                </n-radio>
              </n-space>
            </n-radio-group>
            <div class="knowledge">知识点：{{ q.knowledge_tag || '未标注' }}</div>
          </n-card>

          <div class="actions">
            <n-button type="primary" size="large" @click="submitTest">提交测试</n-button>
          </div>
        </div>

        <aside class="answer-sheet">
          <n-card title="答题卡" size="small" class="sheet-card">
            <div class="sheet-summary">
              已答 {{ answeredCount }} / {{ questions.length }}
            </div>
            <div class="sheet-grid">
              <button
                v-for="(q, index) in questions"
                :key="`sheet-${q.id}`"
                class="sheet-item"
                :class="{
                  answered: ['fill_blank', 'short_answer'].includes(q.question_type)
                    ? Boolean(textAnswers[q.id])
                    : (answers[q.id] !== undefined && answers[q.id] !== null)
                }"
                @click="scrollToQuestion(q.id)"
              >
                {{ index + 1 }}
              </button>
            </div>
            <n-button block type="primary" style="margin-top: 12px" @click="submitTest">
              交卷
            </n-button>
          </n-card>
        </aside>
      </div>
    </n-spin>

    <n-result
      v-if="!loading && started && questions.length === 0"
      status="info"
      title="暂无题目"
      description="当前条件下没有可用题目，请调整筛选条件后重试。"
    />

    <n-modal v-model:show="showResult" preset="card" title="测试结果" style="width: 760px">
      <n-descriptions :column="2" bordered>
        <n-descriptions-item label="正确率">{{ result?.accuracy }}%</n-descriptions-item>
        <n-descriptions-item label="答对题数">{{ result?.correct_count }}/{{ result?.total_questions }}</n-descriptions-item>
        <n-descriptions-item label="已答题数">{{ result?.answered_questions }}</n-descriptions-item>
        <n-descriptions-item label="薄弱知识点">{{ result?.weak_knowledge_points?.join('、') || '无' }}</n-descriptions-item>
      </n-descriptions>
      <n-divider />
      <n-space vertical class="result-list">
        <div class="result-row" v-for="(score, kp) in result?.knowledge_accuracy || {}" :key="kp">
          {{ kp }}：{{ score }}%
        </div>
      </n-space>
      <n-divider />
      <div class="result-title">知识掌握度</div>
      <n-space vertical class="result-list">
        <div class="result-row" v-for="(score, kp) in result?.knowledge_mastery || {}" :key="`m-${kp}`">
          {{ kp }}：{{ Math.round((score || 0) * 100) }}%
        </div>
      </n-space>
      <n-divider />
      <div class="result-title">Q矩阵（题目-知识点作答）</div>
      <n-table :bordered="false" size="small" class="q-matrix-table">
        <thead>
          <tr>
            <th>题目ID</th>
            <th>知识点</th>
            <th>作答结果</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, idx) in result?.q_matrix || []" :key="`q-${idx}`">
            <td class="mono">{{ row.question_id }}</td>
            <td>{{ row.knowledge_tag }}</td>
            <td>
              <n-tag :type="row.is_correct ? 'success' : 'error'" size="small">
                {{ row.is_correct ? '正确' : '错误' }}
              </n-tag>
            </td>
          </tr>
          <tr v-if="!(result?.q_matrix || []).length">
            <td colspan="3" class="empty">暂无数据</td>
          </tr>
        </tbody>
      </n-table>
      <template #action>
        <n-button type="primary" @click="showResult = false">关闭</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NCheckbox,
  NCheckboxGroup,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NFlex,
  NForm,
  NFormItem,
  NGrid,
  NGi,
  NInputNumber,
  NInput,
  NModal,
  NRadio,
  NRadioGroup,
  NButtonGroup,
  NResult,
  NSelect,
  NSpace,
  NSpin,
  NTable,
  NTag,
  useMessage
} from 'naive-ui'
import { simulationApi, categoryApi, practiceApi } from '@/api'

const message = useMessage()
const loading = ref(false)
const started = ref(false)
const questions = ref([])
const answers = ref({})
const textAnswers = ref({})
const multiAnswers = ref({})
const judgeAnswers = ref({})
const showResult = ref(false)
const result = ref(null)
const timer = ref(null)
const timeLeft = ref(0)
const courseOptions = ref([])
const courseTree = ref([])
const practiceOptions = ref([])

const form = ref({
  course_id: '',
  practice_id: '',
  question_count: 10,
  difficulty: null,
  question_types: [],
  type_counts: {
    single_choice: 0,
    multi_choice: 0,
    true_false: 0,
    fill_blank: 0,
    short_answer: 0
  },
  time_limit_minutes: null,
  generation_mode: 'random'
})

const difficultyOptions = [
  { label: '1', value: 1 },
  { label: '2', value: 2 },
  { label: '3', value: 3 },
  { label: '4', value: 4 },
  { label: '5', value: 5 }
]

const generationOptions = [
  { label: '题库抽题', value: 'random' },
  { label: '练习卷', value: 'practice' },
  { label: '智能体生成', value: 'ai' }
]

const questionTypeLabelMap = {
  single_choice: '单选题',
  multi_choice: '多选题',
  true_false: '判断题',
  fill_blank: '填空题',
  short_answer: '简答题'
}

const typeCountTotal = computed(() => {
  const counts = form.value.type_counts || {}
  return Object.values(counts).reduce((sum, item) => sum + (Number(item) || 0), 0)
})

const effectiveTypeCount = computed(() => {
  if (form.value.generation_mode === 'random' && typeCountTotal.value > 0) {
    return Object.values(form.value.type_counts || {}).filter((v) => Number(v) > 0).length
  }
  return form.value.question_types?.length || 0
})

const answeredCount = computed(() => {
  return questions.value.filter((q) => {
    if (['fill_blank', 'short_answer'].includes(q.question_type)) {
      return Boolean(textAnswers.value[q.id])
    }
    if (q.question_type === 'multi_choice') {
      return Array.isArray(multiAnswers.value[q.id]) && multiAnswers.value[q.id].length > 0
    }
    if (q.question_type === 'true_false') {
      return judgeAnswers.value[q.id] !== undefined && judgeAnswers.value[q.id] !== null
    }
    return answers.value[q.id] !== undefined && answers.value[q.id] !== null
  }).length
})

function scrollToQuestion(questionId) {
  const el = document.getElementById(`question-${questionId}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

const showKnowledgeHint = computed(() => {
  if (!form.value.course_id) return false
  const target = findCategoryById(courseTree.value, form.value.course_id)
  if (!target) return false
  const isTopLevel = !target.parent_id
  const hasChildren = !!(target.children && target.children.length > 0)
  return isTopLevel && !hasChildren
})

function findCategoryById(nodes, id) {
  for (const node of nodes || []) {
    if (node.id === id) return node
    const child = findCategoryById(node.children || [], id)
    if (child) return child
  }
  return null
}

const timeLeftText = computed(() => {
  if (!started.value || !timeLeft.value) return ''
  const m = Math.floor(timeLeft.value / 60)
  const s = timeLeft.value % 60
  return `${m}:${String(s).padStart(2, '0')}`
})

async function loadCourses() {
  try {
    const res = await categoryApi.getTree({ max_depth: 100 })
    courseTree.value = res || []
    const flatten = []
    const walk = (items, prefix = '') => {
      ;(items || []).forEach((item) => {
        flatten.push({ label: prefix + item.name, value: item.id })
        if (item.children?.length) walk(item.children, `${prefix}${item.name} / `)
      })
    }
    walk(res || [])
    courseOptions.value = flatten
  } catch (err) {
    console.error(err)
  }
}

async function loadPractices() {
  try {
    const res = await practiceApi.getList({ skip: 0, limit: 200 })
    practiceOptions.value = (res.practices || []).map((item) => ({
      label: `${item.title}（${item.question_count || 0}题）`,
      value: item.id
    }))
  } catch (err) {
    console.error(err)
  }
}

function clearTimer() {
  if (timer.value) {
    clearInterval(timer.value)
    timer.value = null
  }
}

async function startTest() {
  loading.value = true
  try {
    if (form.value.generation_mode === 'practice' && !form.value.practice_id) {
      message.warning('请选择练习卷')
      return
    }
    const payload = {
      ...form.value,
      question_types: form.value.question_types.length ? form.value.question_types : null,
      type_counts:
        form.value.generation_mode === 'random' && typeCountTotal.value > 0
          ? Object.fromEntries(
              Object.entries(form.value.type_counts || {}).filter(([, v]) => Number(v) > 0)
            )
          : null,
      practice_id: form.value.generation_mode === 'practice' ? (form.value.practice_id || null) : null,
      course_id: form.value.course_id || null,
      difficulty: form.value.difficulty || null,
      time_limit_minutes: form.value.time_limit_minutes || null,
      question_count:
        form.value.generation_mode === 'random' && typeCountTotal.value > 0
          ? typeCountTotal.value
          : (form.value.question_count || 10)
    }
    const res = await simulationApi.generate(payload)
    questions.value = res.questions || []
    answers.value = {}
  textAnswers.value = {}
  multiAnswers.value = {}
  judgeAnswers.value = {}
    started.value = true

    clearTimer()
    if (payload.time_limit_minutes) {
      timeLeft.value = payload.time_limit_minutes * 60
      timer.value = setInterval(() => {
        timeLeft.value -= 1
        if (timeLeft.value <= 0) {
          clearTimer()
          submitTest()
        }
      }, 1000)
    } else {
      timeLeft.value = 0
    }
  } catch (err) {
    console.error(err)
    message.error(err.response?.data?.detail || '生成测试失败')
  } finally {
    loading.value = false
  }
}

async function submitTest() {
  if (!questions.value.length) return
  clearTimer()
  loading.value = true
  try {
    const payload = {
      questions: questions.value,
      answers: questions.value.map((q) => ({
        question_id: q.id,
        selected_index: (() => {
          if (['fill_blank', 'short_answer'].includes(q.question_type)) return null
          if (q.question_type === 'multi_choice') {
            const picks = multiAnswers.value[q.id] || []
            return picks.length ? picks[0] : null
          }
          if (q.question_type === 'true_false') return judgeAnswers.value[q.id] ?? null
          return answers.value[q.id] ?? null
        })(),
        answer_text: (() => {
          if (['fill_blank', 'short_answer'].includes(q.question_type)) {
            return (textAnswers.value[q.id] || '').trim()
          }
          if (q.question_type === 'multi_choice') {
            return JSON.stringify(multiAnswers.value[q.id] || [])
          }
          return null
        })()
      }))
    }
    const res = await simulationApi.submit(payload)
    result.value = res
    showResult.value = true
  } catch (err) {
    console.error(err)
    message.error(err.response?.data?.detail || '提交测试失败')
  } finally {
    loading.value = false
  }
}

function resetAll() {
  clearTimer()
  started.value = false
  questions.value = []
  answers.value = {}
  textAnswers.value = {}
  multiAnswers.value = {}
  judgeAnswers.value = {}
  result.value = null
  showResult.value = false
  timeLeft.value = 0
}

onBeforeUnmount(() => clearTimer())
loadCourses()
loadPractices()
</script>

<style scoped>
.simulation-page {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 10px;
  border-radius: 20px;
  background:
    radial-gradient(circle at 10% 10%, #e7f1ff 0, transparent 35%),
    radial-gradient(circle at 90% 20%, #dff3ff 0, transparent 28%),
    linear-gradient(180deg, #f5faff 0%, #ffffff 45%);
}
.hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 18px;
  padding: 26px;
  border-radius: 20px;
  background: linear-gradient(135deg, #f8fbff 0%, #ecf5ff 50%, #dcecff 100%);
  color: #0f172a;
  border: 1px solid #cfe3ff;
  box-shadow: 0 14px 30px rgba(50, 115, 220, 0.12);
}
.hero-main {
  flex: 1;
}
.hero-stats {
  display: flex;
  gap: 10px;
}
.stat-item {
  min-width: 76px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid #d8e9ff;
  text-align: center;
}
.stat-item strong {
  color: #1e40af;
  font-size: 20px;
}
.stat-label {
  display: block;
  color: #64748b;
  font-size: 12px;
}
.eyebrow {
  font-size: 12px;
  color: #2563eb;
  text-transform: uppercase;
  letter-spacing: 0.18em;
  font-weight: 600;
}
.hero h1 {
  margin: 8px 0;
  font-size: 34px;
  color: #12284a;
}
.hero p {
  margin: 0;
  color: #45617f;
}
.config-card,
.question-card {
  border-radius: 18px;
  border: 1px solid #deebff;
  box-shadow: 0 10px 28px rgba(31, 84, 168, 0.08);
}

.config-card {
  background: linear-gradient(180deg, #ffffff, #f8fbff);
  border: 1px solid #cfe3ff;
  box-shadow: 0 12px 30px rgba(52, 108, 195, 0.1);
}
.question-card {
  margin-bottom: 16px;
  background: linear-gradient(180deg, #fbfdff 0%, #f2f8ff 100%);
  border: 1px solid #d7e6ff;
  box-shadow: 0 8px 20px rgba(56, 108, 196, 0.08);
}
.question-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.stem {
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
  color: #14325d;
  line-height: 1.6;
  background: rgba(255, 255, 255, 0.62);
  border: 1px solid #dbe9ff;
  border-radius: 10px;
  padding: 10px 12px;
}
.knowledge {
  margin-top: 14px;
  color: #35587d;
  font-size: 13px;
  background: #eaf3ff;
  border: 1px dashed #bcd4f6;
  border-radius: 10px;
  padding: 8px 10px;
}
.actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 22px;
}
.test-area {
  display: grid;
  width: 100%;
  grid-template-columns: minmax(0, 1fr) minmax(340px, 24%);
  gap: 16px;
  align-items: start;
}
.test-main {
  min-width: 0;
}
.answer-sheet {
  position: sticky;
  top: 16px;
}
.sheet-card {
  width: 100%;
  border-radius: 14px;
}
.sheet-summary {
  color: #31527b;
  font-weight: 600;
  margin-bottom: 10px;
}
.sheet-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
}
.sheet-item {
  height: 34px;
  border-radius: 8px;
  border: 1px solid #cfe1ff;
  background: #f3f8ff;
  color: #24446e;
  font-weight: 600;
  cursor: pointer;
}
.sheet-item.answered {
  background: #dbeafe;
  border-color: #60a5fa;
  color: #1d4ed8;
}
.sheet-item:hover {
  border-color: #3b82f6;
}
.timer {
  margin-bottom: 16px;
}
.count-tip {
  color: #35587d;
  font-size: 13px;
  margin-top: 4px;
}
.type-pills :deep(.n-checkbox) {
  padding: 8px 12px;
  background: #f4f9ff;
  border: 1px solid #cfe1ff;
  border-radius: 999px;
  transition: all 0.2s ease;
}
.type-pills :deep(.n-checkbox:hover) {
  border-color: #7fb0f8;
  background: #eaf3ff;
}
.type-pills :deep(.n-checkbox.n-checkbox--checked) {
  background: #e1eeff;
  border-color: #78a8f0;
}

:deep(.config-card .n-card-header) {
  background: linear-gradient(90deg, rgba(220, 237, 255, 0.9), rgba(239, 247, 255, 0.55));
  border-radius: 14px;
  margin: -2px -2px 14px;
  padding: 12px 14px;
}

:deep(.config-card .n-form-item-label__text) {
  color: #1f3f72;
  font-weight: 600;
}

:deep(.config-card .n-base-selection),
:deep(.config-card .n-input-number),
:deep(.config-card .n-input-number .n-input__input-el) {
  background: #f8fbff;
}

:deep(.config-card .n-base-selection) {
  border: 1px solid #cfe0fa;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.75);
}

:deep(.config-card .n-input-number) {
  border: 1px solid #cfe0fa;
}

:deep(.config-card .n-base-selection:hover),
:deep(.config-card .n-input-number:hover) {
  border-color: #8cb7f3;
}

:deep(.config-card .n-base-selection.n-base-selection--focus),
:deep(.config-card .n-input-number.n-input-number--focus) {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.14);
}
.result-title {
  font-weight: 700;
  color: #1e3a8a;
  margin-bottom: 8px;
}
.result-list {
  background: #f8fbff;
  border: 1px solid #e3eeff;
  border-radius: 12px;
  padding: 10px;
}
.result-row {
  color: #1f3656;
}
.result-row.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
}
.q-matrix-table {
  margin-top: 6px;
  background: #f8fbff;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #e1ebff;
}
.q-matrix-table th {
  background: #eef4ff;
  color: #1e3a8a;
  font-weight: 600;
}
.q-matrix-table td {
  color: #1f2a44;
}
.q-matrix-table .mono {
  font-size: 12px;
  color: #475569;
}
.q-matrix-table .empty {
  text-align: center;
  color: #94a3b8;
  padding: 16px 0;
}

:deep(.hero .n-button--primary-type) {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  border: none;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.28);
}

:deep(.hero .n-button--secondary-type) {
  background: #ffffff;
  color: #1d4ed8;
  border: 1px solid #bfdbfe;
}
:deep(.n-card-header__main) {
  color: #183b72;
  font-weight: 700;
}
:deep(.question-card .n-card-header) {
  background: linear-gradient(90deg, rgba(221, 236, 255, 0.82), rgba(235, 244, 255, 0.52));
  border-radius: 12px;
  margin: -2px -2px 10px;
  padding: 10px 12px;
}
:deep(.n-input-number),
:deep(.n-base-selection) {
  border-radius: 10px;
}

@media (max-width: 760px) {
  .test-area {
    grid-template-columns: 1fr;
  }
  .answer-sheet {
    position: static;
  }
  .hero {
    flex-direction: column;
    align-items: flex-start;
  }
  .hero-stats {
    width: 100%;
  }
  .stat-item {
    flex: 1;
  }
}
</style>
