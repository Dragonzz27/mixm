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
      <n-button v-else-if="isSubmitted" secondary @click="backToConfig">返回配置</n-button>
    </div>

    <n-card v-if="!started" class="config-card" title="测试配置">
      <n-form label-placement="top" :model="form">
        <n-grid :cols="2" :x-gap="16" :y-gap="12">
          <n-gi>
            <n-form-item label="课程 / 分类ID">
              <n-select
                v-model:value="form.course_id"
                :options="courseOptions"
                filterable
                clearable
                placeholder="可选：选择课程范围"
              />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="题目数量">
              <n-input-number
                v-model:value="form.question_count"
                :min="1"
                :max="50"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="难度">
              <n-select v-model:value="form.difficulty" :options="difficultyOptions" clearable />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="时间限制（分钟）">
              <n-input-number
                v-model:value="form.time_limit_minutes"
                :min="1"
                :max="180"
                style="width: 100%"
                clearable
              />
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
                <n-input-number
                  v-model:value="form.type_counts.single_choice"
                  :min="0"
                  :max="50"
                  style="width: 100%"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="多选题">
                <n-input-number
                  v-model:value="form.type_counts.multi_choice"
                  :min="0"
                  :max="50"
                  style="width: 100%"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="判断题">
                <n-input-number
                  v-model:value="form.type_counts.true_false"
                  :min="0"
                  :max="50"
                  style="width: 100%"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="填空题">
                <n-input-number
                  v-model:value="form.type_counts.fill_blank"
                  :min="0"
                  :max="50"
                  style="width: 100%"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="简答题">
                <n-input-number
                  v-model:value="form.type_counts.short_answer"
                  :min="0"
                  :max="50"
                  style="width: 100%"
                />
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
          <n-alert v-if="timeLeftText" type="info" class="timer"
            >剩余时间：{{ timeLeftText }}</n-alert
          >

          <n-card
            v-for="(q, index) in questions"
            :key="q.id"
            class="question-card"
            :id="`question-${q.id}`"
          >
            <template #header>
              <div class="question-head">
                <span>第 {{ index + 1 }} 题</span>
                <n-space>
                  <n-tag size="small">{{
                    questionTypeLabelMap[q.question_type] || q.question_type
                  }}</n-tag>
                  <n-tag size="small" type="warning">{{ Number(q.score || 0).toFixed(2) }}分</n-tag>
                </n-space>
              </div>
            </template>
            <div class="stem">{{ q.stem }}</div>
            <template v-if="q.question_type === 'multi_choice'">
              <n-checkbox-group v-model:value="multiAnswers[q.id]">
                <n-space vertical>
                  <n-checkbox
                    v-for="(opt, optIndex) in q.options"
                    :key="optIndex"
                    :value="optIndex"
                  >
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
            <n-button type="primary" size="large" :disabled="isSubmitted" @click="submitTest">
              {{ isSubmitted ? '已提交' : '提交测试' }}
            </n-button>
          </div>
        </div>

        <aside class="answer-sheet">
          <n-card title="答题卡" size="small" class="sheet-card">
            <div class="sheet-summary">已答 {{ answeredCount }} / {{ questions.length }}</div>
            <div class="sheet-summary">总分：{{ totalPaperScore }} 分</div>
            <div class="sheet-grid">
              <button
                v-for="(q, index) in questions"
                :key="`sheet-${q.id}`"
                class="sheet-item"
                :class="{
                  answered: ['fill_blank', 'short_answer'].includes(q.question_type)
                    ? Boolean(textAnswers[q.id])
                    : answers[q.id] !== undefined && answers[q.id] !== null
                }"
                @click="scrollToQuestion(q.id)"
              >
                {{ index + 1 }}
              </button>
            </div>
            <n-button
              block
              type="primary"
              style="margin-top: 12px"
              :disabled="isSubmitted"
              @click="submitTest"
            >
              {{ isSubmitted ? '已交卷' : '交卷' }}
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
        <n-descriptions-item label="试卷总分">{{ result?.total_score ?? 100 }}</n-descriptions-item>
        <n-descriptions-item label="本次得分">{{ result?.score ?? 0 }}</n-descriptions-item>
        <n-descriptions-item label="正确率">{{ result?.accuracy }}%</n-descriptions-item>
        <n-descriptions-item label="答对题数"
          >{{ result?.correct_count }}/{{ result?.total_questions }}</n-descriptions-item
        >
        <n-descriptions-item label="已答题数">{{ result?.answered_questions }}</n-descriptions-item>
        <n-descriptions-item label="薄弱知识点">{{
          result?.weak_knowledge_points?.join('、') || '无'
        }}</n-descriptions-item>
      </n-descriptions>
      <n-divider />
      <n-space vertical class="result-list">
        <div class="result-row" v-for="(score, kp) in result?.knowledge_accuracy || {}" :key="kp">
          {{ kp }}：{{ score }}%
        </div>
      </n-space>
      <template #action>
        <n-space>
          <n-button secondary @click="openAnswerReview">查看答案与解析</n-button>
          <n-button type="primary" @click="showResult = false">关闭</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showReview" preset="card" title="答案与解析" style="width: 860px">
      <n-space vertical class="result-list">
        <div v-for="(q, idx) in reviewData.questions" :key="`review-${q.id}`" class="analysis-item">
          <div class="analysis-head">
            <strong>第{{ idx + 1 }}题</strong>
            <n-tag size="small" :type="reviewIsCorrect(q) ? 'success' : 'error'">
              {{ reviewIsCorrect(q) ? '答对' : '答错' }}
            </n-tag>
          </div>
          <div class="analysis-line">题目：{{ q.stem }}</div>
          <div class="analysis-line">你的答案：{{ reviewUserAnswer(q) }}</div>
          <div class="analysis-line">正确答案：{{ reviewCorrectAnswer(q) }}</div>
          <div class="analysis-line">知识点：{{ q.knowledge_tag || '未标注' }}</div>
        </div>
      </n-space>
      <template #action>
        <n-button type="primary" @click="showReview = false">关闭</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, ref } from 'vue'
import request from '@/api/request'
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
const showReview = ref(false)
const result = ref(null)
const reviewData = ref({ questions: [], answers: [] })
const timer = ref(null)
const timeLeft = ref(0)
const courseOptions = ref([])
const practiceOptions = ref([])
const currentAttemptId = ref('')
const submittedAttemptId = ref('')
const isSubmitted = ref(false)

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

const totalPaperScore = computed(() => {
  const sum = (questions.value || []).reduce((acc, q) => acc + (Number(q.score) || 0), 0)
  return sum.toFixed(2)
})

function scrollToQuestion(questionId) {
  const el = document.getElementById(`question-${questionId}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
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
    const res = await practiceApi.getList({ offset: 0, limit: 200 })
    practiceOptions.value = (res.items || []).map((item) => ({
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
      practice_id:
        form.value.generation_mode === 'practice' ? form.value.practice_id || null : null,
      course_id: form.value.course_id || null,
      difficulty: form.value.difficulty || null,
      time_limit_minutes: form.value.time_limit_minutes || null,
      question_count:
        form.value.generation_mode === 'random' && typeCountTotal.value > 0
          ? typeCountTotal.value
          : form.value.question_count || 10
    }
    const res = await simulationApi.generate(payload)
    questions.value = res.questions || []
    currentAttemptId.value = res.attempt_id || ''
    isSubmitted.value = (res.attempt_status || '') === 'submitted'
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
    message.error(err.response?.data?.message || '生成测试失败')
  } finally {
    loading.value = false
  }
}

async function submitTest() {
  if (!questions.value.length || isSubmitted.value) return
  clearTimer()
  loading.value = true
  try {
    const payload = {
      attempt_id: currentAttemptId.value || null,
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
    submittedAttemptId.value = res.attempt_id || currentAttemptId.value || ''
    isSubmitted.value = true
    showResult.value = true
  } catch (err) {
    console.error(err)
    message.error(err.response?.data?.message || '提交测试失败')
  } finally {
    loading.value = false
  }
}

function getReviewAnswerByQuestionId(questionId) {
  return (reviewData.value.answers || []).find((item) => item.question_id === questionId) || {}
}

function reviewCorrectAnswer(q) {
  if (q.question_type === 'multi_choice') {
    const ids = q.correct_indices?.length ? q.correct_indices : [q.correct_index]
    return ids
      .map((i) => `${String.fromCharCode(65 + i)}${q.options?.[i] ? `. ${q.options[i]}` : ''}`)
      .join('；')
  }
  if (q.question_type === 'true_false') return q.correct_index === 0 ? '对' : '错'
  if (['fill_blank', 'short_answer'].includes(q.question_type)) return q.correct_answer || '无'
  const i = q.correct_index
  return `${String.fromCharCode(65 + i)}${q.options?.[i] ? `. ${q.options[i]}` : ''}`
}

function reviewUserAnswer(q) {
  const ans = getReviewAnswerByQuestionId(q.id)
  if (q.question_type === 'multi_choice') {
    let picks = []
    try {
      picks = JSON.parse(ans.answer_text || '[]')
      if (!Array.isArray(picks)) picks = []
    } catch (e) {
      picks = []
    }
    if (!picks.length && ans.selected_index !== undefined && ans.selected_index !== null)
      picks = [ans.selected_index]
    if (!picks.length) return '未作答'
    return picks
      .map((i) => `${String.fromCharCode(65 + i)}${q.options?.[i] ? `. ${q.options[i]}` : ''}`)
      .join('；')
  }
  if (q.question_type === 'true_false') {
    if (ans.selected_index === undefined || ans.selected_index === null) return '未作答'
    return ans.selected_index === 0 ? '对' : '错'
  }
  if (['fill_blank', 'short_answer'].includes(q.question_type))
    return (ans.answer_text || '').trim() || '未作答'
  if (ans.selected_index === undefined || ans.selected_index === null) return '未作答'
  return `${String.fromCharCode(65 + ans.selected_index)}${q.options?.[ans.selected_index] ? `. ${q.options[ans.selected_index]}` : ''}`
}

function reviewIsCorrect(q) {
  const ans = getReviewAnswerByQuestionId(q.id)
  if (q.question_type === 'multi_choice') {
    let picks = []
    try {
      picks = JSON.parse(ans.answer_text || '[]')
      if (!Array.isArray(picks)) picks = []
    } catch (e) {
      picks = []
    }
    if (!picks.length && ans.selected_index !== undefined && ans.selected_index !== null)
      picks = [ans.selected_index]
    const expected = new Set(q.correct_indices?.length ? q.correct_indices : [q.correct_index])
    return picks.length > 0 && picks.length === expected.size && picks.every((x) => expected.has(x))
  }
  if (q.question_type === 'true_false') {
    return (
      ans.selected_index !== undefined &&
      ans.selected_index !== null &&
      ans.selected_index === q.correct_index
    )
  }
  if (['fill_blank', 'short_answer'].includes(q.question_type)) {
    const userText = (ans.answer_text || '').toLowerCase().replace(/\s+/g, '')
    const correctText = (q.correct_answer || '').toLowerCase().replace(/\s+/g, '')
    return Boolean(userText) && userText === correctText
  }
  return (
    ans.selected_index !== undefined &&
    ans.selected_index !== null &&
    ans.selected_index === q.correct_index
  )
}

async function openAnswerReview() {
  const attemptId = submittedAttemptId.value || result.value?.attempt_id || currentAttemptId.value
  await openAnswerReviewByAttemptId(attemptId)
}

async function openAnswerReviewByAttemptId(attemptId) {
  if (!attemptId) {
    message.warning('该套试卷缺少 attempt 记录，无法查看解析')
    return
  }
  try {
    const token = localStorage.getItem('token')
    const res = await request.get(`/simulation-attempts/${attemptId}/review`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    reviewData.value = {
      questions: res.questions || [],
      answers: res.answers || []
    }
    showReview.value = true
  } catch (err) {
    console.error(err)
    message.error(err.response?.data?.message || '获取答案解析失败')
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
  showReview.value = false
  reviewData.value = { questions: [], answers: [] }
  timeLeft.value = 0
}

function backToConfig() {
  resetAll()
}

onBeforeUnmount(() => clearTimer())
loadCourses()
loadPractices()
</script>

<style scoped>
.simulation-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
  padding: 30px;
  border-radius: 6px;
  background: var(--mixm-surface-muted);
  border: 1px solid var(--mixm-border);
}
.hero-main {
  flex: 1;
  min-width: 0;
}
.hero-stats {
  display: flex;
  gap: 10px;
}
.stat-item {
  min-width: 76px;
  padding: 12px;
  border-radius: 4px;
  background: var(--mixm-surface);
  border: 1px solid var(--mixm-border);
  text-align: center;
}
.stat-item strong {
  font-family: var(--mixm-font-display);
  font-weight: 400;
  font-size: 26px;
  color: var(--mixm-ink);
}
.stat-label {
  display: block;
  color: var(--mixm-muted);
  font-size: 11px;
}
.eyebrow {
  font-size: 10px;
  color: var(--mixm-accent);
  letter-spacing: 0.15em;
  font-weight: 500;
}
.hero h1 {
  margin: 12px 0;
  font-family: var(--mixm-font-display);
  font-weight: 400;
  font-size: 34px;
  color: var(--mixm-ink);
  letter-spacing: -0.025em;
}
.hero p {
  margin: 0;
  color: var(--mixm-muted);
  font-size: 13px;
}
.config-card,
.question-card {
  border-radius: 6px;
  border: 1px solid var(--mixm-border);
  background: var(--mixm-surface);
}
.question-card {
  margin-bottom: 18px;
}
.question-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  width: 100%;
}
.stem {
  margin-bottom: 18px;
  font-size: 16px;
  font-weight: 500;
  color: var(--mixm-ink);
  line-height: 1.8;
  background: var(--mixm-paper);
  border-left: 2px solid var(--mixm-accent);
  padding: 14px 16px;
}
.knowledge {
  margin-top: 16px;
  color: var(--mixm-muted);
  font-size: 12px;
  background: var(--mixm-paper);
  border: 1px dashed var(--mixm-border);
  border-radius: 4px;
  padding: 10px 14px;
}
.actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 24px;
}
.test-area {
  display: grid;
  width: 100%;
  grid-template-columns: minmax(0, 1fr) minmax(240px, 27%);
  gap: 20px;
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
  border-radius: 6px;
}
.sheet-summary {
  color: var(--mixm-muted);
  font-size: 13px;
  margin-bottom: 14px;
}
.sheet-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}
.sheet-item {
  height: 36px;
  border-radius: 3px;
  border: 1px solid var(--mixm-border);
  background: var(--mixm-paper);
  color: var(--mixm-ink);
  font-weight: 500;
  cursor: pointer;
}
.sheet-item.answered {
  background: var(--mixm-accent-soft);
  border-color: var(--mixm-accent);
  color: var(--mixm-accent-pressed);
}
.sheet-item:hover {
  border-color: var(--mixm-accent);
}
.timer {
  margin-bottom: 16px;
}
.count-tip {
  color: var(--mixm-muted);
  font-size: 12px;
  margin-top: 6px;
}
.type-pills :deep(.n-checkbox) {
  padding: 8px 12px;
  background: var(--mixm-paper);
  border: 1px solid var(--mixm-border);
  border-radius: 4px;
  transition:
    background 0.2s ease,
    border-color 0.2s ease;
}
.type-pills :deep(.n-checkbox:hover) {
  border-color: var(--mixm-accent);
}
.type-pills :deep(.n-checkbox.n-checkbox--checked) {
  background: var(--mixm-accent-soft);
  border-color: var(--mixm-accent);
}
:deep(.config-card .n-card-header),
:deep(.question-card .n-card-header) {
  border-bottom: 1px solid var(--mixm-border);
}
:deep(.n-card-header__main) {
  color: var(--mixm-ink);
  font-weight: 500;
}
.result-list {
  background: var(--mixm-paper);
  border: 1px solid var(--mixm-border);
  border-radius: 4px;
  padding: 14px;
}
.result-row {
  color: var(--mixm-ink);
}
.analysis-item {
  border: 1px solid var(--mixm-border);
  border-radius: 4px;
  background: var(--mixm-paper);
  padding: 14px;
}
.analysis-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}
.analysis-line {
  color: var(--mixm-muted);
  font-size: 13px;
  line-height: 1.8;
}
@media (max-width: 900px) {
  .test-area {
    grid-template-columns: 1fr;
  }
  .answer-sheet {
    position: static;
  }
}
@media (max-width: 680px) {
  .hero {
    padding: 24px;
    flex-direction: column;
    align-items: flex-start;
  }
  .hero h1 {
    font-size: 29px;
  }
  .hero-stats {
    width: 100%;
  }
  .stat-item {
    flex: 1;
    min-width: 0;
  }
}
</style>
