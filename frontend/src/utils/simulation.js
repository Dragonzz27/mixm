export const simulationTypes = [
  { value: 'single_choice', label: '单选题' },
  { value: 'multi_choice', label: '多选题' },
  { value: 'true_false', label: '判断题' },
  { value: 'fill_blank', label: '填空题' },
  { value: 'short_answer', label: '简答题' }
]

export const simulationTypeLabels = Object.fromEntries(
  simulationTypes.map((type) => [type.value, type.label])
)
const textTypes = new Set(['fill_blank', 'short_answer'])

// 答题卡和进度使用同一规则，判断题的 0 与多选题都计入已答题数。
export function hasSimulationAnswer(question, answer) {
  if (textTypes.has(question.question_type)) return !!String(answer ?? '').trim()
  if (question.question_type === 'multi_choice') return Array.isArray(answer) && answer.length > 0
  return Number.isInteger(answer) && answer >= 0
}

export function simulationAnswers(questions, answers) {
  return questions.map((question) => {
    const answer = answers[question.id]
    const multi = question.question_type === 'multi_choice'
    const text = textTypes.has(question.question_type)
    return {
      question_id: question.id,
      selected_index: text ? null : multi ? (answer?.[0] ?? null) : (answer ?? null),
      answer_text: text ? String(answer ?? '').trim() : multi ? JSON.stringify(answer || []) : null
    }
  })
}

function selectedIndices(answer) {
  try {
    const picks = JSON.parse(answer.answer_text || '[]')
    if (Array.isArray(picks) && picks.length) return picks
  } catch {
    /* 旧记录可能只保存了 selected_index。 */
  }
  return answer.selected_index == null ? [] : [answer.selected_index]
}

function correctIndices(question) {
  return question.correct_indices?.length
    ? question.correct_indices
    : question.correct_index == null
      ? []
      : [question.correct_index]
}

function optionText(question, index) {
  return `${String.fromCharCode(65 + index)}${question.options?.[index] ? `. ${question.options[index]}` : ''}`
}

export function simulationUserAnswer(question, answer = {}) {
  if (textTypes.has(question.question_type)) return (answer.answer_text || '').trim() || '未作答'
  if (question.question_type === 'multi_choice') {
    const picks = selectedIndices(answer)
    return picks.length ? picks.map((index) => optionText(question, index)).join('；') : '未作答'
  }
  if (answer.selected_index == null) return '未作答'
  if (question.question_type === 'true_false') return answer.selected_index === 0 ? '正确' : '错误'
  return optionText(question, answer.selected_index)
}

export function simulationCorrectAnswer(question) {
  if (textTypes.has(question.question_type)) return question.correct_answer || '未提供参考答案'
  if (question.question_type === 'multi_choice')
    return (
      correctIndices(question)
        .map((index) => optionText(question, index))
        .join('；') || '未提供参考答案'
    )
  if (question.correct_index == null) return '未提供参考答案'
  if (question.question_type === 'true_false') return question.correct_index === 0 ? '正确' : '错误'
  return optionText(question, question.correct_index)
}

export function simulationIsCorrect(question, answer = {}) {
  if (textTypes.has(question.question_type)) {
    const normalize = (value) =>
      String(value || '')
        .toLowerCase()
        .replace(/\s+/g, '')
    const text = normalize(answer.answer_text)
    return !!text && text === normalize(question.correct_answer)
  }
  if (question.question_type === 'multi_choice') {
    const picks = selectedIndices(answer)
    const expected = new Set(correctIndices(question))
    return (
      picks.length > 0 &&
      picks.length === expected.size &&
      picks.every((index) => expected.has(index))
    )
  }
  return answer.selected_index != null && answer.selected_index === question.correct_index
}
