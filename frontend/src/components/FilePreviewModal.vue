<template>
  <n-modal
    v-model:show="showModal"
    preset="card"
    :title="fileData?.file_name || '文件预览'"
    style="width: 90vw; max-width: 1400px; height: 85vh"
    :closable="true"
    @close="handleClose"
  >
    <FilePreviewPanel :file-data="fileData" class="modal-preview-panel" />
  </n-modal>
</template>

<script setup>
import { computed } from 'vue'
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

// 同步遮罩、Esc 和关闭按钮的状态，使文件预览关闭后可以再次打开。
const showModal = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})
const fileData = computed(() => (props.show ? props.fileData : null))

function handleClose() {
  emit('update:show', false)
  emit('close')
}
</script>

<style scoped>
.modal-preview-panel {
  height: 75vh;
}
</style>
