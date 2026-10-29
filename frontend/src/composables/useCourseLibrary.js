import { computed, ref } from 'vue'
import { categoryApi, contentApi } from '@/api'
import { courseDirectory } from '@/utils/contentNavigation'
import { articleCourses, readAllPages } from '@/utils/courseManagement'

// 两个管理页面使用与阅读目录相同的课程归属规则，并包含草稿和停用记录。
export function useCourseLibrary() {
  const categories = ref([])
  const articles = ref([])
  const loading = ref(false)
  const error = ref('')
  const ready = ref(false)
  const courses = computed(() => courseDirectory(categories.value, articles.value))
  const unassigned = computed(() =>
    articles.value.filter((article) => !articleCourses(categories.value, article).length)
  )
  let version = 0

  async function load() {
    const request = ++version
    loading.value = true
    error.value = ''
    try {
      const [tree, items] = await Promise.all([
        categoryApi.getTree(),
        readAllPages((params) => contentApi.getList(params))
      ])
      if (request !== version) return
      categories.value = tree
      articles.value = items
      ready.value = true
    } catch (cause) {
      if (request === version)
        error.value = cause.response?.data?.message || cause.message || '课程数据加载失败'
    } finally {
      if (request === version) loading.value = false
    }
  }

  return { categories, articles, courses, unassigned, loading, error, ready, load }
}
