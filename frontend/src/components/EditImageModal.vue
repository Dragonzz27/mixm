<template>
  <n-modal
    v-model:show="showModal"
    preset="card"
    title="仓库设置"
    style="width: 480px;"
    :mask-closable="false"
    @close="handleClose"
  >
    <n-form ref="formRef" :model="form">
      <n-form-item label="访问权限" path="is_public">
        <n-radio-group v-model:value="form.is_public">
          <n-space>
            <n-radio :value="false">私有</n-radio>
            <n-radio :value="true">公开</n-radio>
          </n-space>
        </n-radio-group>
      </n-form-item>
    </n-form>
    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 12px;">
        <n-button @click="handleClose">取消</n-button>
        <n-button type="primary" @click="handleSubmit" :loading="submitting">保存</n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { imageApi } from '@/api'

const props = defineProps({
  show: Boolean,
  repository: Object
})

const emit = defineEmits(['update:show', 'success'])

const message = useMessage()

const formRef = ref(null)
const submitting = ref(false)
const form = ref({
  is_public: false
})

const showModal = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})

watch(() => props.show, (val) => {
  if (val && props.repository) {
    form.value = {
      is_public: props.repository.is_public || false
    }
  }
})

function handleClose() {
  emit('update:show', false)
}

async function handleSubmit() {
  try {
    submitting.value = true
    await imageApi.update(props.repository.id, form.value)
    message.success('保存成功')
    emit('success')
    handleClose()
  } catch (err) {
    message.error('保存失败')
  } finally {
    submitting.value = false
  }
}
</script>
