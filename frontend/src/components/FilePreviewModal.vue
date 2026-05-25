<template>
  <n-modal
    v-model:show="showModal"
    preset="card"
    :title="fileData?.file_name || '文件预览'"
    style="width: 90vw; max-width: 1400px; height: 85vh;"
    :closable="true"
    @close="handleClose"
  >
    <FilePreviewPanel :file-data="fileData" class="modal-preview-panel" />
  </n-modal>
</template>

<script setup>
import { ref, watch } from 'vue'
import { NModal } from 'naive-ui'
import FilePreviewPanel from './FilePreviewPanel.vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  fileData: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:show', 'close'])

const showModal = ref(false)
const fileData = ref(null)

watch(() => props.show, (newVal) => {
  showModal.value = newVal
  if (!newVal) fileData.value = null
})

watch(() => props.fileData, (newVal) => {
  fileData.value = newVal
}, { immediate: true })

function handleClose() {
  emit('update:show', false)
  emit('close')
  fileData.value = null
}
</script>

<style scoped>
.modal-preview-panel {
  height: 75vh;
}
</style>
