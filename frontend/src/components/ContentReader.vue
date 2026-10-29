<template>
  <div class="content-reader fullscreen-page">
    <div class="reader-layout" :class="{ 'directory-closed': !directoryOpen }">
      <aside
        v-show="directoryOpen"
        id="reading-directory"
        class="reader-directory"
        aria-label="阅读目录"
      >
        <div class="directory-heading">
          <span>阅读目录</span>
          <n-button
            class="directory-toggle"
            quaternary
            size="tiny"
            :aria-expanded="true"
            aria-controls="reading-directory"
            @click="directoryOpen = false"
          >
            <template #icon>
              <n-icon><ChevronBackOutline /></n-icon>
            </template>
            收起目录
          </n-button>
        </div>
        <div class="category-section">
          <n-input
            v-model:value="search"
            placeholder="搜索课程或文章"
            clearable
            size="small"
            aria-label="搜索课程或文章"
          />
          <n-spin :show="loadingDirectory">
            <n-tree
              v-if="treeData.length"
              :data="treeData"
              :pattern="search"
              :show-irrelevant-nodes="false"
              :selected-keys="selectedArticleKeys"
              v-model:expanded-keys="expandedKeys"
              block-line
              selectable
              :node-props="directoryNodeProps"
              :override-default-node-click-behavior="directoryClickBehavior"
              @update:selected-keys="selectNode"
            />
            <n-empty v-else-if="!loadingDirectory" size="small" description="暂无课程" />
          </n-spin>
        </div>
      </aside>
      <section ref="readingPane" class="reading-pane" aria-label="文章正文" :aria-busy="loading">
        <div v-if="!directoryOpen" class="reading-toolbar">
          <n-button
            class="directory-toggle"
            quaternary
            size="tiny"
            :aria-expanded="false"
            aria-controls="reading-directory"
            @click="directoryOpen = true"
          >
            <template #icon>
              <n-icon><ChevronForwardOutline /></n-icon>
            </template>
            展开目录
          </n-button>
        </div>
        <n-spin :show="loading" class="reading-spinner">
          <div v-if="loadError" class="reader-message">
            <n-result status="error" title="加载内容失败" :description="loadError">
              <template #footer><n-button @click="loadLocation">重试</n-button></template>
            </n-result>
          </div>
          <article v-else-if="article" class="reader-article">
            <header class="article-heading">
              <p v-if="breadcrumb" class="article-breadcrumb">{{ breadcrumb }}</p>
              <h1>{{ article.title }}</h1>
              <div class="article-meta">
                <span>{{ typeName }}</span>
                <span v-if="article.created_at">{{ formatDate(article.created_at) }}</span>
                <div v-if="article.url || article.template_id" class="article-actions">
                  <n-button v-if="article.url" size="small" secondary @click="openResource">
                    访问资源 ↗
                  </n-button>
                  <n-button
                    v-if="article.template_id"
                    size="small"
                    type="primary"
                    :loading="startingInstance"
                    @click="startInstance"
                  >
                    启动实例
                  </n-button>
                </div>
              </div>
            </header>
            <ContentBody :key="article.id" :value="article.content || ''" />
          </article>
          <div v-else-if="!loading" class="reader-message">
            <n-empty
              :description="
                courses.length ? '暂无可阅读的教学文章' : '暂无教学内容，请联系管理员添加'
              "
            />
          </div>
        </n-spin>
      </section>
    </div>
    <StartInstanceByTemplateModal
      v-model:show="showStartModal"
      :template="selectedTemplate"
      :readonly-config="true"
      @success="instanceStarted"
    />
  </div>
</template>

<script setup>
import { computed, h, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NEmpty, NIcon, NInput, NResult, NSpin, NTree, useMessage } from 'naive-ui'
import { BookOutline, ChevronBackOutline, ChevronForwardOutline } from '@vicons/ionicons5'
import { categoryApi, contentApi, templateApi } from '@/api'
import {
  activeCategories,
  categoryArticles,
  categoryIds,
  articleTreeKey,
  courseDirectory,
  findCategory
} from '@/utils/contentNavigation'
import ContentBody from '@/components/ContentBody.vue'
import StartInstanceByTemplateModal from '@/components/StartInstanceByTemplateModal.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const categories = ref([])
const selectedCourse = ref(null)
const courses = ref([])
const article = ref(null)
const expandedKeys = ref([])
const search = ref('')
const directoryOpen = ref(true)
const loading = ref(true)
const loadingDirectory = ref(true)
const loadError = ref('')
const readingPane = ref(null)
const startingInstance = ref(false)
const showStartModal = ref(false)
const selectedTemplate = ref(null)
let directoryLoaded = false
let directoryRequest = null
let navigationVersion = 0
let scrollOnLoad = false

const treeData = computed(() =>
  courses.value.map((course) => ({
    key: `course:${course.id}`,
    label: course.name,
    kind: 'course',
    isLeaf: course.articles.length === 0,
    prefix: () => h(NIcon, null, { default: () => h(BookOutline) }),
    children: course.articles.map((lesson) => ({
      key: articleTreeKey(course.id, lesson.id),
      label: lesson.title,
      courseId: course.id,
      articleId: lesson.id,
      isLeaf: true
    }))
  }))
)
const selectedArticleKeys = computed(() =>
  article.value && selectedCourse.value
    ? [articleTreeKey(selectedCourse.value.id, article.value.id)]
    : []
)
const breadcrumb = computed(() => selectedCourse.value?.name || '')

const typeName = computed(
  () =>
    ({
      tutorial: '教程',
      documentation: '文档',
      experiment: '实验',
      video: '视频',
      exercise: '练习',
      article: '文章',
      resource: '资源'
    })[article.value?.content_type] || '文章'
)

async function loadDirectory() {
  if (directoryLoaded) return
  if (!directoryRequest) {
    loadingDirectory.value = true
    directoryRequest = (async () => {
      const data = await categoryApi.getTree({})
      categories.value = activeCategories(data || [])
      const articles = await categoryArticles(contentApi, categories.value)
      courses.value = courseDirectory(categories.value, articles)
      expandedKeys.value = courses.value.map((course) => `course:${course.id}`)
      directoryLoaded = true
    })().finally(() => {
      loadingDirectory.value = false
      directoryRequest = null
    })
  }
  await directoryRequest
}

// 仅文章节点切换正文；旧分类链接用于定位所属课程及其文章，不新增目录层级。
async function loadLocation() {
  const version = ++navigationVersion
  const detailId = route.params.id ? String(route.params.id) : ''
  const categoryLink = route.query.type === 'category'
  const categoryId = categoryLink ? detailId : String(route.query.category || '')
  const articleId = detailId && !categoryLink ? detailId : String(route.query.article || '')
  loading.value = true
  loadError.value = ''
  article.value = null
  try {
    await loadDirectory()
    if (version !== navigationVersion) return
    const category = categoryId ? findCategory(categories.value, categoryId) : null
    if (categoryId && !category) throw new Error('课程不存在或已停用')
    let course = category
      ? courses.value.find((item) => categoryIds(item).includes(category.id))
      : null
    let detail = null
    if (articleId) {
      detail = await contentApi.getDetail(articleId)
      if (version !== navigationVersion) return
      if (course && !course.articles.some((lesson) => lesson.id === detail.id)) {
        throw new Error('当前课程中没有这篇已发布文章')
      }
      if (!course)
        course = courses.value.find((item) =>
          item.articles.some((lesson) => lesson.id === detail.id)
        )
    } else {
      course ||= courses.value.find((item) => item.articles.length) || courses.value[0]
      let lessons = course?.articles || []
      if (category) {
        const ids = new Set(categoryIds(category))
        lessons = lessons.filter((lesson) =>
          (lesson.categories || []).some((item) => ids.has(item.id))
        )
      }
      if (lessons.length) {
        detail = await contentApi.getDetail(lessons[0].id)
        if (version !== navigationVersion) return
      }
    }
    selectedCourse.value = course || null
    article.value = detail
    await nextTick()
    if (version !== navigationVersion) return
    if (readingPane.value) readingPane.value.scrollTop = 0
    if (scrollOnLoad && window.matchMedia('(max-width: 800px)').matches) {
      readingPane.value?.scrollIntoView({ block: 'start' })
    }
    scrollOnLoad = false
  } catch (error) {
    if (version === navigationVersion)
      loadError.value = error.response?.data?.message || error.message || '请稍后重试'
  } finally {
    if (version === navigationVersion) loading.value = false
  }
}

function directoryClickBehavior({ option }) {
  return option.articleId ? 'toggleSelect' : 'none'
}

function directoryNodeProps({ option }) {
  return { class: option.articleId ? 'lesson-node' : 'course-node' }
}

function selectNode(keys, options) {
  const lesson = options[0]
  if (!keys.length || !lesson?.articleId) return
  scrollOnLoad = true
  router.push({
    path: '/app/contents',
    query: { category: lesson.courseId, article: lesson.articleId }
  })
}

function formatDate(value) {
  return new Date(value).toLocaleDateString('zh-CN')
}
function openResource() {
  window.open(article.value.url, '_blank', 'noopener,noreferrer')
}

async function startInstance() {
  const current = article.value
  if (!current?.template_id) return
  startingInstance.value = true
  try {
    const template = await templateApi.getDetail(current.template_id)
    if (article.value?.id !== current.id) return
    selectedTemplate.value = template
    showStartModal.value = true
  } catch (error) {
    message.error('加载实例模板失败')
  } finally {
    startingInstance.value = false
  }
}

function instanceStarted() {
  message.success('实例已加入创建队列')
  router.push('/app/my-instances')
}

watch(() => route.fullPath, loadLocation, { immediate: true })
onBeforeUnmount(() => {
  navigationVersion++
})
</script>

<style scoped>
.content-reader {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 620px;
  padding: 24px;
}
.reader-layout {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  border: 1px solid var(--mixm-border);
  border-radius: 10px;
  background: var(--mixm-surface);
  overflow: hidden;
}
.reader-layout.directory-closed {
  grid-template-columns: minmax(0, 1fr);
}
.reader-directory {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow-y: auto;
  border-right: 1px solid var(--mixm-border);
  background: var(--mixm-paper);
}
.directory-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 12px 14px 18px;
  color: var(--mixm-ink);
  font-weight: 600;
}
.directory-toggle {
  height: 28px;
  padding: 0 8px;
  border-radius: 6px;
  color: var(--mixm-muted);
  font-size: 12px;
  font-weight: 400;
}
.directory-toggle :deep(.n-button__icon) {
  font-size: 14px;
}
.category-section {
  padding: 0 12px 18px;
}
.category-section :deep(.n-input) {
  margin-bottom: 14px;
}
.category-section :deep(.n-tree-node-content) {
  min-width: 0;
}
.category-section :deep(.n-tree-node-content__text) {
  white-space: normal;
  overflow-wrap: anywhere;
}
.category-section :deep(.lesson-node .n-tree-node-indent),
.category-section :deep(.lesson-node .n-tree-node-switcher) {
  display: none;
}
.category-section :deep(.lesson-node .n-tree-node-content) {
  padding-left: 0;
}
.category-section :deep(.course-node),
.category-section :deep(.course-node .n-tree-node-content) {
  cursor: default;
  font-weight: 600;
}
.reading-pane {
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}
.reading-toolbar {
  padding: 12px 18px 0;
}
.reading-spinner {
  min-height: 100%;
}
.reader-article {
  max-width: 980px;
  margin: 0 auto;
  padding: 36px 48px 64px;
}
.article-heading {
  padding-bottom: 24px;
  margin-bottom: 28px;
  border-bottom: 1px solid var(--mixm-border);
}
.article-breadcrumb {
  margin: 0 0 16px;
  color: var(--mixm-muted);
  font-size: 12px;
  overflow-wrap: anywhere;
}
.article-heading h1 {
  margin: 0 0 18px;
  font-family: var(--mixm-font-display);
  font-size: 30px;
  font-weight: 500;
  line-height: 1.5;
  overflow-wrap: anywhere;
}
.article-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  color: var(--mixm-muted);
  font-size: 12px;
}
.article-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-left: auto;
}
.reader-message {
  min-height: 460px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
}
@media (max-width: 1100px) {
  .reader-layout {
    grid-template-columns: 280px minmax(0, 1fr);
  }
  .reader-article {
    padding: 28px;
  }
  .article-heading h1 {
    font-size: 25px;
  }
}
@media (max-width: 800px) {
  .content-reader {
    height: auto;
    min-height: 100%;
    padding: 20px 16px;
  }
  .reader-layout {
    display: flex;
    flex-direction: column;
    overflow: visible;
  }
  .reader-directory {
    max-height: 360px;
    border-right: 0;
    border-bottom: 1px solid var(--mixm-border);
    border-radius: 10px 10px 0 0;
  }
  .reading-pane {
    overflow: visible;
  }
  .reader-article {
    padding: 24px 18px 40px;
  }
  .article-heading h1 {
    font-size: 23px;
  }
  .article-actions {
    margin-left: 0;
  }
}
</style>
