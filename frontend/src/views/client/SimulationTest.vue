<template>
  <div class="simulation-page">
    <header class="simulation-heading">
      <div>
        <h1>模拟测试</h1>
        <p>
          {{
            started
              ? '专注作答，完成后查看成绩与参考答案。'
              : '按课程与题型配置一份测试，检验你的学习成果。'
          }}
        </p>
      </div>
      <div v-if="started" class="heading-actions">
        <n-tag :bordered="false" :type="isSubmitted ? 'success' : 'info'">
          {{ isSubmitted ? '已完成' : '答题中' }}
        </n-tag>
        <n-button v-if="isSubmitted" secondary @click="resetTest">返回配置</n-button>
      </div>
    </header>

    <n-alert v-if="optionsError && !started" type="warning" :bordered="false">
      {{ optionsError }}
      <n-button text class="retry-options" :loading="optionsLoading" @click="loadOptions">
        重新加载
      </n-button>
    </n-alert>
    <n-alert v-if="error" type="error" :bordered="false">{{ error }}</n-alert>

    <div v-if="!started" class="configuration-layout">
      <n-form :model="form" :disabled="generating" label-placement="top" class="configuration-main">
        <section class="configuration-panel">
          <div class="section-heading">
            <span class="section-number">01</span>
            <div>
              <h2>选择出题方式</h2>
              <p>从题库抽题、使用练习卷，或根据教学内容生成题目。</p>
            </div>
          </div>
          <div class="generation-modes" role="group" aria-label="出题方式">
            <button
              v-for="mode in generationModes"
              :key="mode.value"
              type="button"
              class="generation-mode"
              :class="{ selected: form.generation_mode === mode.value }"
              :aria-pressed="form.generation_mode === mode.value"
              :disabled="generating"
              @click="form.generation_mode = mode.value"
            >
              <n-icon :size="22"><component :is="mode.icon" /></n-icon>
              <strong>{{ mode.label }}</strong>
              <span>{{ mode.description }}</span>
            </button>
          </div>
        </section>
        <section class="configuration-panel">
          <div class="section-heading">
            <span class="section-number">02</span>
            <div>
              <h2>设置测试范围</h2>
              <p>选择课程与难度，也可以不限范围进行综合练习。</p>
            </div>
          </div>
          <div class="configuration-fields">
            <n-form-item label="课程范围">
              <n-select
                v-model:value="form.course_id"
                :options="courseOptions"
                :loading="optionsLoading"
                filterable
                clearable
                placeholder="全部课程"
              />
            </n-form-item>
            <n-form-item v-if="form.generation_mode === 'practice'" label="练习卷">
              <n-select
                v-model:value="form.practice_id"
                :options="practiceOptions"
                :loading="optionsLoading"
                filterable
                clearable
                placeholder="选择一份练习卷"
              />
            </n-form-item>
            <n-form-item label="题目数量">
              <n-input-number
                v-model:value="form.question_count"
                :min="1"
                :max="50"
                :disabled="useTypeCounts"
                placeholder="1～50"
                class="full-width"
              >
                <template #suffix>题</template>
              </n-input-number>
            </n-form-item>
            <n-form-item label="题目难度">
              <n-select
                v-model:value="form.difficulty"
                :options="difficultyOptions"
                clearable
                :placeholder="form.generation_mode === 'ai' ? '默认难度（3）' : '不限难度'"
              />
            </n-form-item>
            <n-form-item label="作答时限">
              <n-input-number
                v-model:value="form.time_limit_minutes"
                :min="1"
                :max="180"
                clearable
                placeholder="不限时"
                class="full-width"
              >
                <template #suffix>分钟</template>
              </n-input-number>
            </n-form-item>
          </div>
          <p class="field-note">题目数量最多 50 题；不设置作答时限时，可按自己的节奏完成。</p>
        </section>
        <section class="configuration-panel">
          <div class="section-heading">
            <span class="section-number">03</span>
            <div>
              <h2>选择题型</h2>
              <p>
                {{
                  useTypeCounts
                    ? '为需要练习的题型填写数量，总计最多 50 题。'
                    : '未选择题型时，将从所有可用题型中出题。'
                }}
              </p>
            </div>
          </div>
          <div v-if="form.generation_mode === 'random'" class="custom-counts-toggle">
            <span>按题型分配数量</span>
            <n-switch
              v-model:value="customCounts"
              :disabled="generating"
              aria-label="按题型分配数量"
            />
          </div>
          <div v-if="useTypeCounts" class="type-counts-grid">
            <n-form-item v-for="type in simulationTypes" :key="type.value" :label="type.label">
              <n-input-number
                v-model:value="form.type_counts[type.value]"
                :min="0"
                :max="50"
                class="full-width"
                :input-props="{ 'aria-label': `${type.label}数量` }"
              />
            </n-form-item>
          </div>
          <n-checkbox-group v-else v-model:value="form.question_types">
            <div class="question-type-grid">
              <n-checkbox
                v-for="type in simulationTypes"
                :key="type.value"
                :value="type.value"
                class="question-type-option"
              >
                {{ type.label }}
              </n-checkbox>
            </div>
          </n-checkbox-group>
          <p
            v-if="useTypeCounts"
            class="field-note"
            :class="{ 'invalid-count': typeCountTotal < 1 || typeCountTotal > 50 }"
          >
            已配置 {{ typeCountTotal }} 题{{
              typeCountTotal < 1 || typeCountTotal > 50
                ? '，请将总题数设置为 1～50 题。'
                : '，将按以上数量抽题。'
            }}
          </p>
        </section>
      </n-form>
      <aside class="configuration-summary">
        <section class="summary-card">
          <h2>本次测试</h2>
          <div class="summary-question-count">
            <strong>{{ requestedCount || '—' }}</strong>
            <span>道题</span>
          </div>
          <dl class="configuration-summary-list">
            <div>
              <dt>出题方式</dt>
              <dd>{{ currentMode.label }}</dd>
            </div>
            <div>
              <dt>课程范围</dt>
              <dd>{{ selectedCourseName }}</dd>
            </div>
            <div v-if="form.generation_mode === 'practice'">
              <dt>练习卷</dt>
              <dd>{{ selectedPracticeName }}</dd>
            </div>
            <div>
              <dt>题目难度</dt>
              <dd>{{ selectedDifficulty }}</dd>
            </div>
            <div>
              <dt>作答时限</dt>
              <dd>{{ form.time_limit_minutes ? `${form.time_limit_minutes} 分钟` : '不限时' }}</dd>
            </div>
          </dl>
          <div class="summary-types">
            <span>题型范围</span>
            <div>
              <n-tag v-for="type in selectedTypes" :key="type" size="small" :bordered="false">
                {{ simulationTypeLabels[type] }}
              </n-tag>
              <span v-if="!selectedTypes.length">全部题型</span>
            </div>
          </div>
          <n-button
            block
            type="primary"
            size="large"
            :loading="generating"
            :disabled="optionsLoading || !!configurationError"
            @click="startTest"
          >
            {{ generating ? '正在准备题目' : '开始测试' }}
          </n-button>
          <p class="summary-note">
            {{ configurationError || '题目数量以实际生成的试卷为准，完成后可查看成绩与参考答案。' }}
          </p>
        </section>
      </aside>
    </div>

    <template v-else-if="questions.length">
      <div class="paper-overview">
        <div>
          <span>{{ currentMode.label }}</span>
          <span>{{ selectedCourseName }}</span>
        </div>
        <div>
          <strong>{{ questions.length }}</strong>
          道题
          <span class="overview-divider">/</span>
          总分 {{ totalPaperScore }} 分
        </div>
      </div>
      <div class="test-layout">
        <div class="test-questions">
          <SimulationQuestion
            v-for="(question, index) in questions"
            :key="question.id"
            v-model:answer="answers[question.id]"
            :question="question"
            :index="index"
            :disabled="isSubmitted || submitting || timeExpired"
          />
          <div class="mobile-submit-bar">
            <span>已答 {{ answeredCount }} / {{ questions.length }} 题</span>
            <n-button
              type="primary"
              :loading="submitting"
              @click="isSubmitted ? (showResult = true) : submitTest()"
            >
              {{ isSubmitted ? '查看结果' : '提交测试' }}
            </n-button>
          </div>
        </div>
        <aside class="answer-sheet">
          <section class="summary-card answer-sheet-card">
            <div class="sheet-heading">
              <h2>答题卡</h2>
              <span>{{ questions.length }} 题</span>
            </div>
            <div
              v-if="timeLimitSeconds && !isSubmitted"
              class="time-remaining"
              :class="{ urgent: timeLeft <= 60 }"
            >
              <n-icon :size="18"><TimeOutline /></n-icon>
              <span>剩余时间</span>
              <strong>{{ timeLeftText }}</strong>
            </div>
            <div class="answer-progress" aria-live="polite">
              <span>答题进度</span>
              <strong>
                {{ answeredCount }}
                <small>/ {{ questions.length }}</small>
              </strong>
            </div>
            <n-progress
              type="line"
              :percentage="answerPercentage"
              :show-indicator="false"
              :height="5"
              color="var(--mixm-accent)"
            />
            <div class="sheet-grid">
              <button
                v-for="(question, index) in questions"
                :key="question.id"
                type="button"
                class="sheet-item"
                :class="{ answered: hasSimulationAnswer(question, answers[question.id]) }"
                :aria-label="`第 ${index + 1} 题，${hasSimulationAnswer(question, answers[question.id]) ? '已作答' : '未作答'}`"
                @click="scrollToQuestion(question.id)"
              >
                {{ index + 1 }}
              </button>
            </div>
            <div class="sheet-legend">
              <span>
                <i class="legend-answered"></i>
                已作答
              </span>
              <span>
                <i></i>
                未作答
              </span>
            </div>
            <template v-if="isSubmitted">
              <n-button block type="primary" @click="showResult = true">查看测试结果</n-button>
              <n-button
                block
                secondary
                :loading="reviewLoading"
                class="review-button"
                @click="openReview"
              >
                答案与解析
              </n-button>
            </template>
            <n-button
              v-else
              block
              type="primary"
              size="large"
              :loading="submitting"
              @click="submitTest"
            >
              提交测试
            </n-button>
            <p class="summary-note">
              {{
                isSubmitted
                  ? '测试已完成，答案已锁定。'
                  : timeExpired
                    ? '时间已到，答案已锁定。若提交失败，可重试提交。'
                    : answeredCount === questions.length
                      ? '全部题目已作答，可以提交测试。'
                      : `还有 ${questions.length - answeredCount} 题未作答，可通过题号快速定位。`
              }}
            </p>
          </section>
        </aside>
      </div>
    </template>
    <n-result
      v-else
      status="info"
      title="暂无题目"
      description="当前条件下没有可用题目，请调整测试配置后重试。"
    >
      <template #footer><n-button @click="resetTest">返回配置</n-button></template>
    </n-result>
    <SimulationResultsModal
      v-model:show="showResult"
      :result="result || {}"
      :review-loading="reviewLoading"
      :review-error="reviewError"
      @review="openReview"
    />
    <SimulationReviewModal
      v-model:show="showReview"
      :questions="reviewData.questions"
      :answers="reviewData.answers"
    />
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCheckbox,
  NCheckboxGroup,
  NForm,
  NFormItem,
  NIcon,
  NInputNumber,
  NProgress,
  NResult,
  NSelect,
  NSwitch,
  NTag
} from 'naive-ui'
import { BookOutline, DocumentTextOutline, SparklesOutline, TimeOutline } from '@vicons/ionicons5'
import { categoryApi, practiceApi, simulationApi } from '@/api'
import SimulationQuestion from '@/components/SimulationQuestion.vue'
import SimulationResultsModal from '@/components/SimulationResultsModal.vue'
import SimulationReviewModal from '@/components/SimulationReviewModal.vue'
import { readAllPages } from '@/utils/courseManagement'
import {
  hasSimulationAnswer,
  simulationAnswers,
  simulationTypeLabels,
  simulationTypes
} from '@/utils/simulation'

const generating = ref(false)
const submitting = ref(false)
const reviewLoading = ref(false)
const reviewError = ref('')
const optionsLoading = ref(false)
const optionsError = ref('')
const error = ref('')
const started = ref(false)
const isSubmitted = ref(false)
const questions = ref([])
const answers = ref({})
const attemptId = ref('')
const showResult = ref(false)
const showReview = ref(false)
const result = ref(null)
const reviewData = ref({ questions: [], answers: [] })
const courseOptions = ref([])
const practiceOptions = ref([])
const timeLeft = ref(0)
const timeLimitSeconds = ref(0)
const customCounts = ref(false)
let timer = null
let expiresAt = 0

const form = ref({
  course_id: null,
  practice_id: null,
  question_count: 10,
  difficulty: null,
  question_types: [],
  type_counts: Object.fromEntries(simulationTypes.map((type) => [type.value, 0])),
  time_limit_minutes: null,
  generation_mode: 'random'
})
const generationModes = [
  { value: 'random', label: '题库抽题', description: '从已审核题目中抽取', icon: BookOutline },
  {
    value: 'practice',
    label: '练习卷',
    description: '选择已配置的练习卷',
    icon: DocumentTextOutline
  },
  { value: 'ai', label: '智能体生成', description: '结合教学内容生成题目', icon: SparklesOutline }
]
const difficultyOptions = ['基础', '较易', '中等', '较难', '挑战'].map((label, index) => ({
  label: `${index + 1} · ${label}`,
  value: index + 1
}))
const currentMode = computed(() =>
  generationModes.find((mode) => mode.value === form.value.generation_mode)
)
const selectedCourseName = computed(
  () =>
    courseOptions.value.find((course) => course.value === form.value.course_id)?.label || '全部课程'
)
const selectedPracticeName = computed(
  () =>
    practiceOptions.value.find((practice) => practice.value === form.value.practice_id)?.label ||
    '尚未选择'
)
const selectedDifficulty = computed(
  () =>
    difficultyOptions.find((item) => item.value === form.value.difficulty)?.label ||
    (form.value.generation_mode === 'ai' ? '默认难度（3）' : '不限难度')
)
const useTypeCounts = computed(() => form.value.generation_mode === 'random' && customCounts.value)
const typeCountTotal = computed(() =>
  Object.values(form.value.type_counts).reduce((sum, value) => sum + (Number(value) || 0), 0)
)
const requestedCount = computed(() =>
  useTypeCounts.value ? typeCountTotal.value : form.value.question_count
)
const selectedTypes = computed(() =>
  useTypeCounts.value
    ? simulationTypes
        .filter((type) => form.value.type_counts[type.value] > 0)
        .map((type) => type.value)
    : form.value.question_types
)
const configurationError = computed(() => {
  if (form.value.generation_mode === 'practice' && !form.value.practice_id)
    return '请先选择一份练习卷。'
  if (
    useTypeCounts.value &&
    Object.values(form.value.type_counts).some(
      (count) => count != null && (!Number.isInteger(count) || count < 0 || count > 50)
    )
  )
    return '每种题型的数量应为 0～50 的整数。'
  if (
    !Number.isInteger(requestedCount.value) ||
    requestedCount.value < 1 ||
    requestedCount.value > 50
  )
    return '请将总题数设置为 1～50 题。'
  const minutes = form.value.time_limit_minutes
  if (minutes != null && (!Number.isInteger(minutes) || minutes < 1 || minutes > 180))
    return '作答时限应为 1～180 分钟，或留空不限时。'
  return ''
})
const answeredCount = computed(
  () =>
    questions.value.filter((question) => hasSimulationAnswer(question, answers.value[question.id]))
      .length
)
const answerPercentage = computed(() =>
  questions.value.length ? Math.round((answeredCount.value / questions.value.length) * 100) : 0
)
const totalPaperScore = computed(() =>
  questions.value.reduce((sum, question) => sum + (Number(question.score) || 0), 0).toFixed(2)
)
const timeExpired = computed(
  () => started.value && !isSubmitted.value && timeLimitSeconds.value > 0 && timeLeft.value <= 0
)
const timeLeftText = computed(
  () =>
    `${String(Math.floor(timeLeft.value / 60)).padStart(2, '0')}:${String(timeLeft.value % 60).padStart(2, '0')}`
)

// 课程选择与阅读目录保持一致，只展示启用的顶级课程。
async function loadOptions() {
  optionsLoading.value = true
  optionsError.value = ''
  const [courses, practices] = await Promise.allSettled([
    categoryApi.getTree(),
    readAllPages((params) => practiceApi.getList(params))
  ])
  const failed = []
  if (courses.status === 'fulfilled')
    courseOptions.value = courses.value
      .filter((course) => course.is_active !== false)
      .map((course) => ({ label: course.name, value: course.id }))
  else failed.push('课程')
  if (practices.status === 'fulfilled')
    practiceOptions.value = practices.value.map((practice) => ({
      label: `${practice.title}（${practice.question_count || 0} 题）`,
      value: practice.id
    }))
  else failed.push('练习卷')
  if (failed.length) optionsError.value = `${failed.join('、')}选项加载失败，请重新加载。`
  optionsLoading.value = false
}

async function startTest() {
  if (generating.value || started.value || configurationError.value) return
  generating.value = true
  error.value = ''
  try {
    const payload = {
      generation_mode: form.value.generation_mode,
      course_id: form.value.course_id || null,
      practice_id: form.value.generation_mode === 'practice' ? form.value.practice_id : null,
      question_count: requestedCount.value,
      difficulty: form.value.difficulty || null,
      time_limit_minutes: form.value.time_limit_minutes || null,
      question_types: selectedTypes.value.length ? [...selectedTypes.value] : null,
      type_counts: useTypeCounts.value
        ? Object.fromEntries(
            Object.entries(form.value.type_counts).filter(([, count]) => count > 0)
          )
        : null
    }
    const data = await simulationApi.generate(payload)
    questions.value = data.questions || []
    attemptId.value = data.attempt_id || ''
    answers.value = {}
    isSubmitted.value = data.attempt_status === 'submitted'
    started.value = true
    result.value = data.result || null
    clearTimer()
    timeLimitSeconds.value = (payload.time_limit_minutes || 0) * 60
    timeLeft.value = timeLimitSeconds.value
    if (questions.value.length && timeLimitSeconds.value && !isSubmitted.value) {
      expiresAt = Date.now() + timeLimitSeconds.value * 1000
      timer = setInterval(() => {
        timeLeft.value = Math.max(0, Math.ceil((expiresAt - Date.now()) / 1000))
        if (timeLeft.value === 0) {
          clearTimer()
          submitTest()
        }
      }, 1000)
    }
    await nextTick()
    document.getElementById('main-content')?.scrollTo({ top: 0 })
  } catch (cause) {
    error.value = cause.response?.data?.message || cause.message || '准备题目失败，请重试。'
  } finally {
    generating.value = false
  }
}

// 交卷期间锁定输入，失败后保留全部答案和剩余计时，允许直接重试。
async function submitTest() {
  if (!questions.value.length || isSubmitted.value || submitting.value) return
  submitting.value = true
  error.value = ''
  try {
    const data = await simulationApi.submit({
      attempt_id: attemptId.value,
      answers: simulationAnswers(questions.value, answers.value)
    })
    result.value = data
    attemptId.value = data.attempt_id || attemptId.value
    isSubmitted.value = true
    clearTimer()
    showResult.value = true
  } catch (cause) {
    error.value = cause.response?.data?.message || cause.message || '提交失败，请重试。'
  } finally {
    submitting.value = false
  }
}

async function openReview() {
  if (reviewLoading.value) return
  if (!attemptId.value) {
    error.value = '未找到本次测试记录，暂时无法查看答案。'
    return
  }
  reviewLoading.value = true
  reviewError.value = ''
  error.value = ''
  try {
    const data = await simulationApi.review(attemptId.value)
    reviewData.value = { questions: data.questions || [], answers: data.answers || [] }
    showResult.value = false
    showReview.value = true
  } catch (cause) {
    error.value = cause.response?.data?.message || cause.message || '答案加载失败，请重试。'
    reviewError.value = error.value
  } finally {
    reviewLoading.value = false
  }
}

function scrollToQuestion(id) {
  const element = document.getElementById(`question-${id}`)
  element?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  element?.focus({ preventScroll: true })
}
function clearTimer() {
  if (timer) clearInterval(timer)
  timer = null
}
function resetTest() {
  clearTimer()
  started.value = false
  isSubmitted.value = false
  questions.value = []
  answers.value = {}
  attemptId.value = ''
  result.value = null
  showResult.value = false
  showReview.value = false
  reviewData.value = { questions: [], answers: [] }
  timeLeft.value = 0
  timeLimitSeconds.value = 0
  reviewError.value = ''
  error.value = ''
}
onMounted(loadOptions)
onBeforeUnmount(clearTimer)
</script>

<style scoped>
.simulation-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.simulation-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
}
.simulation-heading h1 {
  margin: 0;
  font-family: var(--mixm-font-display);
  font-size: 30px;
  font-weight: 400;
}
.simulation-heading p {
  margin: 8px 0 0;
  color: var(--mixm-muted);
  font-size: 13px;
}
.heading-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.configuration-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 290px;
  align-items: start;
  gap: 24px;
}
.configuration-main {
  display: flex;
  flex-direction: column;
  gap: 18px;
  min-width: 0;
}
.configuration-panel {
  padding: 24px;
  background: var(--mixm-surface);
  border: 1px solid var(--mixm-border);
  border-radius: 10px;
}
.section-heading {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 22px;
}
.section-number {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border: 1px solid var(--mixm-border);
  border-radius: 50%;
  color: var(--mixm-accent);
  font-size: 12px;
}
.section-heading h2,
.summary-card h2 {
  margin: 0;
  font-family: var(--mixm-font-body);
  font-size: 15px;
  font-weight: 600;
}
.section-heading p {
  margin: 4px 0 0;
  color: var(--mixm-muted);
  font-size: 12px;
  line-height: 1.7;
}
.generation-modes {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}
.generation-mode {
  display: flex;
  align-items: flex-start;
  flex-direction: column;
  gap: 8px;
  padding: 18px 16px;
  font: inherit;
  text-align: left;
  color: var(--mixm-muted);
  background: var(--mixm-paper);
  border: 1px solid var(--mixm-border);
  border-radius: 7px;
  cursor: pointer;
  transition:
    background-color 0.2s,
    border-color 0.2s;
}
.generation-mode strong {
  color: var(--mixm-ink);
  font-size: 14px;
  font-weight: 500;
}
.generation-mode > span {
  font-size: 12px;
  line-height: 1.6;
}
.generation-mode.selected {
  border-color: var(--mixm-accent);
  background: var(--mixm-accent-soft);
  color: var(--mixm-accent);
}
.generation-mode:focus-visible,
.sheet-item:focus-visible {
  outline: 2px solid var(--mixm-accent);
  outline-offset: 3px;
}
.generation-mode:disabled {
  cursor: wait;
  opacity: 0.65;
}
.configuration-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 20px;
}
.configuration-fields :deep(.n-form-item),
.type-counts-grid :deep(.n-form-item) {
  min-width: 0;
}
.full-width {
  width: 100%;
}
.field-note {
  margin: 0;
  color: var(--mixm-muted);
  font-size: 12px;
  line-height: 1.8;
}
.custom-counts-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
  color: var(--mixm-muted);
  font-size: 13px;
}
.question-type-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}
.question-type-option {
  padding: 12px;
  border: 1px solid var(--mixm-border);
  border-radius: 6px;
  background: var(--mixm-paper);
}
.question-type-option.n-checkbox--checked {
  border-color: var(--mixm-accent);
  background: var(--mixm-accent-soft);
}
.type-counts-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0 16px;
}
.invalid-count {
  color: var(--mixm-error);
}
.configuration-summary,
.answer-sheet {
  position: sticky;
  top: 0;
  min-width: 0;
}
.summary-card {
  padding: 24px;
  border: 1px solid var(--mixm-border);
  border-radius: 10px;
  background: var(--mixm-surface);
}
.summary-question-count {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin: 24px 0;
  color: var(--mixm-muted);
}
.summary-question-count strong {
  font-family: var(--mixm-font-display);
  font-size: 48px;
  font-weight: 400;
  line-height: 1;
  color: var(--mixm-ink);
}
.summary-question-count > span {
  font-size: 13px;
}
.configuration-summary-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin: 0;
  padding: 20px 0;
  border-top: 1px solid var(--mixm-border);
  border-bottom: 1px solid var(--mixm-border);
}
.configuration-summary-list > div {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 18px;
  font-size: 12px;
}
.configuration-summary-list dt {
  flex-shrink: 0;
  color: var(--mixm-muted);
}
.configuration-summary-list dd {
  margin: 0;
  text-align: right;
  overflow-wrap: anywhere;
}
.summary-types {
  margin: 20px 0 24px;
}
.summary-types > span {
  color: var(--mixm-muted);
  font-size: 12px;
}
.summary-types > div {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
  font-size: 13px;
}
.summary-note {
  margin: 12px 0 0;
  color: var(--mixm-muted);
  font-size: 12px;
  line-height: 1.8;
}
.retry-options {
  margin-left: 12px;
}
.paper-overview {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px 20px;
  border: 1px solid var(--mixm-border);
  border-radius: 8px;
  background: var(--mixm-surface);
  font-size: 13px;
}
.paper-overview > div:first-child {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 20px;
  color: var(--mixm-muted);
}
.paper-overview > div:last-child {
  white-space: nowrap;
}
.paper-overview strong {
  font-weight: 600;
}
.overview-divider {
  margin: 0 12px;
  color: var(--mixm-border);
}
.test-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 270px;
  grid-template-areas: 'questions sheet';
  align-items: start;
  gap: 24px;
}
.test-questions {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
  grid-area: questions;
}
.answer-sheet {
  grid-area: sheet;
}
.sheet-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sheet-heading > span {
  font-size: 12px;
  color: var(--mixm-muted);
}
.time-remaining {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  margin-top: 20px;
  background: var(--mixm-paper);
  border-radius: 6px;
  color: var(--mixm-muted);
  font-size: 12px;
}
.time-remaining strong {
  margin-left: auto;
  font-size: 18px;
  font-weight: 500;
  font-variant-numeric: tabular-nums;
  color: var(--mixm-ink);
}
.time-remaining.urgent strong {
  color: var(--mixm-error);
}
.answer-progress {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 24px 0 10px;
  color: var(--mixm-muted);
  font-size: 12px;
}
.answer-progress strong {
  font-size: 20px;
  color: var(--mixm-ink);
  font-weight: 500;
}
.answer-progress small {
  margin-left: 5px;
  color: var(--mixm-muted);
  font-size: 12px;
  font-weight: 400;
}
.sheet-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
  margin-top: 20px;
  max-height: 280px;
  overflow-y: auto;
  padding: 3px;
}
.sheet-item {
  height: 36px;
  font: inherit;
  font-size: 12px;
  color: var(--mixm-muted);
  background: var(--mixm-paper);
  border: 1px solid var(--mixm-border);
  border-radius: 5px;
  cursor: pointer;
}
.sheet-item.answered {
  border-color: var(--mixm-accent);
  background: var(--mixm-accent-soft);
  color: var(--mixm-accent);
}
.sheet-item:hover {
  border-color: var(--mixm-accent);
}
.sheet-legend {
  display: flex;
  gap: 16px;
  margin: 16px 0 24px;
  color: var(--mixm-muted);
  font-size: 11px;
}
.sheet-legend > span {
  display: flex;
  align-items: center;
  gap: 6px;
}
.sheet-legend i {
  width: 8px;
  height: 8px;
  border: 1px solid var(--mixm-border);
  border-radius: 2px;
  background: var(--mixm-paper);
}
.sheet-legend .legend-answered {
  background: var(--mixm-accent-soft);
  border-color: var(--mixm-accent);
}
.review-button {
  margin-top: 10px;
}
.mobile-submit-bar {
  display: none;
}
@media (max-width: 1100px) {
  .configuration-layout {
    grid-template-columns: minmax(0, 1fr) 250px;
    gap: 18px;
  }
  .configuration-panel,
  .summary-card {
    padding: 20px;
  }
  .generation-modes {
    gap: 8px;
  }
  .generation-mode {
    padding: 14px 12px;
  }
}
@media (max-width: 900px) {
  .configuration-layout,
  .test-layout {
    grid-template-columns: minmax(0, 1fr);
  }
  .configuration-summary,
  .answer-sheet {
    position: static;
  }
  .test-layout {
    grid-template-areas: 'sheet' 'questions';
  }
  .sheet-grid {
    grid-template-columns: repeat(10, minmax(0, 1fr));
  }
  .mobile-submit-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 18px;
    background: var(--mixm-surface);
    border: 1px solid var(--mixm-border);
    border-radius: 8px;
    font-size: 12px;
    color: var(--mixm-muted);
  }
}
@media (max-width: 600px) {
  .simulation-page {
    gap: 18px;
  }
  .simulation-heading h1 {
    font-size: 26px;
  }
  .configuration-fields {
    grid-template-columns: minmax(0, 1fr);
  }
  .generation-modes {
    grid-template-columns: minmax(0, 1fr);
  }
  .generation-mode {
    display: grid;
    grid-template-columns: 26px minmax(0, 1fr);
    gap: 4px 10px;
  }
  .generation-mode > .n-icon {
    grid-row: span 2;
    align-self: center;
  }
  .generation-mode > span {
    grid-column: 2;
  }
  .question-type-grid,
  .type-counts-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .sheet-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
  .configuration-panel,
  .summary-card {
    padding: 20px 18px;
  }
  .question-type-option {
    padding: 10px;
  }
}
</style>
