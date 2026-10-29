<template>
  <div class="categories">
    <n-card title="分类管理">
      <template #header-extra>
        <n-button type="primary" :disabled="!ready" @click="editCourse()">新建课程</n-button>
      </template>
      <p class="management-description">
        课程为一级目录，教学文章为二级目录；发布后的文章会按相同结构展示在内容浏览中。
      </p>
      <n-alert v-if="error" type="error" :bordered="false" class="library-error">
        {{ error }}
        <n-button size="small" class="load-retry" @click="load">重试</n-button>
      </n-alert>
      <div class="course-library-toolbar">
        <span>{{ courses.length }} 门课程 · {{ articles.length }} 篇教学文章</span>
        <n-input
          v-model:value="searchKey"
          placeholder="搜索课程或文章"
          clearable
          class="course-search"
          aria-label="搜索课程或文章"
        />
      </div>
      <n-spin :show="loading || mutating">
        <div class="course-management-tree" aria-label="课程与教学文章">
          <section v-for="course in filteredCourses" :key="course.id" class="course-group">
            <header class="course-group-header">
              <n-button
                quaternary
                circle
                size="small"
                :aria-label="`${isExpanded(course.id) ? '收起' : '展开'}${course.name}`"
                :aria-expanded="isExpanded(course.id)"
                @click="toggleCourse(course.id)"
              >
                <template #icon>
                  <n-icon>
                    <ChevronDownOutline v-if="isExpanded(course.id)" />
                    <ChevronForwardOutline v-else />
                  </n-icon>
                </template>
              </n-button>
              <div class="course-group-info">
                <div class="course-group-title">
                  <n-icon><BookOutline /></n-icon>
                  <h3>{{ course.name }}</h3>
                  <n-tag
                    size="small"
                    :bordered="false"
                    :type="course.is_active === false ? 'default' : 'success'"
                  >
                    {{ course.is_active === false ? '已停用' : '启用' }}
                  </n-tag>
                  <span>{{ course.articles.length }} 篇文章</span>
                </div>
                <p v-if="course.description">{{ course.description }}</p>
              </div>
              <div class="course-actions">
                <n-button size="small" :disabled="mutating" @click="newArticle(course)">
                  新增文章
                </n-button>
                <n-button size="small" :disabled="mutating" @click="openAssociation(course)">
                  关联文章
                </n-button>
                <n-button size="small" quaternary :disabled="mutating" @click="editCourse(course)">
                  编辑课程
                </n-button>
                <n-popconfirm @positive-click="removeCourse(course)">
                  <template #trigger>
                    <n-button size="small" quaternary type="error" :disabled="mutating">
                      删除课程
                    </n-button>
                  </template>
                  删除课程及其目录？教学文章会保留在内容管理中。
                </n-popconfirm>
              </div>
            </header>
            <div v-if="isExpanded(course.id)" class="course-articles">
              <div
                v-for="article in matchingArticles(course)"
                :key="article.id"
                class="course-article-row"
              >
                <div class="course-article-title">
                  <strong>{{ article.title }}</strong>
                  <span>{{ contentTypes.getTypeName(article.content_type) }}</span>
                </div>
                <n-tag size="small" :bordered="false" :type="articleStatus(article).type">
                  {{ articleStatus(article).label }}
                </n-tag>
                <div class="article-actions">
                  <n-button
                    size="small"
                    quaternary
                    :disabled="mutating"
                    @click="editArticle(article)"
                  >
                    编辑文章
                  </n-button>
                  <n-button
                    size="small"
                    quaternary
                    :disabled="mutating"
                    @click="removeArticle(course, article)"
                  >
                    移出课程
                  </n-button>
                </div>
              </div>
              <div v-if="!matchingArticles(course).length" class="empty-course">
                {{
                  searchKey ? '没有匹配的教学文章' : '课程中暂无文章，可新增文章或关联已有文章。'
                }}
              </div>
            </div>
          </section>
          <n-empty
            v-if="!loading && !filteredCourses.length"
            :description="searchKey ? '没有匹配的课程或文章' : '暂无课程，先创建一门课程'"
          />
        </div>
      </n-spin>
      <div v-if="unassigned.length" class="unassigned-note">
        <span>{{ unassigned.length }} 篇教学文章尚未分配课程</span>
        <n-button text size="small" @click="router.push('/admin/contents?course=unassigned')">
          前往内容管理 →
        </n-button>
      </div>
    </n-card>

    <n-modal
      :show="showCourseModal"
      preset="card"
      :title="editingCourseId ? '编辑课程' : '新建课程'"
      class="course-form-modal"
      :closable="!savingCourse"
      :mask-closable="!savingCourse"
      :close-on-esc="!savingCourse"
      @update:show="
        (value) => {
          if (!savingCourse) showCourseModal = value
        }
      "
    >
      <n-form ref="courseFormRef" :model="courseForm" :rules="courseRules" :disabled="savingCourse">
        <n-form-item label="课程名称" path="name">
          <n-input v-model:value="courseForm.name" placeholder="输入课程名称" />
        </n-form-item>
        <n-form-item label="课程说明">
          <n-input
            v-model:value="courseForm.description"
            type="textarea"
            :rows="3"
            placeholder="简要介绍课程内容"
          />
        </n-form-item>
        <n-form-item label="排序">
          <n-input-number v-model:value="courseForm.sort_order" :min="0" />
        </n-form-item>
        <n-form-item label="启用课程">
          <n-switch v-model:value="courseForm.is_active" />
        </n-form-item>
      </n-form>
      <n-alert v-if="courseError" type="error" :bordered="false">{{ courseError }}</n-alert>
      <template #action>
        <n-space justify="end">
          <n-button :disabled="savingCourse" @click="showCourseModal = false">取消</n-button>
          <n-button type="primary" :loading="savingCourse" @click="saveCourse">保存课程</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      :show="showAssociation"
      preset="card"
      :title="`${associationCourse?.name || '课程'} · 关联文章`"
      class="course-association-modal"
      :closable="!binding"
      :mask-closable="!binding"
      :close-on-esc="!binding"
      @update:show="
        (value) => {
          if (!binding) showAssociation = value
        }
      "
    >
      <p class="management-description">选择已有教学文章加入课程；同一篇文章可以关联多个课程。</p>
      <n-input
        v-model:value="associationSearch"
        placeholder="搜索文章标题"
        clearable
        aria-label="搜索待关联文章"
        class="association-search"
      />
      <n-data-table
        :columns="associationColumns"
        :data="associationArticles"
        :pagination="{ pageSize: 10 }"
        :bordered="false"
        :scroll-x="560"
        :row-key="(row) => row.id"
        v-model:checked-row-keys="checkedArticleIds"
      />
      <n-alert v-if="associationError" type="error" :bordered="false" class="association-error">
        {{ associationError }}
      </n-alert>
      <template #action>
        <n-space justify="end">
          <n-button :disabled="binding" @click="showAssociation = false">取消</n-button>
          <n-button
            type="primary"
            :loading="binding"
            :disabled="!checkedArticleIds.length"
            @click="associateArticles"
          >
            关联 {{ checkedArticleIds.length || '' }} 篇文章
          </n-button>
        </n-space>
      </template>
    </n-modal>
    <ArticleEditorModal
      v-model:show="showEditor"
      :article-id="editingArticleId"
      :default-course-id="defaultCourseId"
      :courses="categories"
      @saved="load"
    />
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NEmpty,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NInputNumber,
  NModal,
  NPopconfirm,
  NSpace,
  NSpin,
  NSwitch,
  NTag,
  useMessage
} from 'naive-ui'
import { BookOutline, ChevronDownOutline, ChevronForwardOutline } from '@vicons/ionicons5'
import { bindingApi, categoryApi } from '@/api'
import ArticleEditorModal from '@/components/ArticleEditorModal.vue'
import { useCourseLibrary } from '@/composables/useCourseLibrary'
import { useContentTypeStore } from '@/stores/contentType'
import { articleStatus, deleteCourse, removeArticleFromCourse } from '@/utils/courseManagement'

const router = useRouter()
const message = useMessage()
const contentTypes = useContentTypeStore()
const { categories, articles, courses, unassigned, loading, error, ready, load } =
  useCourseLibrary()
const searchKey = ref('')
const collapsedIds = ref([])
const mutating = ref(false)
const filteredCourses = computed(() => {
  const keyword = searchKey.value.trim().toLocaleLowerCase()
  return courses.value.filter(
    (course) =>
      course.name.toLocaleLowerCase().includes(keyword) ||
      course.articles.some((article) => article.title.toLocaleLowerCase().includes(keyword))
  )
})
function matchingArticles(course) {
  const keyword = searchKey.value.trim().toLocaleLowerCase()
  return course.name.toLocaleLowerCase().includes(keyword)
    ? course.articles
    : course.articles.filter((article) => article.title.toLocaleLowerCase().includes(keyword))
}
function isExpanded(id) {
  return !collapsedIds.value.includes(id)
}
function toggleCourse(id) {
  collapsedIds.value = collapsedIds.value.includes(id)
    ? collapsedIds.value.filter((key) => key !== id)
    : [...collapsedIds.value, id]
}
watch(searchKey, () => {
  collapsedIds.value = []
})

const showEditor = ref(false)
const editingArticleId = ref(null)
const defaultCourseId = ref(null)
function newArticle(course) {
  editingArticleId.value = null
  defaultCourseId.value = course.id
  showEditor.value = true
}
function editArticle(article) {
  editingArticleId.value = article.id
  defaultCourseId.value = null
  showEditor.value = true
}

const showCourseModal = ref(false)
const editingCourseId = ref(null)
const courseFormRef = ref(null)
const courseForm = ref({})
const courseError = ref('')
const savingCourse = ref(false)
const courseRules = {
  name: [{ required: true, whitespace: true, message: '请输入课程名称', trigger: 'blur' }]
}
function editCourse(course = null) {
  editingCourseId.value = course?.id || null
  courseForm.value = {
    name: course?.name || '',
    description: course?.description || '',
    sort_order: course?.sort_order || 0,
    is_active: course?.is_active !== false
  }
  courseError.value = ''
  showCourseModal.value = true
}
async function saveCourse() {
  try {
    await courseFormRef.value?.validate()
  } catch {
    return
  }
  savingCourse.value = true
  courseError.value = ''
  try {
    const payload = { ...courseForm.value, name: courseForm.value.name.trim(), parent_id: null }
    if (editingCourseId.value) await categoryApi.update(editingCourseId.value, payload)
    else await categoryApi.create(payload)
    showCourseModal.value = false
    message.success('课程已保存')
    await load()
  } catch (cause) {
    courseError.value = cause.response?.data?.message || '课程保存失败'
  } finally {
    savingCourse.value = false
  }
}
async function removeCourse(course) {
  if (mutating.value) return
  mutating.value = true
  try {
    await deleteCourse(categoryApi, course)
    message.success('课程已删除，教学文章已保留')
  } catch (cause) {
    message.error(cause.response?.data?.message || '删除课程未完成，请刷新后重试')
  } finally {
    await load()
    mutating.value = false
  }
}
async function removeArticle(course, article) {
  if (mutating.value) return
  mutating.value = true
  try {
    await removeArticleFromCourse(bindingApi, course, article)
    message.success('文章已移出课程')
  } catch (cause) {
    message.error(cause.response?.data?.message || '移出课程未完成，请重试')
  } finally {
    await load()
    mutating.value = false
  }
}

const showAssociation = ref(false)
const associationCourse = ref(null)
const associationSearch = ref('')
const checkedArticleIds = ref([])
const binding = ref(false)
const associationError = ref('')
const associationArticles = computed(() => {
  const assigned = new Set(associationCourse.value?.articles.map((article) => article.id) || [])
  const keyword = associationSearch.value.trim().toLocaleLowerCase()
  return articles.value.filter(
    (article) => !assigned.has(article.id) && article.title.toLocaleLowerCase().includes(keyword)
  )
})
const associationColumns = [
  { type: 'selection', width: 48, disabled: () => binding.value },
  { title: '文章标题', key: 'title', minWidth: 260, ellipsis: { tooltip: true } },
  {
    title: '类型',
    key: 'content_type',
    width: 110,
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
  }
]
function openAssociation(course) {
  associationCourse.value = course
  associationSearch.value = ''
  checkedArticleIds.value = []
  associationError.value = ''
  showAssociation.value = true
}
async function associateArticles() {
  if (binding.value || !checkedArticleIds.value.length) return
  binding.value = true
  associationError.value = ''
  try {
    await bindingApi.batchBind(associationCourse.value.id, { content_ids: checkedArticleIds.value })
    showAssociation.value = false
    message.success('教学文章已关联到课程')
    await load()
  } catch (cause) {
    associationError.value = cause.response?.data?.message || '关联文章失败，请重试'
  } finally {
    binding.value = false
  }
}
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
.course-library-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 20px;
}
.course-library-toolbar > span {
  color: var(--mixm-muted);
  font-size: 13px;
}
.course-search {
  width: 260px;
  max-width: 100%;
}
.course-group {
  border: 1px solid var(--mixm-border);
  border-radius: 8px;
  overflow: hidden;
}
.course-group + .course-group {
  margin-top: 16px;
}
.course-group-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 18px 16px;
  background: var(--mixm-paper);
}
.course-group-info {
  flex: 1;
  min-width: 0;
}
.course-group-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.course-group-title > .n-icon {
  color: var(--mixm-muted);
}
.course-group-title h3 {
  margin: 0;
  font-family: var(--mixm-font-body);
  font-size: 15px;
  font-weight: 600;
  overflow-wrap: anywhere;
}
.course-group-title > span {
  color: var(--mixm-muted);
  font-size: 12px;
}
.course-group-info p {
  margin: 6px 0 0;
  color: var(--mixm-muted);
  font-size: 12px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.course-actions,
.article-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.course-articles {
  border-top: 1px solid var(--mixm-border);
}
.course-article-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  padding: 14px 18px;
}
.course-article-row + .course-article-row {
  border-top: 1px solid var(--mixm-border);
}
.course-article-title {
  flex: 1;
  min-width: 160px;
}
.course-article-title strong {
  font-size: 13px;
  font-weight: 500;
  overflow-wrap: anywhere;
}
.course-article-title > span {
  color: var(--mixm-muted);
  font-size: 12px;
  margin-left: 12px;
}
.empty-course {
  padding: 24px 18px;
  color: var(--mixm-muted);
  font-size: 13px;
}
.unassigned-note {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--mixm-border);
  color: var(--mixm-muted);
  font-size: 13px;
}
.course-form-modal {
  width: min(520px, calc(100vw - 32px));
}
.course-association-modal {
  width: min(900px, calc(100vw - 32px));
}
.course-association-modal :deep(.n-card__content) {
  max-height: 65dvh;
  overflow-y: auto;
}
.association-search {
  margin-bottom: 16px;
}
.association-error {
  margin-top: 16px;
}
@media (max-width: 1100px) {
  .course-group-header {
    flex-wrap: wrap;
  }
  .course-actions {
    flex-basis: 100%;
    padding-left: 40px;
  }
}
@media (max-width: 600px) {
  .course-article-row {
    gap: 10px;
  }
  .article-actions {
    flex-basis: 100%;
    justify-content: flex-end;
  }
  .course-actions {
    padding-left: 0;
  }
}
</style>
