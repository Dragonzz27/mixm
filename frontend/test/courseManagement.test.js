import test from 'node:test'
import assert from 'node:assert/strict'
import {
  articleCourses,
  articleStatus,
  courseBindingChanges,
  deleteCourse,
  readAllPages,
  removeArticleFromCourse,
  syncArticleCourses
} from '../src/utils/courseManagement.js'
import { courseDirectory } from '../src/utils/contentNavigation.js'

const courses = [
  { id: 'a', name: '课程 A', children: [{ id: 'chapter', children: [{ id: 'deep' }] }] },
  { id: 'b', name: '课程 B', is_active: false },
  { id: 'c', name: '课程 C' }
]

test('administration and reading use the same course membership for deep and shared articles', () => {
  const article = { id: 'article', categories: [{ id: 'deep' }, { id: 'b' }] }
  assert.deepEqual(
    articleCourses(courses, article).map((course) => course.id),
    ['a', 'b']
  )
  assert.deepEqual(
    courseDirectory(courses, [article]).map((course) => course.articles.length),
    [1, 1, 0]
  )
  assert.deepEqual(articleCourses(courses, { categories: [] }), [])
})

test('course changes preserve unchanged descendant bindings and remove the entire deselected course', () => {
  assert.deepEqual(courseBindingChanges(courses, ['chapter', 'deep', 'b'], ['a', 'c']), {
    add: ['c'],
    remove: ['b']
  })
  assert.deepEqual(courseBindingChanges(courses, ['a', 'chapter', 'deep', 'b'], ['b']), {
    add: [],
    remove: ['a', 'chapter', 'deep']
  })
  assert.deepEqual(courseBindingChanges(courses, ['deep'], []), { add: [], remove: ['deep'] })
})

test('all pages are read for course counts and binding edits without publication filters', async () => {
  const records = Array.from({ length: 1003 }, (_, id) => ({ id }))
  const calls = []
  const result = await readAllPages(async (params) => {
    calls.push(params)
    return {
      items: records.slice(params.offset, params.offset + params.limit),
      total: records.length
    }
  })
  assert.equal(result.length, 1003)
  assert.deepEqual(calls, [
    { offset: 0, limit: 1000 },
    { offset: 1000, limit: 1000 }
  ])
  await assert.rejects(
    readAllPages(async () => ({ items: [], total: 3 })),
    /不完整/
  )
  await assert.rejects(
    readAllPages(async () => ({ items: [], total: '3' })),
    /无效/
  )
})

test('retrying a partially completed course reassignment keeps the article and avoids duplicate additions', async () => {
  const links = new Set(['deep', 'b'])
  const calls = []
  let fail = true
  const api = {
    async getByContent(id) {
      assert.equal(id, 'article')
      return { items: [...links].map((category_id) => ({ category_id })), total: links.size }
    },
    async batchBind(id, data) {
      assert.deepEqual(data, { content_ids: ['article'] })
      calls.push(`add:${id}`)
      links.add(id)
    },
    async batchUnbind(id) {
      calls.push(`remove:${id}`)
      if (fail) throw new Error('暂时失败')
      links.delete(id)
    }
  }
  await assert.rejects(syncArticleCourses(api, courses, 'article', ['c']), /暂时失败/)
  assert.deepEqual([...links], ['deep', 'b', 'c'])
  fail = false
  await syncArticleCourses(api, courses, 'article', ['c'])
  assert.deepEqual([...links], ['c'])
  assert.equal(calls.filter((call) => call === 'add:c').length, 1)
})

test('removing a shared article clears only the chosen course including its old descendants', async () => {
  const calls = []
  await removeArticleFromCourse(
    {
      async batchUnbind(id, data) {
        calls.push({ id, data })
      }
    },
    courses[0],
    {
      id: 'article',
      categories: [{ id: 'a' }, { id: 'deep' }, { id: 'b' }]
    }
  )
  assert.deepEqual(calls, [
    { id: 'a', data: { content_ids: ['article'] } },
    { id: 'deep', data: { content_ids: ['article'] } }
  ])
})

test('deleting a course removes descendants before the root and never calls article deletion', async () => {
  const calls = []
  await deleteCourse(
    {
      async delete(id) {
        calls.push(id)
      }
    },
    courses[0]
  )
  assert.deepEqual(calls, ['deep', 'chapter', 'a'])
  const failedCalls = []
  await assert.rejects(
    deleteCourse(
      {
        async delete(id) {
          failedCalls.push(id)
          throw new Error('失败')
        }
      },
      courses[0]
    )
  )
  assert.deepEqual(failedCalls, ['deep'])
})

test('draft and disabled articles remain distinguishable in administration', () => {
  assert.equal(articleStatus({ status: 'draft' }).label, '草稿')
  assert.equal(articleStatus({ status: 'published' }).label, '已发布')
  assert.equal(articleStatus({ status: 'archived' }).label, '已归档')
  assert.equal(articleStatus({ status: 'published', is_active: false }).label, '已停用')
})
