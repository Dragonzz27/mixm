<template>
  <div class="contents">
    <n-card title="内容管理">
      <template #header-extra>
        <n-button type="primary" :disabled="!ready" @click="openEditor()">新建文章</n-button>
      </template>
      <p class="management-description">按课程管理教学文章，课程与文章的归属会同步用于内容浏览。</p>
      <n-alert v-if="error" type="error" :bordered="false" class="library-error">
        {{ error }}
        <n-button size="small" class="load-retry" @click="load">重试</n-button>
      </n-alert>
      <div class="content-management-layout">
        <aside class="course-filter" aria-label="按课程筛选文章">
          <div class="filter-heading">课程目录</div>
          <button
            class="course-filter-item"
            :class="{ active: selectedCourse === 'all' }"
            @click="selectCourse('all')"
          >
            <span>全部文章</span>
            <span class="filter-count">{{ articles.length }}</span>
          </button>
          <button
            v-for="course in courses"
            :key="course.id"
            class="course-filter-item"
            :class="{ active: selectedCourse === course.id }"
            @click="selectCourse(course.id)"
          >
            <span>
              {{ course.name }}
              <small v-if="course.is_active === false">已停用</small>
            </span>
            <span class="filter-count">{{ course.articles.length }}</span>
          </button>
          <button
            class="course-filter-item unassigned-filter"
            :class="{ active: selectedCourse === 'unassigned' }"
            @click="selectCourse('unassigned')"
          >
            <span>未分配课程</span>
            <span class="filter-count">{{ unassigned.length }}</span>
          </button>
        </aside>
        <section class="article-management" aria-label="教学文章列表">
          <div class="article-list-toolbar">
            <div>
              <strong>{{ selectedCourseName }}</strong>
              <span>{{ filteredArticles.length }} 篇文章</span>
            </div>
            <n-input
              v-model:value="searchKey"
              placeholder="搜索文章标题"
              clearable
              class="article-search"
              aria-label="搜索文章标题"
            />
          </div>
          <n-data-table
            :columns="columns"
            :data="filteredArticles"
            :pagination="pagination"
            :loading="loading"
            :scroll-x="1010"
            :bordered="false"
            :row-key="(row) => row.id"
          />
        </section>
      </div>
    </n-card>
    <ArticleEditorModal
      v-model:show="showEditor"
      :article-id="editingId"
      :default-course-id="defaultCourseId"
      :courses="categories"
      @saved="load"
    />
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NAlert, NButton, NCard, NDataTable, NInput, NSpace, NTag, useMessage } from 'naive-ui'
import { contentApi } from '@/api'
import ArticleEditorModal from '@/components/ArticleEditorModal.vue'
import { useCourseLibrary } from '@/composables/useCourseLibrary'
import { useContentTypeStore } from '@/stores/contentType'
import { articleCourses, articleStatus, formatArticleDate } from '@/utils/courseManagement'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const contentTypes = useContentTypeStore()
const { categories, articles, courses, unassigned, loading, error, ready, load } =
  useCourseLibrary()
const searchKey = ref('')
const selectedCourse = computed(() => String(route.query.course || 'all'))
const selectedCourseName = computed(() =>
  selectedCourse.value === 'all'
    ? '全部教学文章'
    : selectedCourse.value === 'unassigned'
      ? '未分配课程'
      : courses.value.find((course) => course.id === selectedCourse.value)?.name || '课程不存在'
)
const filteredArticles = computed(() => {
  let items = articles.value
  if (selectedCourse.value === 'unassigned') items = unassigned.value
  else if (selectedCourse.value !== 'all')
    items = courses.value.find((course) => course.id === selectedCourse.value)?.articles || []
  const keyword = searchKey.value.trim().toLocaleLowerCase()
  return items.filter((article) => article.title.toLocaleLowerCase().includes(keyword))
})
const showEditor = ref(false)
const editingId = ref(null)
const defaultCourseId = ref(null)
const pagination = ref({
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange(page) {
    pagination.value.page = page
  },
  onUpdatePageSize(size) {
    pagination.value.pageSize = size
    pagination.value.page = 1
  }
})
const columns = [
  {
    title: '序号',
    key: 'index',
    width: 60,
    render: (_, index) => (pagination.value.page - 1) * pagination.value.pageSize + index + 1
  },
  { title: '文章标题', key: 'title', minWidth: 220, ellipsis: { tooltip: true } },
  {
    title: '所属课程',
    key: 'courses',
    minWidth: 180,
    render: (article) => {
      const memberships = articleCourses(categories.value, article)
      return memberships.length
        ? h(NSpace, { size: 4 }, () =>
            memberships.map((course) =>
              h(NTag, { size: 'small', bordered: false }, () => course.name)
            )
          )
        : h('span', { class: 'unassigned-label' }, '未分配课程')
    }
  },
  {
    title: '类型',
    key: 'content_type',
    width: 100,
    render: (article) => contentTypes.getTypeName(article.content_type)
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (article) => {
      const status = articleStatus(article)
      return h(NTag, { size: 'small', type: status.type, bordered: false }, () => status.label)
    }
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 200,
    minWidth: 200,
    className: 'created-at-cell',
    render: (article) => formatArticleDate(article.created_at)
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (article) =>
      h(NSpace, { size: 8 }, () => [
        h(NButton, { size: 'small', onClick: () => openEditor(article.id) }, () => '编辑'),
        h(
          NButton,
          { size: 'small', type: 'error', onClick: () => deleteArticle(article) },
          () => '删除'
        )
      ])
  }
]

function selectCourse(id) {
  router.replace({ query: id === 'all' ? {} : { course: id } })
}
function openEditor(id = null) {
  editingId.value = id
  defaultCourseId.value = categories.value.some((course) => course.id === selectedCourse.value)
    ? selectedCourse.value
    : null
  showEditor.value = true
}
async function deleteArticle(article) {
  try {
    await contentApi.delete(article.id)
    message.success('文章已删除')
    await load()
  } catch (cause) {
    message.error(cause.response?.data?.message || '删除文章失败')
  }
}
watch([selectedCourse, searchKey], () => {
  pagination.value.page = 1
})
watch(
  () => filteredArticles.value.length,
  (length) => {
    pagination.value.page = Math.min(
      pagination.value.page,
      Math.max(1, Math.ceil(length / pagination.value.pageSize))
    )
  }
)
onMounted(load)
</script>

<style scoped>
.load-retry {
  margin-left: 12px;
}
.management-description {
  margin: 0 0 20px;
  color: var(--mixm-muted);
  font-size: 13px;
}
.library-error {
  margin-bottom: 20px;
}
.content-management-layout {
  display: grid;
  grid-template-columns: 200px minmax(0, 1fr);
  gap: 20px;
}
.course-filter {
  align-self: start;
  padding: 12px;
  border: 1px solid var(--mixm-border);
  border-radius: 8px;
  background: var(--mixm-paper);
}
.filter-heading {
  padding: 6px 10px 14px;
  font-size: 13px;
  font-weight: 600;
}
.course-filter-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  padding: 10px;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: var(--mixm-muted);
  font: inherit;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
}
.course-filter-item > span:first-child {
  min-width: 0;
  overflow-wrap: anywhere;
}
.course-filter-item small {
  display: block;
  font-size: 11px;
  margin-top: 2px;
}
.course-filter-item:hover {
  background: var(--mixm-surface-muted);
}
.course-filter-item.active {
  background: var(--mixm-accent-soft);
  color: var(--mixm-accent);
}
.course-filter-item:focus-visible {
  outline: 2px solid var(--mixm-accent);
}
.filter-count {
  flex-shrink: 0;
  font-size: 12px;
}
.unassigned-filter {
  margin-top: 10px;
  border-top: 1px solid var(--mixm-border);
  border-radius: 0;
}
.article-management {
  min-width: 0;
}
.article-list-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}
.article-list-toolbar strong {
  font-size: 14px;
  font-weight: 600;
}
.article-list-toolbar span {
  margin-left: 10px;
  font-size: 12px;
  color: var(--mixm-muted);
}
.article-search {
  width: 220px;
  max-width: 100%;
}
.contents :deep(.created-at-cell) {
  white-space: nowrap;
}
.contents :deep(.unassigned-label) {
  color: var(--mixm-muted);
  font-size: 12px;
}
@media (max-width: 900px) {
  .content-management-layout {
    grid-template-columns: minmax(0, 1fr);
  }
  .course-filter {
    max-height: 240px;
    overflow-y: auto;
  }
}
</style>
