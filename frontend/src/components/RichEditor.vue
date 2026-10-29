<template>
  <div class="dual-editor">
    <div class="editor-tabs">
      <n-radio-group :value="editorMode" size="small" @update:value="handleModeChange">
       
        <n-radio-button value="md">Markdown</n-radio-button>
         <n-radio-button value="web">超文本</n-radio-button>
      </n-radio-group>
    </div>
    
    <div class="editor-content markdown-content" v-show="editorMode === 'md'">
      <div ref="vditorRef" class="vditor-container"></div>
    </div>
    <div class="editor-content" v-show="editorMode === 'web'">
      <QuillEditor 
        contentType="html"
        :options="editorOptions"
        @ready="onEditorReady"
      />
    </div>
    <n-modal v-model:show="showConfirmModal" preset="dialog" title="确认切换" style="width: 400px;">
      <div>切换编辑器将丢失当前内容，是否继续？</div>
      <template #action>
        <n-button @click="cancelModeChange">取消</n-button>
        <n-button type="error" @click="confirmModeChange">确定切换</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { API_BASE_URL } from '@/api/config'
import { apiUrl } from '@/api/protocol'
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import Vditor from 'vditor'
import { VDITOR_ASSET_BASE } from '@/utils/vditorAssets'
import 'vditor/dist/index.css'
import { contentApi } from '@/api'
import { useMessage, NRadioGroup, NRadioButton, NInput, NModal, NButton } from 'naive-ui'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue'])

const message = useMessage()
const vditorRef = ref(null)
const uploading = ref(false)
const editorMode = ref('md')
const markdownContent = ref('')
const showConfirmModal = ref(false)
const pendingMode = ref(null)
const hasContent = ref(false)
let vditorInstance = null
let htmlQuill = null
let htmlContent = ''

const editorOptions = {
  theme: 'snow',
  modules: {
    toolbar: {
      container: [
        [{ 'header': [1, 2, 3, 4, 5, 6, false] }],
        ['bold', 'italic', 'underline', 'strike'],
        [{ 'color': [] }, { 'background': [] }],
        [{ 'script': 'sub'}, { 'script': 'super' }],
        [{ 'indent': '-1'}, { 'indent': '+1' }],
        [{ 'list': 'ordered'}, { 'list': 'bullet' }],
        [{ 'direction': 'rtl' }],
        [{ 'align': [] }],
        ['blockquote', 'code-block'],
        ['link', 'image', 'video'],
        ['clean']
      ]
    }
  }
}

function initFromProps() {
  const content = props.modelValue || ''
  if (content.startsWith('[md]')) {
    editorMode.value = 'md'
    markdownContent.value = content.slice(4)
    nextTick(() => initVditor())
  } else if (content.startsWith('[web]')) {
    editorMode.value = 'web'
    markdownContent.value = ''
    htmlContent = content.slice(5)
    nextTick(renderHtml)
  } else {
    editorMode.value = 'md'
    markdownContent.value = content
    nextTick(() => initVditor())
  }
}

function initVditor() {
  if (!vditorRef.value || editorMode.value !== 'md') return
  if (vditorInstance) {
    vditorInstance.setValue(markdownContent.value)
    return
  }
  
  vditorInstance = new Vditor(vditorRef.value, {
    cdn: VDITOR_ASSET_BASE,
    id: 'vditor-editor',
    cache: {
      enable: false
    },
    value: markdownContent.value,
    height: 700,
    mode: 'wysiwyg',
    placeholder: '请输入 Markdown 内容...',
    upload: {
      url: apiUrl(API_BASE_URL, '/contents/files'),
      linkToImgUrl: apiUrl(API_BASE_URL, '/contents/files'),
      handler: (files) => {
        handleVditorUpload(files[0])
      }
    },
    resize: {
      enable: true,
      position: 'bottom'
    },
    input: (value) => {
      markdownContent.value = value
    }
  })
}

function destroyVditor() {
  if (vditorInstance) {
    vditorInstance.destroy()
    vditorInstance = null
  }
}

onMounted(() => {
  initFromProps()
})

onBeforeUnmount(() => {
  destroyVditor()
  htmlQuill = null
})

watch(editorMode, (newMode) => {
  if (newMode === 'web') {
    destroyVditor()
    nextTick(renderHtml)
  } else {
    nextTick(() => initVditor())
  }
})

function handleModeChange(newMode) {
  if (newMode === editorMode.value) {
    return
  }
  
  const currentContent = editorMode.value === 'web' 
    ? htmlQuill?.root.innerHTML
    : markdownContent.value
  
  const hasCurrentContent = editorMode.value === 'web'
    ? (currentContent && currentContent !== '<p><br></p>')
    : (currentContent && currentContent.trim() !== '')
  
  if (hasCurrentContent) {
    hasContent.value = true
    pendingMode.value = newMode
    showConfirmModal.value = true
  } else {
    applyModeChange(newMode)
  }
}

function confirmModeChange() {
  showConfirmModal.value = false
  hasContent.value = false
  if (pendingMode.value) {
    applyModeChange(pendingMode.value)
    pendingMode.value = null
  }
}

function cancelModeChange() {
  showConfirmModal.value = false
  pendingMode.value = null
}

watch(markdownContent, (val) => {
  if (editorMode.value !== 'md') return
  const fullContent = val ? `[md]${val}` : ''
  emit('update:modelValue', fullContent)
})

function applyModeChange(mode) {
  htmlContent = ''
  markdownContent.value = ''
  editorMode.value = mode
  emit('update:modelValue', '')
}

// 等待富文本编辑器就绪后再写入正文，避免首次打开时访问尚未初始化的实例。
function renderHtml() {
  if (htmlQuill && editorMode.value === 'web') htmlQuill.root.innerHTML = htmlContent
}

function onEditorReady(quill) {
  htmlQuill = quill
  renderHtml()
  
  const toolbar = quill.getModule('toolbar')
  if (toolbar) {
    toolbar.addHandler('image', () => handleImageUpload())
    toolbar.addHandler('video', () => handleVideoUpload())
  }
  
  quill.on('text-change', () => {
    const html = quill.root.innerHTML
    if (editorMode.value === 'web') {
      htmlContent = html
      const fullContent = html && html !== '<p><br></p>' ? `[web]${html}` : ''
      emit('update:modelValue', fullContent)
    }
  })
}

function handleImageUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.onchange = async (e) => {
    const file = e.target.files[0]
    if (file) {
      await uploadFile(file, 'image')
    }
  }
  input.click()
}

function handleVideoUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'video/*'
  input.onchange = async (e) => {
    const file = e.target.files[0]
    if (file) {
      await uploadFile(file, 'video')
    }
  }
  input.click()
}

async function uploadFile(file, type) {
  if (uploading.value) return
  uploading.value = true
  try {
    const res = await contentApi.uploadFile(file)
    const { url } = res
    
    if (editorMode.value === 'web') {
      const quill = htmlQuill
      if (!quill) return
      
      const range = quill.getSelection(true)
      quill.insertEmbed(range.index, type, url)
      quill.setSelection(range.index + 1)
    } else if (vditorInstance) {
      vditorInstance.insertValue(`![${file.name}](${url})`)
    }
    
    message.success(`${type === 'image' ? '图片' : '视频'}上传成功`)
  } catch (err) {
    console.error(err)
    message.error(`${type === 'image' ? '图片' : '视频'}上传失败`)
  } finally {
    uploading.value = false
  }
}

async function handleVditorUpload(file) {
  await uploadFile(file, 'image')
}

function setContent(content) {
  if (editorMode.value === 'web') {
    htmlContent = content || ''
    renderHtml()
  } else {
    if (vditorInstance) {
      vditorInstance.setValue(content || '')
    } else {
      markdownContent.value = content || ''
    }
  }
}

defineExpose({
  setContent,
  getQuill: () => htmlQuill,
  getVditor: () => vditorInstance
})
</script>

<style scoped>
.dual-editor {
  display: flex;
  flex-direction: column;
}

.editor-tabs {
  margin-bottom: 8px;
}

.editor-content {
  flex: 1;
}

.markdown-content {
  min-height: 300px;
}

.vditor-container {
  min-height: 300px;
}

.markdown-content :deep(.n-input__textarea-el) {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 14px;
}

:deep(.ql-container) {
  min-height: 250px;
}

:deep(.ql-editor) {
  min-height: 250px;
}
</style>
