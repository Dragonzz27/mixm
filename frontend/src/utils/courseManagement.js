import { categoryIds } from './contentNavigation.js'

// 管理页面读取全部状态的文章，不使用阅读接口默认的已发布筛选。
export async function readAllPages(fetchPage) {
  const items = []
  let total = 0
  do {
    const page = await fetchPage({ offset: items.length, limit: 1000 })
    if (!Array.isArray(page.items) || !Number.isInteger(page.total) || page.total < 0)
      throw new Error('课程数据分页响应无效')
    items.push(...page.items)
    total = page.total
    if (!page.items.length && items.length < total) throw new Error('课程数据分页响应不完整')
  } while (items.length < total)
  return items
}

export function articleCourses(courses, article) {
  const ids = new Set((article.categories || []).map((category) => category.id))
  return courses.filter((course) => categoryIds(course).some((id) => ids.has(id)))
}

// 保留仍属于所选课程的既有绑定，移出课程时同时处理它的全部下级分类。
export function courseBindingChanges(courses, categoryIDs, selectedCourseIDs) {
  const selected = new Set(selectedCourseIDs)
  const bound = new Set(categoryIDs)
  const add = []
  const remove = []
  for (const course of courses) {
    const existing = categoryIds(course).filter((id) => bound.has(id))
    if (selected.has(course.id)) {
      if (!existing.length) add.push(course.id)
    } else remove.push(...existing)
  }
  return { add, remove }
}

// 每次保存重新读取绑定，部分请求失败后重试也不会重复创建文章或丢失已有归属。
export async function syncArticleCourses(api, courses, articleID, selectedCourseIDs) {
  const bindings = await readAllPages((params) => api.getByContent(articleID, params))
  const changes = courseBindingChanges(
    courses,
    bindings.map((binding) => binding.category_id),
    selectedCourseIDs
  )
  for (const id of changes.add) await api.batchBind(id, { content_ids: [articleID] })
  for (const id of changes.remove) await api.batchUnbind(id, { content_ids: [articleID] })
}

export async function removeArticleFromCourse(api, course, article) {
  const ids = new Set((article.categories || []).map((category) => category.id))
  for (const id of categoryIds(course)) {
    if (ids.has(id)) await api.batchUnbind(id, { content_ids: [article.id] })
  }
}

// 先删除下级目录，避免数据库把它们提升为新的顶级课程；文章本身保持保留。
export async function deleteCourse(api, course) {
  for (const id of categoryIds(course).reverse()) await api.delete(id)
}

export function articleStatus(article) {
  if (article.is_active === false) return { label: '已停用', type: 'default' }
  if (article.status === 'published') return { label: '已发布', type: 'success' }
  if (article.status === 'archived') return { label: '已归档', type: 'default' }
  return { label: '草稿', type: 'warning' }
}

export function formatArticleDate(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date
    .toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hourCycle: 'h23'
    })
    .replaceAll('/', '-')
}
