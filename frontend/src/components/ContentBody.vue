<template>
  <div class="content-body">
    <n-alert v-if="renderError" type="warning" title="正文显示失败">{{ renderError }}</n-alert>
    <div v-if="mode === 'md'" ref="markdownPreview" class="markdown-preview"></div>
    <QuillEditor
      v-else-if="mode === 'web'"
      content-type="html"
      :options="htmlOptions"
      @ready="onHtmlReady"
    />
    <n-empty v-else description="这篇文章暂未填写正文" />
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { NAlert, NEmpty } from 'naive-ui'
import { QuillEditor } from '@vueup/vue-quill'
import VditorPreview from 'vditor/dist/method.min'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import 'vditor/dist/index.css'
import { VDITOR_ASSET_BASE } from '@/utils/vditorAssets'

const props = defineProps({ value: { type: String, default: '' } })
const markdownPreview = ref(null)
let htmlQuill = null
const renderError = ref('')
let renderVersion = 0

const mode = computed(() => {
  if (!props.value.trim()) return ''
  if (props.value.startsWith('[md]')) return 'md'
  if (props.value.startsWith('[web]')) return 'web'
  return /<\/?[a-z][\s\S]*>/i.test(props.value.trim()) ? 'web' : 'md'
})
const text = computed(() => {
  if (props.value.startsWith('[md]')) return props.value.slice(4)
  if (props.value.startsWith('[web]')) return props.value.slice(5)
  return props.value
})
const htmlOptions = { theme: 'snow', readOnly: true, modules: { toolbar: false } }

function renderHtml() {
  const quill = htmlQuill
  if (!quill) return
  quill.root.innerHTML = text.value
  quill.enable(false)
}

function onHtmlReady(quill) {
  htmlQuill = quill
  renderHtml()
}

// 每篇文章独立渲染，快速切换时不显示上一篇正文的异步错误。
watch(
  [mode, text],
  async () => {
    const version = ++renderVersion
    if (mode.value !== 'web') htmlQuill = null
    renderError.value = ''
    await nextTick()
    if (version !== renderVersion) return
    if (mode.value === 'web') return renderHtml()
    if (mode.value !== 'md' || !markdownPreview.value) return
    try {
      await VditorPreview.preview(markdownPreview.value, text.value, {
        mode: 'light',
        cdn: VDITOR_ASSET_BASE,
        theme: { path: `${VDITOR_ASSET_BASE}/dist/css/content-theme` }
      })
    } catch (error) {
      if (version === renderVersion) renderError.value = '请刷新页面后重试。'
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  htmlQuill = null
  renderVersion++
})
</script>

<style scoped>
.content-body {
  color: var(--mixm-ink);
  font-size: 15px;
  line-height: 1.9;
  overflow-wrap: anywhere;
}
.markdown-preview :deep(.vditor-reset) {
  padding: 0;
}
.content-body :deep(.ql-toolbar) {
  display: none;
}
.content-body :deep(.ql-container) {
  border: 0;
  font: inherit;
}
.content-body :deep(.ql-editor) {
  padding: 0;
  line-height: 1.9;
  overflow-wrap: anywhere;
}
.content-body :deep(img),
.content-body :deep(video) {
  max-width: 100%;
  height: auto;
}
.content-body :deep(pre) {
  max-width: 100%;
  overflow-x: auto;
  border-radius: 6px;
}
.content-body :deep(a) {
  color: var(--mixm-accent);
}
</style>
