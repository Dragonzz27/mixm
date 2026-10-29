import test from 'node:test'
import assert from 'node:assert/strict'
import {
  hasSimulationAnswer,
  simulationAnswers,
  simulationCorrectAnswer,
  simulationIsCorrect,
  simulationUserAnswer
} from '../src/utils/simulation.js'

const questions = [
  { id: 'single', question_type: 'single_choice', options: ['pwd', 'cd'], correct_index: 0 },
  {
    id: 'multi',
    question_type: 'multi_choice',
    options: ['列表', '字符串', '字典'],
    correct_indices: [0, 2]
  },
  { id: 'judge', question_type: 'true_false', correct_index: 0 },
  { id: 'blank', question_type: 'fill_blank', correct_answer: 'print' },
  { id: 'essay', question_type: 'short_answer', correct_answer: 'Persistent Volume' }
]

test('progress counts all five question types, including index zero, and ignores whitespace-only text', () => {
  const answers = { single: 0, multi: [0, 2], judge: 0, blank: 'print', essay: 'Persistent Volume' }
  assert.equal(
    questions.filter((question) => hasSimulationAnswer(question, answers[question.id])).length,
    5
  )
  for (const question of questions) assert.equal(hasSimulationAnswer(question, undefined), false)
  assert.equal(hasSimulationAnswer(questions[1], []), false)
  assert.equal(hasSimulationAnswer(questions[3], '  \n '), false)
  assert.equal(hasSimulationAnswer(questions[2], 1), true)
})

test('submission payload retains every multiple-choice selection and trims written answers', () => {
  const answers = simulationAnswers(questions, {
    single: 0,
    multi: [0, 2],
    judge: 1,
    blank: ' print ',
    essay: ' Persistent Volume '
  })
  assert.deepEqual(answers, [
    { question_id: 'single', selected_index: 0, answer_text: null },
    { question_id: 'multi', selected_index: 0, answer_text: '[0,2]' },
    { question_id: 'judge', selected_index: 1, answer_text: null },
    { question_id: 'blank', selected_index: null, answer_text: 'print' },
    { question_id: 'essay', selected_index: null, answer_text: 'Persistent Volume' }
  ])
  assert.deepEqual(
    simulationAnswers(questions, {}).map((answer) => answer.selected_index),
    [null, null, null, null, null]
  )
})

test('answer review formats and grades the same selections as the existing server contract', () => {
  const answers = simulationAnswers(questions, {
    single: 0,
    multi: [2, 0],
    judge: 0,
    blank: 'PRINT',
    essay: 'persistent\nvolume'
  })
  questions.forEach((question, index) =>
    assert.equal(simulationIsCorrect(question, answers[index]), true)
  )
  assert.equal(simulationUserAnswer(questions[2], answers[2]), '正确')
  assert.equal(simulationCorrectAnswer(questions[1]), 'A. 列表；C. 字典')
  assert.equal(simulationUserAnswer(questions[1], { answer_text: '[2,0]' }), 'C. 字典；A. 列表')
  assert.equal(simulationIsCorrect(questions[1], { answer_text: '[0]' }), false)
  assert.equal(simulationIsCorrect(questions[1], { answer_text: '[0,1,2]' }), false)
  assert.equal(simulationIsCorrect(questions[2], { selected_index: 1 }), false)
  questions.forEach((question) => assert.equal(simulationIsCorrect(question), false))
})

test('review handles unattempted questions and historical multiple-choice selections', () => {
  questions.forEach((question) => assert.equal(simulationUserAnswer(question), '未作答'))
  assert.equal(
    simulationUserAnswer(questions[1], { answer_text: 'invalid', selected_index: 2 }),
    'C. 字典'
  )
  assert.equal(simulationUserAnswer(questions[1], { answer_text: '{}' }), '未作答')
  assert.equal(simulationCorrectAnswer({ question_type: 'single_choice' }), '未提供参考答案')
})
