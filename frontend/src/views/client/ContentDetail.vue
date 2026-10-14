<template>
  <div class="content-detail">
    <div class="detail-header">
      <router-link to="/app/contents" class="back-link" v-if="viewType === 'category'">
        <n-icon>
          <ArrowBackOutline />
        </n-icon>
        返回列表
      </router-link>
      <div v-else class="detail-header-actions">
        <div class="back-link" @click="viewType = 'category'">
        <n-icon>
          <ArrowBackOutline />
        </n-icon>
        返回课程
      </div>
      <div class="article-actions">
        <n-button type="primary" @click="openUrl(content?.url)" v-if="content?.url">
          <template #icon>
            <n-icon>
              <LinkOutline />
            </n-icon>
          </template>
          访问资源链接
        </n-button>
        <n-button type="success" @click="handleStartInstance" v-if="content?.template_id" :loading="startingInstance">
          <template #icon>
            <n-icon>
              <PlayOutline />
            </n-icon>
          </template>
          启动实例
        </n-button>
      </div>
      </div>
    </div>

    <n-spin :show="loading">
      <div class="detail-layout">
        <main class="detail-main">
          <div v-if="viewType === 'category'" class="category-view">
            <div class="category-header">
              <!-- <n-tag type="info" size="small">分类</n-tag> -->
              <h1 class="title">{{ currentCategory?.name }}</h1>
              <div class="meta">
                <span class="time">创建于 {{ formatTime(currentCategory?.created_at) }}</span>
              </div>
            </div>

            <div v-if="currentCategory?.description" class="category-description">
              <p>{{ currentCategory.description }}</p>
            </div>

            <div class="divider" v-if="currentCategoryContents.length > 0">
              <span class="divider-text">绑定内容</span>
            </div>

            <div v-if="currentCategoryContents.length > 0" class="contents-grid">
              <div v-for="item in currentCategoryContents" :key="item.id" class="content-item"
                @click="viewContent(item)">
                <div class="content-thumb" :style="{ background: getThumbColor(item.content_type) }">
                  <n-tag :type="getTypeTagType(item.content_type)" size="small">
                    {{ getTypeName(item.content_type) }}
                  </n-tag>
                </div>
                <div class="content-info">
                  <h4>{{ item.title }}</h4>
                  <p>{{ getPreviewText(item.content) }}</p>
                  <div class="content-meta">
                    <span class="time">{{ formatTime(item.created_at) }}</span>
                  </div>
                </div>
              </div>
            </div>
            <n-empty v-else-if="!hasSubCategories" description="该分类下暂无内容" />
          </div>

          <div v-else-if="viewType === 'content'" class="content-view">
            <article class="article">
              <header class="article-header">
                <n-tag :type="getTypeTagType(content.content_type)" size="small">
                  {{ getTypeName(content.content_type) }}
                </n-tag>
                <h1 class="title">{{ content.title }}</h1>
                <div class="meta">
                  <span class="time">创建于 {{ formatTime(content.created_at) }}</span>
                </div>
              </header>
              <div class="article-body" v-if="content.content">
                <div v-if="contentRenderMode === 'md'" ref="markdownPreviewRef" class="markdown-preview"></div>
                <QuillEditor v-else-if="contentRenderMode === 'web'" ref="quillPreviewRef" contentType="html"
                  :options="quillPreviewOptions" @ready="onQuillPreviewReady" />
                <div v-else v-html="renderFallbackContent"></div>
              </div>


            </article>
          </div>
        </main>

        <aside class="detail-sidebar" v-if="viewType === 'category'">
          <div class="sidebar-header">
            <span>目录</span>
          </div>
          <div class="sidebar-wrapper">
            <n-input v-model:value="searchPattern" placeholder="搜索..." clearable size="small"
              style="margin-bottom: 12px" />
            <n-tree :data="treeData" :selected-keys="selectedKeys" :expanded-keys="expandedKeys"
              :pattern="searchPattern" :show-search="false" block-line selectable default-expand-all
              @update:selected-keys="handleSelect" @update:expanded-keys="handleExpand" />
          </div>
        </aside>
      </div>

      <StartInstanceByTemplateModal v-model:show="showStartModal" :template="selectedTemplate"
        :enable-user-select="false" :readonly-config="true" @success="onInstanceStarted" />

      <n-empty v-if="!loading && !content && viewType !== 'category'" description="内容不存在" style="padding: 60px 0;">
        <template #extra>
          <n-button @click="$router.push('/app/contents')">返回列表</n-button>
        </template>
      </n-empty>
    </n-spin>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, nextTick, watch, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import VditorPreview from 'vditor/dist/method.min'
import 'vditor/dist/index.css'
import { contentApi, templateApi } from '@/api'
import StartInstanceByTemplateModal from '@/components/StartInstanceByTemplateModal.vue'
import { ArrowBackOutline, LinkOutline, PlayOutline } from '@vicons/ionicons5'

const router = useRouter()
const route = useRoute()
const message = useMessage()

const loading = ref(true)
const startingInstance = ref(false)
const showStartModal = ref(false)
const selectedTemplate = ref(null)
const content = ref(null)
const categoryTree = ref(null)
const currentCategory = ref(null)
const currentCategoryContents = ref([])
const selectedKeys = ref([])
const expandedKeys = ref([])
const searchPattern = ref('')
const viewType = ref('category')
const markdownPreviewRef = ref(null)
const quillPreviewRef = ref(null)

const quillPreviewOptions = {
  theme: 'snow',
  readOnly: true,
  modules: {
    toolbar: false
  }
}

const hasSubCategories = computed(() => {
  if (!currentCategory.value) return false
  const node = findNodeById(categoryTree.value, currentCategory.value.id)
  return node?.children?.length > 0
})

const treeData = computed(() => {
  if (!categoryTree.value) return []

  function buildTree(node) {
    const result = {
      key: node.id,
      label: node.name
    }

    if (node.children && node.children.length > 0) {
      result.children = node.children.map(buildTree)
    }

    return result
  }

  return [buildTree(categoryTree.value)]
})

const typeMap = {
  article: { name: '文章', color: '#2080f0', tag: 'info' },
  video: { name: '视频', color: '#f0a020', tag: 'warning' },
  exercise: { name: '练习', color: '#18a058', tag: 'success' },
  resource: { name: '资源', color: '#d03050', tag: 'error' }
}

const contentRenderMode = computed(() => {
  const raw = content.value?.content || ''
  if (!raw) return ''
  if (raw.startsWith('[md]')) return 'md'
  if (raw.startsWith('[web]')) return 'web'
  const noPrefix = raw.trim()
  return /<\/?[a-z][\s\S]*>/i.test(noPrefix) ? 'web' : 'md'
})

const normalizedContent = computed(() => {
  const raw = content.value?.content || ''
  if (!raw) return ''
  if (raw.startsWith('[md]')) return raw.slice(4)
  if (raw.startsWith('[web]')) return raw.slice(5)
  return raw
})

const renderFallbackContent = computed(() => {
  return normalizedContent.value
})

async function renderMarkdownPreview() {
  if (!markdownPreviewRef.value || contentRenderMode.value !== 'md') return
  await VditorPreview.preview(markdownPreviewRef.value, normalizedContent.value || '', {
    mode: 'light'
  })
  VditorPreview.mermaidRender(document)
}

function renderQuillPreview() {
  if (contentRenderMode.value !== 'web') return
  const quill = quillPreviewRef.value?.getQuill()
  if (!quill) return
  quill.root.innerHTML = normalizedContent.value || ''
  quill.enable(false)
}

function onQuillPreviewReady() {
  renderQuillPreview()
}

function getPreviewText(text) {
  if (!text) return ''
  if (text.startsWith('[md]')) {
    const mdText = text.substring(3)
    return mdText.length > 60 ? mdText.substring(0, 60) + '...' : mdText
  }
  const stripped = text.replace(/<[^>]+>/g, '')
  return stripped.length > 60 ? stripped.substring(0, 60) + '...' : stripped
}

watch(
  () => [contentRenderMode.value, normalizedContent.value, viewType.value],
  async () => {
    if (viewType.value !== 'content') return
    await nextTick()
    if (contentRenderMode.value === 'md') {
      await renderMarkdownPreview()
      return
    }
    if (contentRenderMode.value === 'web') {
      renderQuillPreview()
    }
  },
  { immediate: true }
)

function getTypeName(type) {
  return typeMap[type]?.name || '文章'
}

function getTypeTagType(type) {
  return typeMap[type]?.tag || 'info'
}

function getThumbColor(type) {
  return typeMap[type]?.color || '#2080f0'
}

function getStatusName(status) {
  const map = { draft: '草稿', published: '已发布', archived: '已归档' }
  return map[status] || status
}

function truncate(text, length) {
  if (!text) return ''
  return text.length > length ? text.substring(0, length) + '...' : text
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  return new Date(timestamp).toLocaleString()
}

function openUrl(url) {
  if (url) {
    window.open(url, '_blank')
  }
}

async function handleStartInstance() {
  if (!content.value?.template_id) return

  startingInstance.value = true
  try {
    const template = await templateApi.getDetail(content.value.template_id)
    selectedTemplate.value = template
    showStartModal.value = true
  } catch (err) {
    console.error('启动实例失败:', err)
    message.error('启动实例失败')
  } finally {
    startingInstance.value = false
  }
}

function onInstanceStarted() {
  message.success('实例已加入创建队列，请通过右上角队列按钮查看进度')
  router.push('/app/my-instances')
}

async function viewContent(item) {
  loading.value = true
  try {
    const res = await contentApi.getDetail(item.id)
    content.value = res
    viewType.value = 'content'
    selectedKeys.value = [item.id]
  } catch (err) {
    console.error('Failed to fetch content detail:', err)
  } finally {
    loading.value = false
  }
}

function findNodeById(node, id) {
  if (node.id === id) return node
  if (node.children) {
    for (const child of node.children) {
      const found = findNodeById(child, id)
      if (found) return found
    }
  }
  return null
}

async function handleSelect(keys) {
  if (keys.length === 0) return

  const selectedId = keys[0]
  selectedKeys.value = [selectedId]

  const node = findNodeById(categoryTree.value, selectedId)
  if (node) {
    currentCategory.value = {
      id: node.id,
      name: node.name,
      description: node.description,
      created_at: node.created_at
    }

    loading.value = true
    try {
      const res = await contentApi.getByCategory(selectedId)
      currentCategoryContents.value = res.items || []
      viewType.value = 'category'
    } catch (err) {
      console.error('Failed to fetch category contents:', err)
    } finally {
      loading.value = false
    }
  }
}

function handleExpand(keys) {
  expandedKeys.value = keys
}

async function fetchData() {
  loading.value = true
  try {
    const id = String(route.params.id || '')
    const isCategory = route.query.type === 'category'

    if (isCategory) {
      const res = await contentApi.getCategoryTree(id)
      categoryTree.value = res

      currentCategory.value = {
        id: res.id,
        name: res.name,
        description: res.description,
        created_at: res.created_at
      }

      const contentsRes = await contentApi.getByCategory(id)
      currentCategoryContents.value = contentsRes.items || []

      selectedKeys.value = [id]
      expandedKeys.value = getAllKeys(res)
      viewType.value = 'category'
    } else {
      const res = await contentApi.getDetail(id)
      content.value = res
      viewType.value = 'content'
      selectedKeys.value = [id]
    }
  } catch (err) {
    console.error('Failed to fetch data:', err)
  } finally {
    loading.value = false
  }
}

function getAllKeys(node) {
  const keys = [node.id]
  if (node.children) {
    for (const child of node.children) {
      keys.push(...getAllKeys(child))
    }
  }
  return keys
}

onMounted(() => {
  fetchData()
})

onBeforeUnmount(() => {
  if (markdownPreviewRef.value) {
    markdownPreviewRef.value.innerHTML = ''
  }
})
</script>

<style scoped>
.content-detail {
  padding-bottom: 24px;
}

.detail-header {
  margin-bottom: 24px;
}

.detail-header-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #666;
  text-decoration: none;
  transition: color 0.2s;
}

.back-link:hover {
  color: #18a058;
}

.detail-layout {
  display: flex;
  gap: 24px;
}

.detail-main {
  flex: 1;
  min-width: 0;
}

.category-view,
.content-view {
  background: white;
  border-radius: 8px;
  padding: 32px;
  border: 1px solid #e0e0e6;
}

.category-header {
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid #e0e0e6;
}

.category-header .title {
  font-size: 28px;
  font-weight: 600;
  color: #333;
  margin: 0 0 16px 0;
}

.category-header .description {
  font-size: 14px;
  color: #666;
  line-height: 1.6;
}

.category-description {
  padding: 16px 0;
  margin-bottom: 16px;
  /* border-bottom: 1px solid #f0f0f0; */
}

.category-description p {
  margin: 0;
  font-size: 15px;
  color: #333;
  line-height: 1.8;
}

.divider {
  display: flex;
  align-items: center;
  gap: 16px;
  margin: 24px 0 16px;
}

.divider-line {
  flex: 1;
  height: 1px;
  background: #e0e0e6;
}

.divider-text {
  font-size: 14px;
  color: #999;
  white-space: nowrap;
  font-weight: 500;
}

.contents-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.content-item {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #e8e8ec;
}

.content-item:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  border-color: #18a058;
}

.content-thumb {
  height: 100px;
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  padding: 12px;
  position: relative;
}

.content-thumb::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 40px;
  background: linear-gradient(to top, rgba(255, 255, 255, 0.9), transparent);
}

.content-info {
  padding: 16px;
}

.content-info h4 {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  margin: 0 0 8px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 42px;
}

.content-info p {
  font-size: 13px;
  color: #888;
  margin: 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.content-item .content-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.content-item .content-meta .time {
  font-size: 12px;
  color: #999;
}

.article-header {
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid #e0e0e6;
}

.title {
  font-size: 28px;
  font-weight: 600;
  color: #333;
  margin: 16px 0;
  line-height: 1.4;
}

.meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: #999;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.status-badge.published {
  background: #e6f7ef;
  color: #18a058;
}

.article-body {
  font-size: 16px;
  line-height: 1.8;
  color: #333;
}

.markdown-preview :deep(.vditor-reset) {
  padding: 0;
}

.article-body :deep(.ql-toolbar) {
  display: none;
}

.article-body :deep(.ql-container) {
  border: none;
  font-size: 16px;
}

.article-body :deep(.ql-editor) {
  padding: 0;
  line-height: 1.8;
}

.article-body :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  margin: 16px 0;
}

.article-body :deep(video) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  margin: 16px 0;
}

.article-body :deep(p) {
  margin-bottom: 16px;
}

.article-body :deep(a) {
  color: #18a058;
  text-decoration: none;
}

.article-body :deep(a:hover) {
  text-decoration: underline;
}

.article-body :deep(blockquote) {
  border-left: 4px solid #18a058;
  padding-left: 16px;
  margin: 16px 0;
  color: #666;
  background: #f9f9f9;
  padding: 12px 16px;
}

.article-body :deep(pre) {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
}

.article-body :deep(code) {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

.article-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-sidebar {
  width: 280px;
  flex-shrink: 0;
  background: white;
  border-radius: 8px;
  border: 1px solid #e0e0e6;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
  position: sticky;
  top: 0;
}

.sidebar-header {
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 600;
  color: #333;
  border-bottom: 1px solid #e0e0e6;
}

.sidebar-wrapper {
  padding: 12px;
}

@media (max-width: 1024px) {
  .detail-layout {
    flex-direction: column;
  }

  .detail-sidebar {
    width: 100%;
    position: static;
    max-height: none;
  }

  .contents-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .contents-grid {
    grid-template-columns: 1fr;
  }
}
</style>
