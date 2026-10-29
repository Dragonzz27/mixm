// 读取启用的课程与下级分类，供两层阅读目录汇总文章。
export function activeCategories(categories) {
  return categories
    .filter((category) => category.is_active !== false)
    .map((category) => ({
      ...category,
      children: activeCategories(category.children || [])
    }))
}

export function findCategory(categories, id) {
  for (const category of categories) {
    if (category.id === id) return category
    const child = findCategory(category.children || [], id)
    if (child) return child
  }
  return null
}

export function categoryIds(category) {
  return [category.id, ...(category.children || []).flatMap(categoryIds)]
}

// 顶级分类作为课程，中间分类只用于归属判断，不作为额外目录层级。
export function courseDirectory(categories, articles) {
  return categories.map((course) => {
    const ids = new Set(categoryIds(course))
    const seen = new Set()
    const lessons = articles.filter((article) => {
      if (
        seen.has(article.id) ||
        !(article.categories || []).some((category) => ids.has(category.id))
      )
        return false
      seen.add(article.id)
      return true
    })
    return { ...course, articles: lessons }
  })
}

export function articleTreeKey(courseId, articleId) {
  return `article:${courseId}:${articleId}`
}

// 批量读取所有启用课程及子分类下的文章，避免按课程重复请求与分页截断。
export async function categoryArticles(api, categories) {
  const ids = categories.flatMap(categoryIds)
  if (!ids.length) return []
  const items = []
  let total = 0
  do {
    const page = await api.getByCategories(ids, {
      offset: items.length,
      limit: 1000,
      status: 'published',
      is_active: true
    })
    items.push(...page.items)
    total = page.total
    if (page.items.length === 0 && items.length < total) throw new Error('文章目录分页响应不完整')
  } while (items.length < total)
  return items
}
