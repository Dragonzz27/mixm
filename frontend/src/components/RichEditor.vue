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
        ref="quillEditorRef"
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
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import Vditor from 'vditor'
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
const quillEditorRef = ref(null)
const vditorRef = ref(null)
const uploading = ref(false)
const editorMode = ref('md')
const markdownContent = ref('')
const showConfirmModal = ref(false)
const pendingMode = ref(null)
const hasContent = ref(false)
let vditorInstance = null

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
    markdownContent.value = content.split('[md]')[1] || ''
    nextTick(() => initVditor())
  } else if (content.startsWith('[web]')) {
    editorMode.value = 'web'
    markdownContent.value = ''
    nextTick(() => {
      const quill = quillEditorRef.value?.getQuill()
      if (quill) {
        quill.root.innerHTML = content.split('[web]')[1] || ''
      }
    })
  } else {
    editorMode.value = 'md'
    markdownContent.value = content
    nextTick(() => initVditor())
  }
}

function initVditor() {
  if (vditorInstance) {
    vditorInstance.setValue(markdownContent.value)
    return
  }
  
  vditorInstance = new Vditor(vditorRef.value, {
    id: 'vditor-editor',
    cache: {
      enable: false
    },
    value: markdownContent.value,
    height: 700,
    mode: 'wysiwyg',
    placeholder: '请输入 Markdown 内容...',
    upload: {
      url: '/api/contents/upload',
      linkToImgUrl: '/api/contents/upload',
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
})

watch(editorMode, (newMode) => {
  if (newMode === 'web') {
    destroyVditor()
    nextTick(() => {
      const quill = quillEditorRef.value?.getQuill()
      if (quill) {
        quill.root.innerHTML = ''
      }
    })
  } else {
    markdownContent.value = ''
    nextTick(() => initVditor())
  }
})

function handleModeChange(newMode) {
  if (newMode === editorMode.value) {
    return
  }
  
  const currentContent = editorMode.value === 'web' 
    ? quillEditorRef.value?.getQuill()?.root.innerHTML 
    : markdownContent.value
  
  const hasCurrentContent = editorMode.value === 'web'
    ? (currentContent && currentContent !== '<p><br></p>')
    : (currentContent && currentContent.trim() !== '')
  
  if (hasCurrentContent) {
    hasContent.value = true
    pendingMode.value = newMode
    showConfirmModal.value = true
  } else {
    editorMode.value = newMode
  }
}

function confirmModeChange() {
  showConfirmModal.value = false
  hasContent.value = false
  if (pendingMode.value) {
    editorMode.value = pendingMode.value
    pendingMode.value = null
  }
}

function cancelModeChange() {
  showConfirmModal.value = false
  pendingMode.value = null
}

watch(markdownContent, (val) => {
  const fullContent = val ? `[md]${val}` : ''
  emit('update:modelValue', fullContent)
})

function onEditorReady() {
  initQuillHandlers()
}

function initQuillHandlers() {
  const quill = quillEditorRef.value?.getQuill()
  if (!quill) return
  
  const toolbar = quill.getModule('toolbar')
  if (toolbar) {
    toolbar.addHandler('image', () => handleImageUpload())
    toolbar.addHandler('video', () => handleVideoUpload())
  }
  
  quill.on('text-change', () => {
    const html = quill.root.innerHTML
    if (editorMode.value === 'web') {
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
      const quill = quillEditorRef.value?.getQuill()
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
    const quill = quillEditorRef.value?.getQuill()
    if (quill) {
      quill.root.innerHTML = content || ''
    }
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
  getQuill: () => quillEditorRef.value?.getQuill(),
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
