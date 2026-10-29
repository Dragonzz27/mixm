import test from 'node:test'
import assert from 'node:assert/strict'
import {
  activeCategories,
  categoryArticles,
  categoryIds,
  articleTreeKey,
  courseDirectory,
  findCategory
} from '../src/utils/contentNavigation.js'

test('reading navigation finds deep categories and excludes disabled branches', () => {
  const input = [
    {
      id: 'root',
      name: '课程',
      children: [
        {
          id: 'disabled',
          name: '已停用',
          is_active: false,
          children: [{ id: 'hidden', name: '不可见' }]
        },
        { id: 'chapter', name: '章节', children: [{ id: 'lesson', name: '文章分类' }] }
      ]
    }
  ]
  const tree = activeCategories(input)
  assert.deepEqual(categoryIds(tree[0]), ['root', 'chapter', 'lesson'])
  assert.equal(findCategory(tree, 'lesson').name, '文章分类')
  assert.equal(findCategory(tree, 'hidden'), null)
  assert.equal(input[0].children.length, 2)
})

test('a category opens published articles from its descendants across every page', async () => {
  const items = Array.from({ length: 1003 }, (_, index) => ({ id: `article-${index}` }))
  const calls = []
  const api = {
    async getByCategories(ids, params) {
      calls.push({ ids, params })
      return {
        items: items.slice(params.offset, params.offset + params.limit),
        total: items.length
      }
    }
  }
  const result = await categoryArticles(api, [{ id: 'root', children: [{ id: 'child' }] }])
  assert.equal(result.length, 1003)
  assert.equal(result.at(-1).id, 'article-1002')
  assert.deepEqual(
    calls.map((call) => call.params.offset),
    [0, 1000]
  )
  assert.deepEqual(calls[0].ids, ['root', 'child'])
  assert.equal(calls[0].params.status, 'published')
  assert.equal(calls[0].params.is_active, true)
})

test('empty categories render as empty while incomplete pagination is an error', async () => {
  const category = { id: 'empty' }
  assert.deepEqual(
    await categoryArticles({ getByCategories: async () => ({ items: [], total: 0 }) }, [category]),
    []
  )
  await assert.rejects(
    categoryArticles({ getByCategories: async () => ({ items: [], total: 1 }) }, [category]),
    /文章目录分页响应不完整/
  )
})

test('course directory gathers descendants directly into each course and deduplicates lessons', () => {
  const courses = activeCategories([
    { id: 'course-a', name: '课程 A', children: [{ id: 'chapter', children: [{ id: 'deep' }] }] },
    { id: 'course-b', name: '课程 B' },
    { id: 'empty', name: '空课程' }
  ])
  const first = { id: 'one', title: '文章一', categories: [{ id: 'deep' }] }
  const shared = {
    id: 'shared',
    title: '共享文章',
    categories: [{ id: 'chapter' }, { id: 'course-b' }]
  }
  const directory = courseDirectory(courses, [
    first,
    shared,
    first,
    { id: 'unbound', categories: [] }
  ])
  assert.deepEqual(
    directory[0].articles.map((item) => item.id),
    ['one', 'shared']
  )
  assert.deepEqual(
    directory[1].articles.map((item) => item.id),
    ['shared']
  )
  assert.deepEqual(directory[2].articles, [])
  assert.notEqual(articleTreeKey('course-a', 'shared'), articleTreeKey('course-b', 'shared'))
})

test('an empty course tree does not query the entire content library', async () => {
  const result = await categoryArticles(
    {
      getByCategories() {
        throw new Error('unexpected request')
      }
    },
    []
  )
  assert.deepEqual(result, [])
})
