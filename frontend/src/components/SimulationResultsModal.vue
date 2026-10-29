<template>
  <n-modal
    :show="show"
    preset="card"
    title="测试结果"
    class="simulation-results-modal"
    @update:show="emit('update:show', $event)"
  >
    <n-alert v-if="reviewError" type="error" :bordered="false" class="review-error">
      {{ reviewError }}
    </n-alert>
    <div class="result-overview">
      <div class="result-score">
        <span>本次得分</span>
        <div>
          <strong>{{ number(result.score) }}</strong>
          <span>/ {{ number(result.total_score ?? 100) }}</span>
        </div>
      </div>
      <div class="result-metrics">
        <div>
          <span>正确率</span>
          <strong>
            {{ number(result.accuracy) }}
            <small>%</small>
          </strong>
        </div>
        <div>
          <span>答对题数</span>
          <strong>
            {{ result.correct_count ?? 0 }}
            <small>/ {{ result.total_questions ?? 0 }}</small>
          </strong>
        </div>
        <div>
          <span>已答题数</span>
          <strong>
            {{ result.answered_questions ?? 0 }}
            <small>/ {{ result.total_questions ?? 0 }}</small>
          </strong>
        </div>
      </div>
    </div>
    <section class="knowledge-results">
      <h3>知识点掌握情况</h3>
      <div
        v-for="(accuracy, name) in result.knowledge_accuracy || {}"
        :key="name"
        class="knowledge-result"
      >
        <div>
          <span>{{ name }}</span>
          <strong>{{ number(accuracy) }}%</strong>
        </div>
        <n-progress
          type="line"
          :percentage="Number(accuracy) || 0"
          :show-indicator="false"
          :color="Number(accuracy) < 60 ? 'var(--mixm-warning)' : 'var(--mixm-success)'"
          :height="6"
        />
      </div>
      <p v-if="!Object.keys(result.knowledge_accuracy || {}).length" class="result-empty">
        本次测试未标注知识点。
      </p>
    </section>
    <section v-if="result.weak_knowledge_points?.length" class="weak-points">
      <h3>建议复习</h3>
      <n-space :size="8">
        <n-tag
          v-for="point in result.weak_knowledge_points"
          :key="point"
          type="warning"
          :bordered="false"
          size="small"
        >
          {{ point }}
        </n-tag>
      </n-space>
    </section>
    <template #action>
      <n-space justify="end">
        <n-button @click="emit('update:show', false)">返回试卷</n-button>
        <n-button type="primary" :loading="reviewLoading" @click="emit('review')">
          查看答案与解析
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup>
import { NAlert, NButton, NModal, NProgress, NSpace, NTag } from 'naive-ui'
defineProps({
  show: Boolean,
  result: { type: Object, default: () => ({}) },
  reviewLoading: Boolean,
  reviewError: { type: String, default: '' }
})
const emit = defineEmits(['update:show', 'review'])
function number(value) {
  return (Number(value) || 0).toFixed(2)
}
</script>

<style scoped>
.review-error {
  margin-bottom: 18px;
}
.simulation-results-modal {
  width: min(760px, calc(100vw - 32px));
}
.simulation-results-modal :deep(.n-card__content) {
  max-height: 65dvh;
  overflow-y: auto;
}
.result-overview {
  display: grid;
  grid-template-columns: 1fr 1.3fr;
  gap: 28px;
  padding: 24px;
  border: 1px solid var(--mixm-border);
  border-radius: 8px;
  background: var(--mixm-paper);
}
.result-score > span,
.result-metrics > div > span {
  color: var(--mixm-muted);
  font-size: 12px;
}
.result-score > div {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-top: 12px;
}
.result-score strong {
  font-family: var(--mixm-font-display);
  font-weight: 400;
  font-size: 48px;
  line-height: 1.2;
  color: var(--mixm-accent);
}
.result-score > div > span {
  color: var(--mixm-muted);
  font-size: 12px;
}
.result-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-content: center;
  gap: 16px;
}
.result-metrics > div {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.result-metrics strong {
  font-size: 23px;
  font-weight: 500;
  white-space: nowrap;
}
.result-metrics small {
  color: var(--mixm-muted);
  font-size: 12px;
  font-weight: 400;
  margin-left: 3px;
}
.knowledge-results,
.weak-points {
  margin-top: 26px;
}
.simulation-results-modal h3 {
  margin: 0 0 16px;
  font-family: var(--mixm-font-body);
  font-size: 14px;
  font-weight: 600;
}
.knowledge-result + .knowledge-result {
  margin-top: 18px;
}
.knowledge-result > div:first-child {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  font-size: 13px;
  margin-bottom: 8px;
}
.knowledge-result > div:first-child > span {
  overflow-wrap: anywhere;
}
.knowledge-result strong {
  font-size: 12px;
  font-weight: 500;
  flex-shrink: 0;
}
.result-empty {
  color: var(--mixm-muted);
  font-size: 13px;
}
@media (max-width: 650px) {
  .result-overview {
    grid-template-columns: minmax(0, 1fr);
    padding: 20px;
    gap: 22px;
  }
  .result-metrics {
    border-top: 1px solid var(--mixm-border);
    padding-top: 20px;
  }
  .result-metrics strong {
    font-size: 20px;
  }
}
</style>
