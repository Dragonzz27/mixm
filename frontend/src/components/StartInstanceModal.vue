<template>
  <n-modal v-model:show="showModal" preset="dialog" title="启动实例" style="width: 500px;">
    <n-form :model="formData">
      <n-form-item v-if="template" label="模板">
        <n-input :value="template.name" disabled />
      </n-form-item>
      <n-form-item v-if="!template" label="模板" path="template_id">
        <n-select v-model:value="formData.template_id" :options="templateOptions" placeholder="选择模板" />
      </n-form-item>
      <n-form-item label="实例名称" path="name">
        <n-input v-model:value="formData.name" placeholder="请输入实例名称" />
      </n-form-item>
      <n-form-item label="命名空间" path="namespace">
        <n-input v-model:value="formData.namespace" placeholder="例如：default, production" />
      </n-form-item>
      <n-form-item label="绑定用户" path="user_id">
        <n-select v-model:value="formData.user_id" :options="userOptions" placeholder="选择用户" />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="showModal = false">取消</n-button>
      <n-button type="primary" :loading="loading" @click="handleSubmit">启动</n-button>
    </template>
  </n-modal>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { NModal, NForm, NFormItem, NInput, NSelect, NButton, useMessage } from 'naive-ui'
import { instanceApi, authApi, templateApi } from '@/api'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  template: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:show', 'success'])

const message = useMessage()
const loading = ref(false)
const showModal = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})

const users = ref([])
const templates = ref([])
const formData = ref({
  template_id: null,
  template_name: '',
  name: '',
  namespace: 'default',
  user_id: null
})

const userOptions = computed(() => {
  return users.value.map(u => ({ label: u.username, value: u.id }))
})

const templateOptions = computed(() => {
  return templates.value.map(t => ({ label: t.name, value: t.id }))
})

async function loadUsers() {
  try {
    const res = await authApi.getUsers({})
    users.value = res.users || []
  } catch (err) {
    console.error(err)
  }
}

async function loadTemplates() {
  try {
    const res = await templateApi.getList({})
    templates.value = res.templates || []
  } catch (err) {
    console.error(err)
  }
}

function initFormData() {
  if (props.template) {
    formData.value = {
      template_id: props.template.id,
      template_name: props.template.name,
      name: `instance-${Date.now()}`,
      namespace: 'default',
      user_id: null
    }
  } else {
    formData.value = {
      template_id: null,
      template_name: '',
      name: `instance-${Date.now()}`,
      namespace: 'default',
      user_id: null
    }
  }
}

async function handleSubmit() {
  if (!formData.value.template_id) {
    message.warning('请选择模板')
    return
  }
  if (!formData.value.user_id) {
    message.warning('请选择用户')
    return
  }
  if (!formData.value.name) {
    message.warning('请输入实例名称')
    return
  }

  const selectedTemplate = templates.value.find(t => t.id === formData.value.template_id)

  loading.value = true
  try {
    const res = await instanceApi.startByTemplate({
      name: formData.value.name,
      template_id: formData.value.template_id
    })
    message.success(res.message || '实例已加入创建队列，请稍后查看进度')
    showModal.value = false
    emit('success')
  } catch (err) {
    console.error(err)
    const errorMsg = err.response?.data?.detail || err.message || '实例启动失败，请检查配置或联系管理员'
    message.error(errorMsg)
  } finally {
    loading.value = false
  }
}

watch(showModal, (val) => {
  if (val) {
    initFormData()
    loadUsers()
    loadTemplates()
  }
})

watch(() => props.template, () => {
  initFormData()
})
</script>
