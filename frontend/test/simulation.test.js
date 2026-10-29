import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import vm from 'node:vm'
import { readAllPages } from '../src/utils/courseManagement.js'
import { hasSimulationAnswer, simulationAnswers, simulationTypeLabels, simulationTypes } from '../src/utils/simulation.js'

test('simulation supports generation, submission and answer review with scoring alone', async () => {
  const calls = []
  const errors = []
  const question = { id: 'question', question_type: 'single_choice', correct_index: 0, score: 100 }
  const submitted = { attempt_id: 'attempt', score: 100, accuracy: 100, knowledge_accuracy: { Go: 100 } }
  const source = fs.readFileSync(new URL('../src/views/client/SimulationTest.vue', import.meta.url), 'utf8')
    .match(/<script setup>([\s\S]*?)<\/script>/)[1]
    .replace(/^import[\s\S]*?from ['"][^'"]+['"]\s*$/gm, '')
  const page = vm.runInNewContext(`${source}
    ;({startTest, submitTest, openReview, resetTest,
      answers, result, started, showResult, showReview, reviewData, error})`, {
    ref: value => ({ value }),
    computed: getter => ({ get value() { return getter() } }),
    onBeforeUnmount() {},
    onMounted() {},
    nextTick: async () => {},
    document: { getElementById: () => null },
    BookOutline: {}, DocumentTextOutline: {}, SparklesOutline: {}, TimeOutline: {},
    simulationTypes, simulationTypeLabels, simulationAnswers, hasSimulationAnswer, readAllPages,
    useMessage: () => ({ warning: value => errors.push(value), error: value => errors.push(value) }),
    console: { error: value => errors.push(value) },
    localStorage: { getItem: () => 'token' },
    categoryApi: { getTree: async () => [] },
    practiceApi: { getList: async () => ({ items: [] }) },
    simulationApi: {
      generate: async () => {
        calls.push('generate')
        return { attempt_id: 'attempt', questions: [question] }
      },
      submit: async payload => {
        calls.push(JSON.parse(JSON.stringify(payload)))
        return submitted
      },
      review: async attemptId => {
        calls.push(`/simulation-attempts/${attemptId}/review`)
        return { questions: [question], answers: [{ question_id: 'question', selected_index: 0 }] }
      }
    }
  })

  await page.startTest()
  assert.equal(page.started.value, true)
  page.answers.value.question = 0
  await page.submitTest()
  assert.equal(page.showResult.value, true)
  assert.deepEqual(page.result.value, submitted)
  await page.submitTest()
  await page.openReview()
  assert.equal(page.showReview.value, true)
  assert.equal(page.reviewData.value.questions.length, 1)
  assert.deepEqual(calls, [
    'generate',
    { attempt_id: 'attempt', answers: [{ question_id: 'question', selected_index: 0, answer_text: null }] },
    '/simulation-attempts/attempt/review'
  ])
  assert.deepEqual(errors, [])
  assert.equal(page.error.value, '')
  page.resetTest()
  assert.equal(page.started.value, false)
  assert.equal(page.result.value, null)
})
