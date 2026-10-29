<template>
  <n-modal
    :show="show"
    preset="card"
    :title="articleId ? '编辑教学文章' : '新建教学文章'"
    class="article-editor-modal"
    :closable="!saving"
    :mask-closable="!saving"
    :close-on-esc="!saving"
    @update:show="updateShow"
  >
    <n-alert v-if="loadError" type="error" :bordered="false">
      {{ loadError }}
      <n-button size="small" class="load-retry" @click="loadEditor">重试</n-button>
    </n-alert>
    <n-spin v-else :show="loading">
      <div class="article-editor-layout">
        <n-form
          ref="formRef"
          :model="form"
          :rules="rules"
          class="article-fields"
          :disabled="saving || loading"
        >
          <n-form-item label="文章标题" path="title">
            <n-input v-model:value="form.title" placeholder="输入教学文章标题" />
          </n-form-item>
          <n-form-item label="所属课程">
            <n-select
              v-model:value="form.course_ids"
              :options="courseOptions"
              multiple
              filterable
              clearable
              placeholder="选择课程，可关联多个课程"
            />
          </n-form-item>
          <p class="course-field-hint">
            未分配课程的文章保留在内容管理中，加入课程后会出现在对应目录下。
          </p>
          <n-form-item label="文章类型" path="content_type">
            <n-select v-model:value="form.content_type" :options="typeOptions" />
          </n-form-item>
          <n-form-item label="发布状态">
            <n-select v-model:value="form.status" :options="statusOptions" />
          </n-form-item>
          <n-form-item label="启用文章"><n-switch v-model:value="form.is_active" /></n-form-item>
          <n-form-item label="实验模板">
            <n-select
              v-model:value="form.template_id"
              :options="templateOptions"
              filterable
              clearable
              placeholder="可选，用于启动实例"
            />
          </n-form-item>
          <n-form-item label="资源链接">
            <n-input v-model:value="form.url" clearable placeholder="可选，补充外部资源" />
          </n-form-item>
        </n-form>
        <section class="article-body-editor" aria-label="教学文章正文">
          <div class="article-body-label">
            <strong>文章正文</strong>
            <span>支持 Markdown 和富文本</span>
          </div>
          <RichEditor v-if="!loading" v-model="form.content" />
        </section>
      </div>
    </n-spin>
    <template #action>
      <n-alert v-if="saveError" type="error" :bordered="false" class="article-save-error">
        {{ saveError }}
      </n-alert>
      <div class="article-editor-actions">
        <span>仅启用且已发布的课程文章会在内容浏览中展示。</span>
        <n-button :disabled="saving" @click="updateShow(false)">取消</n-button>
        <n-button
          type="primary"
          :loading="saving"
          :disabled="loading || !!loadError"
          @click="saveArticle"
        >
          保存文章
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSpin,
  NSwitch,
  useMessage
} from 'naive-ui'
import { bindingApi, contentApi, templateApi } from '@/api'
import { useContentTypeStore } from '@/stores/contentType'
import { articleCourses, readAllPages, syncArticleCourses } from '@/utils/courseManagement'
import RichEditor from '@/components/RichEditor.vue'

const props = defineProps({
  show: Boolean,
  articleId: { type: String, default: null },
  defaultCourseId: { type: String, default: null },
  courses: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:show', 'saved'])
const message = useMessage()
const contentTypes = useContentTypeStore()
const formRef = ref(null)
const form = ref(emptyArticle())
const templates = ref([])
const loading = ref(false)
const saving = ref(false)
const loadError = ref('')
const saveError = ref('')
const savedId = ref(null)
let version = 0

const rules = {
  title: [{ required: true, whitespace: true, message: '请输入文章标题', trigger: 'blur' }],
  content_type: [{ required: true, message: '请选择文章类型', trigger: 'change' }]
}
const courseOptions = computed(() =>
  props.courses.map((course) => ({
    label: `${course.name}${course.is_active === false ? '（已停用）' : ''}`,
    value: course.id
  }))
)
const typeOptions = computed(() =>
  contentTypes.contentTypes.map((type) => ({ label: type.name, value: type.code }))
)
const templateOptions = computed(() =>
  templates.value.map((template) => ({ label: template.name, value: template.id }))
)
const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '已发布', value: 'published' },
  { label: '已归档', value: 'archived' }
]

function emptyArticle() {
  return {
    title: '',
    content: '',
    content_type: 'article',
    status: 'draft',
    is_active: true,
    template_id: null,
    url: '',
    course_ids: []
  }
}

function updateShow(value) {
  if (!saving.value) emit('update:show', value)
}

async function loadEditor() {
  const request = ++version
  loading.value = true
  loadError.value = ''
  saveError.value = ''
  savedId.value = props.articleId
  form.value = {
    ...emptyArticle(),
    course_ids: props.defaultCourseId ? [props.defaultCourseId] : []
  }
  try {
    const [detail, options] = await Promise.all([
      props.articleId ? contentApi.getDetail(props.articleId) : null,
      readAllPages((params) => templateApi.getList(params))
    ])
    if (request !== version) return
    templates.value = options
    if (detail)
      form.value = {
        title: detail.title || '',
        content: detail.content || '',
        content_type: detail.content_type || 'article',
        status: detail.status || 'draft',
        is_active: detail.is_active !== false,
        template_id: detail.template_id || null,
        url: detail.url || '',
        course_ids: articleCourses(props.courses, detail).map((course) => course.id)
      }
  } catch (error) {
    if (request === version)
      loadError.value = error.response?.data?.message || error.message || '文章加载失败'
  } finally {
    if (request === version) loading.value = false
  }
}

// 文章与课程使用现有接口保存，记录新文章 ID，绑定失败后重试不会重复新建。
async function saveArticle() {
  if (saving.value || loading.value || loadError.value) return
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  saving.value = true
  saveError.value = ''
  let articleSaved = false
  try {
    const { course_ids, ...fields } = form.value
    const payload = { ...fields, title: fields.title.trim() }
    if (savedId.value) await contentApi.update(savedId.value, payload)
    else {
      const article = await contentApi.create(payload)
      savedId.value = article.id
    }
    articleSaved = true
    await syncArticleCourses(bindingApi, props.courses, savedId.value, course_ids)
    message.success('教学文章已保存')
    emit('saved')
    emit('update:show', false)
  } catch (error) {
    const detail = error.response?.data?.message || error.message || '请重试'
    saveError.value = articleSaved
      ? `文章已保存，课程归属尚未保存完成，请再次点击保存重试。${detail}`
      : `保存失败：${detail}`
    if (articleSaved) emit('saved')
  } finally {
    saving.value = false
  }
}

watch(
  [() => props.show, () => props.articleId],
  ([show]) => {
    if (show) loadEditor()
    else version++
  },
  { immediate: true }
)
onBeforeUnmount(() => {
  version++
})
</script>

<style scoped>
.load-retry {
  margin-left: 12px;
}
.article-editor-modal {
  width: min(1320px, calc(100vw - 48px));
}
.article-editor-modal :deep(.n-card__content) {
  max-height: 70dvh;
  overflow-y: auto;
}
.article-editor-layout {
  display: grid;
  grid-template-columns: 270px minmax(0, 1fr);
  gap: 28px;
}
.article-fields {
  min-width: 0;
}
.course-field-hint {
  margin: -12px 0 20px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--mixm-muted);
}
.article-body-editor {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 520px;
}
.article-body-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.article-body-label strong {
  font-size: 14px;
  font-weight: 500;
}
.article-body-label span {
  color: var(--mixm-muted);
  font-size: 12px;
}
.article-editor-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}
.article-editor-actions > span {
  margin-right: auto;
  color: var(--mixm-muted);
  font-size: 12px;
}
.article-save-error {
  margin-bottom: 14px;
}
@media (max-width: 800px) {
  .article-editor-layout {
    grid-template-columns: minmax(0, 1fr);
  }
  .article-editor-actions {
    flex-wrap: wrap;
  }
  .article-editor-actions > span {
    flex-basis: 100%;
  }
  .article-body-editor {
    min-height: 400px;
  }
}
</style>
