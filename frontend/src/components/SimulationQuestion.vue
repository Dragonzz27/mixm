<template>
  <article
    :id="`question-${question.id}`"
    class="simulation-question"
    tabindex="-1"
    :aria-label="`第 ${index + 1} 题`"
  >
    <header class="question-heading">
      <div>
        <span class="question-number">{{ String(index + 1).padStart(2, '0') }}</span>
        <n-tag size="small" :bordered="false">
          {{ simulationTypeLabels[question.question_type] || question.question_type }}
        </n-tag>
      </div>
      <span class="question-score">{{ Number(question.score || 0).toFixed(2) }} 分</span>
    </header>
    <p class="question-stem">{{ question.stem }}</p>
    <n-checkbox-group
      v-if="question.question_type === 'multi_choice'"
      :value="answer || []"
      :disabled="disabled"
      @update:value="emit('update:answer', $event)"
    >
      <div class="answer-options">
        <n-checkbox
          v-for="(option, optionIndex) in question.options"
          :key="optionIndex"
          :value="optionIndex"
          class="answer-option"
          :class="{ selected: answer?.includes(optionIndex) }"
        >
          <span class="option-letter">{{ String.fromCharCode(65 + optionIndex) }}.</span>
          {{ option }}
        </n-checkbox>
      </div>
    </n-checkbox-group>
    <n-radio-group
      v-else-if="question.question_type === 'true_false'"
      :value="answer"
      :disabled="disabled"
      class="judge-options"
      @update:value="emit('update:answer', $event)"
    >
      <n-radio
        v-for="(option, optionIndex) in ['正确', '错误']"
        :key="optionIndex"
        :value="optionIndex"
        class="answer-option"
        :class="{ selected: answer === optionIndex }"
      >
        {{ option }}
      </n-radio>
    </n-radio-group>
    <n-input
      v-else-if="['fill_blank', 'short_answer'].includes(question.question_type)"
      :value="answer || ''"
      :disabled="disabled"
      type="textarea"
      :autosize="{ minRows: 3, maxRows: 8 }"
      :placeholder="question.question_type === 'fill_blank' ? '填写你的答案' : '写下你的思路与答案'"
      :aria-label="`第 ${index + 1} 题答案`"
      @update:value="emit('update:answer', $event)"
    />
    <n-radio-group
      v-else
      :value="answer"
      :disabled="disabled"
      class="single-options"
      @update:value="emit('update:answer', $event)"
    >
      <div class="answer-options">
        <n-radio
          v-for="(option, optionIndex) in question.options"
          :key="optionIndex"
          :value="optionIndex"
          class="answer-option"
          :class="{ selected: answer === optionIndex }"
        >
          <span class="option-letter">{{ String.fromCharCode(65 + optionIndex) }}.</span>
          {{ option }}
        </n-radio>
      </div>
    </n-radio-group>
    <footer class="question-meta">
      <span>知识点 · {{ question.knowledge_tag || '未标注' }}</span>
      <span v-if="question.difficulty">难度 {{ question.difficulty }} / 5</span>
    </footer>
  </article>
</template>

<script setup>
import { NCheckbox, NCheckboxGroup, NInput, NRadio, NRadioGroup, NTag } from 'naive-ui'
import { simulationTypeLabels } from '@/utils/simulation'
defineProps({
  question: { type: Object, required: true },
  index: { type: Number, required: true },
  answer: { type: [Number, String, Array], default: undefined },
  disabled: Boolean
})
const emit = defineEmits(['update:answer'])
</script>

<style scoped>
.simulation-question {
  padding: 26px 28px;
  border: 1px solid var(--mixm-border);
  border-radius: 10px;
  background: var(--mixm-surface);
  scroll-margin-top: 20px;
  outline: none;
}
.simulation-question:focus-visible {
  outline: 2px solid var(--mixm-accent);
  outline-offset: 3px;
}
.question-heading,
.question-heading > div {
  display: flex;
  align-items: center;
  gap: 12px;
}
.question-heading {
  justify-content: space-between;
}
.question-number {
  color: var(--mixm-accent);
  font-size: 20px;
  font-family: var(--mixm-font-display);
}
.question-score {
  color: var(--mixm-muted);
  font-size: 12px;
}
.question-stem {
  margin: 20px 0;
  font-size: 16px;
  line-height: 1.9;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.answer-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.single-options {
  display: block;
}
.answer-option {
  width: 100%;
  box-sizing: border-box;
  margin: 0;
  padding: 12px 14px;
  border: 1px solid var(--mixm-border);
  border-radius: 6px;
  background: var(--mixm-paper);
  transition:
    background-color 0.2s,
    border-color 0.2s;
}
.answer-option:hover {
  border-color: var(--mixm-accent);
}
.answer-option.selected {
  border-color: var(--mixm-accent);
  background: var(--mixm-accent-soft);
}
.answer-option :deep(.n-radio__label),
.answer-option :deep(.n-checkbox__label) {
  white-space: normal;
  overflow-wrap: anywhere;
  line-height: 1.7;
}
.option-letter {
  margin-right: 8px;
  color: var(--mixm-muted);
}
.judge-options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}
.question-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 20px;
  padding-top: 18px;
  margin-top: 22px;
  border-top: 1px solid var(--mixm-border);
  color: var(--mixm-muted);
  font-size: 12px;
  overflow-wrap: anywhere;
}
@media (max-width: 600px) {
  .simulation-question {
    padding: 20px 18px;
  }
  .question-stem {
    font-size: 15px;
  }
}
</style>
