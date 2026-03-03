<template>
  <n-modal
    v-model:show="showModal"
    preset="card"
    title="创建镜像仓库"
    style="width: 480px;"
    :mask-closable="false"
    @close="handleClose"
  >
    <n-form ref="formRef" :model="form" :rules="rules">
      <n-form-item label="仓库名称" path="name">
        <n-input v-model:value="form.name" placeholder="请输入仓库名称" />
      </n-form-item>
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
        <n-button type="primary" @click="handleSubmit" :loading="submitting">创建</n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { imageApi } from '@/api'

const props = defineProps({
  show: Boolean
})

const emit = defineEmits(['update:show', 'success'])

const message = useMessage()

const formRef = ref(null)
const submitting = ref(false)
const form = ref({
  name: '',
  is_public: false
})

const rules = {
  name: [
    { required: true, message: '请输入仓库名称', trigger: 'blur' },
    { min: 1, max: 100, message: '长度在 1 到 100 个字符', trigger: 'blur' }
  ]
}

const showModal = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})

watch(() => props.show, (val) => {
  if (!val) {
    form.value = { name: '', is_public: false }
  }
})

function handleClose() {
  emit('update:show', false)
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
    submitting.value = true
    await imageApi.create(form.value)
    message.success('创建成功')
    emit('success')
    handleClose()
  } catch (err) {
    if (err.response?.data?.detail) {
      message.error(err.response.data.detail)
    } else {
      message.error('创建失败')
    }
  } finally {
    submitting.value = false
  }
}
</script>
