import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import vm from 'node:vm'
import {
  hasSimulationAnswer,
  simulationAnswers,
  simulationTypeLabels,
  simulationTypes
} from '../src/utils/simulation.js'

test('failed submission retains answers and countdown; expiry submits once and successful submission stops the timer', async () => {
  let now = 0
  let interval = null
  let completeSubmission
  const submissions = []
  const question = { id: 'one', question_type: 'single_choice', score: 100 }
  const source = fs
    .readFileSync(new URL('../src/views/client/SimulationTest.vue', import.meta.url), 'utf8')
    .match(/<script setup>([\s\S]*?)<\/script>/)[1]
    .replace(/^import[\s\S]*?from ['"][^'"]+['"]\s*$/gm, '')
  const page = vm.runInNewContext(
    `${source}
    ;({ startTest, submitTest, form, answers, submitting, isSubmitted, timeLeft, timeExpired, error, showResult })`,
    {
      ref: (value) => ({ value }),
      computed: (getter) => ({
        get value() {
          return getter()
        }
      }),
      onMounted() {},
      onBeforeUnmount() {},
      nextTick: async () => {},
      document: { getElementById: () => null },
      BookOutline: {},
      DocumentTextOutline: {},
      SparklesOutline: {},
      TimeOutline: {},
      hasSimulationAnswer,
      simulationAnswers,
      simulationTypeLabels,
      simulationTypes,
      Date: { now: () => now },
      setInterval(callback) {
        interval = callback
        return 1
      },
      clearInterval() {
        interval = null
      },
      simulationApi: {
        async generate() {
          return { attempt_id: 'attempt', questions: [question] }
        },
        submit(payload) {
          submissions.push(JSON.parse(JSON.stringify(payload)))
          return new Promise((resolve, reject) => {
            completeSubmission = { resolve, reject }
          })
        }
      }
    }
  )
  page.form.value.time_limit_minutes = 1
  await page.startTest()
  assert.equal(page.timeLeft.value, 60)
  page.answers.value.one = 0
  now = 10000
  interval()
  assert.equal(page.timeLeft.value, 50)

  const pending = page.submitTest()
  await page.submitTest()
  assert.equal(submissions.length, 1)
  assert.equal(page.submitting.value, true)
  now = 11000
  interval()
  assert.equal(page.timeLeft.value, 49)
  completeSubmission.reject(new Error('暂时失败'))
  await pending
  assert.equal(page.error.value, '暂时失败')
  assert.equal(page.answers.value.one, 0)
  assert.equal(page.isSubmitted.value, false)
  assert.equal(page.submitting.value, false)
  assert.equal(typeof interval, 'function')

  now = 61000
  interval()
  assert.equal(page.timeLeft.value, 0)
  assert.equal(page.timeExpired.value, true)
  assert.equal(submissions.length, 2)
  assert.equal(interval, null)
  await page.submitTest()
  assert.equal(submissions.length, 2)
  completeSubmission.reject(new Error('自动交卷暂时失败'))
  await new Promise((resolve) => setImmediate(resolve))
  assert.equal(page.timeExpired.value, true)
  assert.equal(page.isSubmitted.value, false)
  assert.equal(page.submitting.value, false)
  assert.equal(page.answers.value.one, 0)
  const retry = page.submitTest()
  assert.equal(submissions.length, 3)
  completeSubmission.resolve({ attempt_id: 'attempt', score: 100 })
  await retry
  assert.equal(page.isSubmitted.value, true)
  assert.equal(page.showResult.value, true)
  assert.equal(page.error.value, '')
  assert.equal(interval, null)
  await page.submitTest()
  assert.equal(submissions.length, 3)
  assert.deepEqual(submissions[1].answers, [
    { question_id: 'one', selected_index: 0, answer_text: null }
  ])
})
