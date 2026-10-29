<template>
  <n-modal v-model:show="showModal" preset="dialog" title="启动实例" style="width: 600px;">
    <n-form :model="formData" label-placement="left" label-width="100px">
      <n-form-item v-if="isAdmin" label="所属用户" path="user_id">
        <n-select
          v-model:value="formData.user_id"
          :options="userOptions"
          placeholder="选择实例所属用户"
          :disabled="props.readonlyConfig"
        />
      </n-form-item>
      <n-form-item label="镜像仓库" path="repo_id">
        <n-select 
          v-model:value="formData.repo_id" 
          :options="repoOptions" 
          placeholder="选择镜像仓库"
          :disabled="props.readonlyConfig"
          @update:value="handleRepoChange"
        />
      </n-form-item>
      <n-form-item label="镜像标签" path="repo_tag">
        <n-select 
          v-model:value="formData.repo_tag" 
          :options="tagOptions" 
          placeholder="选择镜像标签"
          :loading="tagsLoading"
          :disabled="props.readonlyConfig || !formData.repo_id"
        />
      </n-form-item>
      <n-form-item label="CPU" path="cpu_limit">
        <n-select 
          v-model:value="formData.cpu_limit" 
          :options="cpuOptions" 
          placeholder="选择CPU核数"
          :disabled="props.readonlyConfig"
        />
      </n-form-item>
      <n-form-item label="内存" path="memory_limit">
        <n-select 
          v-model:value="formData.memory_limit" 
          :options="memoryOptions" 
          placeholder="选择内存大小"
          :disabled="props.readonlyConfig"
        />
      </n-form-item>
      <n-form-item label="临时存储" path="storage_limit">
        <n-select 
          v-model:value="formData.storage_limit" 
          :options="storageOptions" 
          placeholder="选择存储大小"
          :disabled="props.readonlyConfig"
        />
      </n-form-item>
      <n-form-item label="GPU数量" path="gpu_limit">
        <n-select
          v-model:value="formData.gpu_limit"
          :options="gpuOptions"
          placeholder="选择GPU数量"
          :disabled="props.readonlyConfig"
        />
      </n-form-item>
      <n-form-item label="单卡显存" path="gpu_memory_limit">
        <n-select
          v-model:value="formData.gpu_memory_limit"
          :options="gpuMemoryOptions"
          placeholder="选择每张GPU显存"
          :disabled="props.readonlyConfig || formData.gpu_limit === 0"
        />
      </n-form-item>
      <n-form-item label="实例名称" path="name">
        <n-input v-model:value="formData.name" placeholder="请输入实例名称" :disabled="props.readonlyConfig" />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="showModal = false">取消</n-button>
      <n-button type="primary" :loading="loading" @click="handleSubmit">启动</n-button>
    </template>
  </n-modal>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { NModal, NForm, NFormItem, NInput, NSelect, NButton, useMessage } from 'naive-ui'
import { instanceApi, authApi, imageApi } from '@/api'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  readonlyConfig: {
    type: Boolean,
    default: false
  },
  enableUserSelect: {
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

const isAdmin = computed(() => props.enableUserSelect)

const users = ref([])
const repos = ref([])
const tags = ref([])
const tagsLoading = ref(false)

const cpuOptions = [
  { label: '0.5核', value: 0.5 },
  { label: '1核', value: 1 },
  { label: '2核', value: 2 },
  { label: '4核', value: 4 },
  { label: '8核', value: 8 },
  { label: '16核', value: 16 },
]

const memoryOptions = [
  { label: '512Mi', value: '512Mi' },
  { label: '1Gi', value: '1Gi' },
  { label: '2Gi', value: '2Gi' },
  { label: '4Gi', value: '4Gi' },
  { label: '8Gi', value: '8Gi' },
  { label: '16Gi', value: '16Gi' },
  { label: '32Gi', value: '32Gi' },
]

const storageOptions = [
  { label: '1Gi', value: 1 },
  { label: '5Gi', value: 5 },
  { label: '10Gi', value: 10 },
  { label: '20Gi', value: 20 },
  { label: '40Gi', value: 40 },
  { label: '200Gi', value: 200 },
]

const gpuOptions = [
  { label: '0张', value: 0 },
  { label: '1张', value: 1 },
  { label: '2张', value: 2 },
  { label: '4张', value: 4 },
  { label: '8张', value: 8 },
]

const gpuMemoryOptions = [
  { label: '2Gi', value: 2048 },
  { label: '4Gi', value: 4096 },
  { label: '8Gi', value: 8192 },
  { label: '12Gi', value: 12288 },
  { label: '16Gi', value: 16384 },
  { label: '24Gi', value: 24576 },
  { label: '32Gi', value: 32600 },
]

const formData = ref({
  name: '',
  user_id: null,
  repo_id: null,
  repo_tag: null,
  cpu_limit: 1,
  memory_limit: '2Gi',
  storage_limit: 1,
  gpu_limit: 1,
  gpu_memory_limit: 2048
})

const userOptions = computed(() => {
  return users.value.map(u => ({ label: u.username, value: u.id }))
})

const repoOptions = computed(() => {
  return repos.value.map(r => ({ label: r.name, value: r.id }))
})

const tagOptions = computed(() => {
  return tags.value.map(t => ({ label: t.name, value: t.name }))
})

function parseConfig(configStr) {
  if (!configStr) return {}
  if (typeof configStr === 'object') return configStr
  try {
    return JSON.parse(configStr)
  } catch (e) {
    return {}
  }
}

function pickNumber(value, fallback) {
  if (value === null || value === undefined || value === '') {
    return fallback
  }
  const parsed = Number(value)
  return Number.isNaN(parsed) ? fallback : parsed
}

function initFormData() {
  const templateConfig = props.template ? parseConfig(props.template.config) : null
  
  formData.value = {
    name: templateConfig?.name ?? `MixM-${Date.now()}`,
    user_id: null,
    repo_id: templateConfig?.repo_id ?? null,
    repo_tag: templateConfig?.repo_tag ?? null,
    cpu_limit: pickNumber(templateConfig?.cpu_limit, 1),
    memory_limit: templateConfig?.memory_limit_str ?? '2Gi',
    storage_limit: pickNumber(templateConfig?.storage, 1),
    gpu_limit: pickNumber(templateConfig?.gpu_limit, 0),
    gpu_memory_limit: pickNumber(templateConfig?.gpu_memory_limit, 0)
  }

  if (formData.value.gpu_limit === 0) {
    formData.value.gpu_memory_limit = 0
  } else if (formData.value.gpu_memory_limit <= 0) {
    formData.value.gpu_memory_limit = 2048
  }

  tags.value = []
  
  if (formData.value.repo_id) {
    loadTags(formData.value.repo_id)
  }
}

async function loadRepos() {
  try {
    const [myRes, publicRes] = await Promise.all([
      imageApi.getList({ type: 'my' }),
      imageApi.getList({ type: 'public' })
    ])
    const myRepos = myRes.items || []
    const publicRepos = publicRes.items || []
    const mergedRepos = [...myRepos, ...publicRepos]
    const repoMap = new Map()

    mergedRepos.forEach((repo) => {
      const repoKey = repo.id ?? repo.name
      if (repoKey === undefined || repoKey === null) {
        return
      }
      if (!repoMap.has(repoKey)) {
        repoMap.set(repoKey, repo)
      }
    })

    repos.value = Array.from(repoMap.values())
  } catch (err) {
    console.error(err)
  }
}

async function loadTags(repoId) {
  if (!repoId) {
    tags.value = []
    return
  }
  tagsLoading.value = true
  tags.value = []
  try {
    const currentTag = formData.value.repo_tag
    const res = await imageApi.getAllTags(repoId)
    if (formData.value.repo_id !== repoId) return
    tags.value = res.items || []

    if (currentTag && tags.value.some(t => t.name === currentTag)) {
      formData.value.repo_tag = currentTag
    } else if (!currentTag && tags.value.length > 0) {
      formData.value.repo_tag = tags.value[0].name
    } else if (currentTag && tags.value.length > 0) {
      formData.value.repo_tag = tags.value[0].name
    } else {
      formData.value.repo_tag = null
    }
  } catch (err) {
    console.error(err)
    message.error('加载镜像标签失败')
  } finally {
    tagsLoading.value = false
  }
}

function handleRepoChange(repoId) {
  if (props.readonlyConfig) return
  formData.value.repo_tag = null
  loadTags(repoId)
}

watch(() => formData.value.gpu_limit, (val) => {
  if (val === 0) {
    formData.value.gpu_memory_limit = 0
  }
})

async function loadUsers() {
  if (!isAdmin.value) {
    users.value = []
    return
  }

  try {
    const res = await authApi.getUsers({})
    users.value = res.items || []
  } catch (err) {
    console.error(err)
  }
}

async function handleSubmit() {
  if (!formData.value.repo_id) {
    message.warning('请选择镜像仓库')
    return
  }
  if (!formData.value.repo_tag) {
    message.warning('请选择镜像标签')
    return
  }
  if (!formData.value.name) {
    message.warning('请输入实例名称')
    return
  }
  if (formData.value.gpu_limit > 0 && formData.value.gpu_memory_limit <= 0) {
    message.warning('请选择单卡显存')
    return
  }
  if (isAdmin.value && !formData.value.user_id) {
    message.warning('请选择所属用户')
    return
  }

  loading.value = true
  try {
    const payload = {
      name: formData.value.name,
      repo_id: formData.value.repo_id,
      repo_tag: formData.value.repo_tag,
      cpu_limit: formData.value.cpu_limit,
      memory_limit: formData.value.memory_limit,
      storage_limit: formData.value.storage_limit,
      gpu_limit: formData.value.gpu_limit,
      gpu_memory_limit: formData.value.gpu_memory_limit
    }

    if (isAdmin.value && formData.value.user_id) {
      payload.user_id = formData.value.user_id
    }

    const res = await instanceApi.startQueued(payload)
    message.success(res.message || '实例已加入创建队列，请稍后查看进度')
    showModal.value = false
    emit('success')
  } catch (err) {
    console.error(err)
    const errorMsg = err.response?.data?.message || err.message || '实例启动失败，请检查配置或联系管理员'
    message.error(errorMsg)
  } finally {
    loading.value = false
  }
}

watch(showModal, (val) => {
  if (val) {
    initFormData()
    loadUsers()
    loadRepos()
  }
})
</script>
