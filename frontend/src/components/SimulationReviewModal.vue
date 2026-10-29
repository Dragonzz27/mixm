<template>
  <n-modal
    :show="show"
    preset="card"
    title="答案与解析"
    class="simulation-review-modal"
    @update:show="emit('update:show', $event)"
  >
    <div class="review-summary">共 {{ questions.length }} 题，结合参考答案检查你的作答。</div>
    <section v-for="(question, index) in questions" :key="question.id" class="review-question">
      <header>
        <div>
          <strong>第 {{ index + 1 }} 题</strong>
          <span>{{ simulationTypeLabels[question.question_type] }}</span>
        </div>
        <n-tag
          :type="simulationIsCorrect(question, answerMap[question.id]) ? 'success' : 'error'"
          size="small"
          :bordered="false"
        >
          {{ simulationIsCorrect(question, answerMap[question.id]) ? '答对' : '答错' }}
        </n-tag>
      </header>
      <p class="review-stem">{{ question.stem }}</p>
      <div class="review-answers">
        <div :class="{ incorrect: !simulationIsCorrect(question, answerMap[question.id]) }">
          <span>你的答案</span>
          <p>{{ simulationUserAnswer(question, answerMap[question.id]) }}</p>
        </div>
        <div class="reference-answer">
          <span>参考答案</span>
          <p>{{ simulationCorrectAnswer(question) }}</p>
        </div>
      </div>
      <div v-if="question.explanation" class="review-explanation">
        <strong>解析</strong>
        <p>{{ question.explanation }}</p>
      </div>
      <footer>知识点 · {{ question.knowledge_tag || '未标注' }}</footer>
    </section>
    <template #action><n-button @click="emit('update:show', false)">返回试卷</n-button></template>
  </n-modal>
</template>

<script setup>
import { computed } from 'vue'
import { NButton, NModal, NTag } from 'naive-ui'
import {
  simulationCorrectAnswer,
  simulationIsCorrect,
  simulationTypeLabels,
  simulationUserAnswer
} from '@/utils/simulation'
const props = defineProps({
  show: Boolean,
  questions: { type: Array, default: () => [] },
  answers: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:show'])
const answerMap = computed(() =>
  Object.fromEntries(props.answers.map((answer) => [answer.question_id, answer]))
)
</script>

<style scoped>
.simulation-review-modal {
  width: min(900px, calc(100vw - 32px));
}
.simulation-review-modal :deep(.n-card__content) {
  max-height: 65dvh;
  overflow-y: auto;
}
.review-summary {
  margin-bottom: 20px;
  color: var(--mixm-muted);
  font-size: 13px;
}
.review-question {
  padding: 20px;
  border: 1px solid var(--mixm-border);
  border-radius: 8px;
}
.review-question + .review-question {
  margin-top: 16px;
}
.review-question header,
.review-question header > div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.review-question header strong {
  font-size: 14px;
  font-weight: 600;
}
.review-question header > div > span {
  font-size: 12px;
  color: var(--mixm-muted);
}
.review-stem {
  margin: 16px 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  line-height: 1.8;
}
.review-answers {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}
.review-answers > div {
  padding: 14px;
  background: var(--mixm-paper);
  border-radius: 6px;
}
.review-answers > div > span {
  font-size: 12px;
  color: var(--mixm-muted);
}
.review-answers p,
.review-explanation p {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.8;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.review-answers > .incorrect p {
  color: var(--mixm-error);
}
.review-answers > .reference-answer {
  background: var(--mixm-success-soft);
}
.review-explanation {
  margin-top: 18px;
}
.review-explanation strong {
  font-size: 13px;
  font-weight: 500;
}
.review-question footer {
  margin-top: 18px;
  color: var(--mixm-muted);
  font-size: 12px;
  overflow-wrap: anywhere;
}
@media (max-width: 600px) {
  .review-question {
    padding: 16px;
  }
  .review-answers {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
