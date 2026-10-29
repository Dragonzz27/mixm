<template>
  <div class="templates">
    <n-card title="模板管理">
      <template #header-extra>
        <n-space>
          <n-input v-model:value="searchKeyword" placeholder="搜索模板名称" clearable style="width: 200px;" @keyup.enter="handleSearch" />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleAddClick">添加模板</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="templates"
        :pagination="pagination"
        remote
        :bordered="false"
      />
    </n-card>
    <n-modal v-model:show="showModal" preset="dialog" :title="isEdit ? '编辑模板' : '添加模板'" style="width: 800px; max-height: 80vh;" :show-override="true">
      <n-scrollbar style="max-height: 60vh;">
        <n-form :model="formData" :rules="rules" ref="formRef" label-placement="left" label-width="120px">
          <n-divider>基本信息</n-divider>
          <n-form-item label="模板名称" path="name">
            <n-input v-model:value="formData.name" placeholder="请输入模板名称" />
          </n-form-item>
          <n-form-item label="描述" path="description">
            <n-input v-model:value="formData.description" type="textarea" :rows="2" placeholder="请输入描述" />
          </n-form-item>

          <n-divider>容器配置</n-divider>
          <n-form-item label="镜像仓库" path="config.repo_id">
            <n-select 
              v-model:value="formData.config.repo_id" 
              :options="repoOptions" 
              placeholder="选择镜像仓库"
              @update:value="handleRepoChange"
            />
          </n-form-item>
          <n-form-item label="镜像标签" path="config.repo_tag">
            <n-select 
              v-model:value="formData.config.repo_tag" 
              :options="tagOptions" 
              placeholder="选择镜像标签"
              :loading="tagsLoading"
              :disabled="!formData.config.repo_id"
            />
          </n-form-item>

          <n-divider>资源限制</n-divider>
          <n-row :gutter="24">
            <n-col :span="12">
              <n-form-item label="CPU" path="config.cpu_limit">
                <n-select 
                  v-model:value="formData.config.cpu_limit" 
                  :options="cpuOptions" 
                  placeholder="选择CPU核数"
                />
              </n-form-item>
            </n-col>
            <n-col :span="12">
              <n-form-item label="内存" path="config.memory_limit_str">
                <n-select 
                  v-model:value="formData.config.memory_limit_str" 
                  :options="memoryOptions" 
                  placeholder="选择内存大小"
                />
              </n-form-item>
            </n-col>
          </n-row>
          <n-row :gutter="24">
            <n-col :span="12">
              <n-form-item label="临时存储" path="config.storage">
                <n-select 
                  v-model:value="formData.config.storage" 
                  :options="storageOptions" 
                  placeholder="选择存储大小"
                />
              </n-form-item>
            </n-col>
            <n-col :span="12">
              <n-form-item label="GPU数量" path="config.gpu_limit">
                <n-select
                  v-model:value="formData.config.gpu_limit"
                  :options="gpuOptions"
                  placeholder="选择GPU数量"
                />
              </n-form-item>
            </n-col>
          </n-row>
          <n-form-item label="单卡显存" path="config.gpu_memory_limit">
            <n-select
              v-model:value="formData.config.gpu_memory_limit"
              :options="gpuMemoryOptions"
              placeholder="选择每张GPU显存"
              :disabled="formData.config.gpu_limit === 0"
            />
          </n-form-item>
        </n-form>
      </n-scrollbar>
      <template #action>
        <n-button @click="showModal = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleSubmit">确定</n-button>
      </template>
    </n-modal>
    <StartInstanceByTemplateModal
      v-model:show="showStartModal"
      :template="selectedTemplate"
      :enable-user-select="true"
      @success="onInstanceStarted"
    />
    <n-modal v-model:show="showBatchStartModal" preset="dialog" title="批量启动实例" style="width: 640px;">
      <n-space vertical :size="16">
        <n-alert type="info" v-if="batchTemplate">
          模板：{{ batchTemplate.name }}，请选择需要启动实例的用户。
        </n-alert>
        <n-input
          v-model:value="batchUserKeyword"
          placeholder="按用户名/邮箱/手机号筛选用户"
          clearable
        />
        <n-select
          v-model:value="batchRoleId"
          :options="batchRoleOptions"
          placeholder="按角色筛选用户"
          clearable
        />
        <n-space justify="space-between" align="center">
          <n-text depth="3">已选用户：{{ batchUserIds.length }}</n-text>
          <n-space>
            <n-button size="small" @click="selectAllFilteredUsers">全选筛选结果</n-button>
            <n-button size="small" @click="clearBatchUserSelection">清空选择</n-button>
          </n-space>
        </n-space>
        <n-space :size="8" align="center">
          <n-tag type="info">筛选结果：{{ filteredBatchUsers.length }}</n-tag>
          <n-tag type="success">已勾选：{{ batchUserIds.length }}</n-tag>
          <n-tag v-if="batchRoleId" type="warning">
            角色筛选：{{ batchRoleOptions.find(item => item.value === batchRoleId)?.label || '-' }}
          </n-tag>
          <n-tag v-if="batchUserKeyword" type="default">关键词：{{ batchUserKeyword }}</n-tag>
        </n-space>
        <div class="batch-user-table-wrap">
          <n-data-table
            size="small"
            :columns="batchUserTableColumns"
            :data="filteredBatchUsers"
            :row-key="batchUserRowKey"
            :max-height="420"
          />
        </div>
      </n-space>
      <template #action>
        <n-button @click="showBatchStartModal = false">取消</n-button>
        <n-button type="primary" :loading="batchLoading" @click="handleBatchStartSubmit">批量启动</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h, watch } from 'vue'
import { NCard, NButton, NDataTable, NSpace, NModal, NForm, NFormItem, NInput, NInputNumber, NDivider, NRow, NCol, NScrollbar, NSelect, NAlert, NText, NCheckbox, NTag, useMessage } from 'naive-ui'
import { templateApi, imageApi, authApi, instanceApi, roleApi } from '@/api'
import StartInstanceByTemplateModal from '@/components/StartInstanceByTemplateModal.vue'

const message = useMessage()
const templates = ref([])
const showModal = ref(false)
const showStartModal = ref(false)
const showBatchStartModal = ref(false)
const selectedTemplate = ref(null)
const batchTemplate = ref(null)
const isEdit = ref(false)
const loading = ref(false)
const batchLoading = ref(false)
const searchKeyword = ref('')
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    pagination.value.page = page
    loadTemplates()
  },
  onUpdatePageSize: (pageSize) => {
    pagination.value.pageSize = pageSize
    pagination.value.page = 1
    loadTemplates()
  }
})
const formRef = ref(null)
const users = ref([])
const roles = ref([])
const batchUserIds = ref([])
const batchUserKeyword = ref('')
const batchRoleId = ref(null)

const repos = ref([])
const tags = ref([])
const tagsLoading = ref(false)

// 默认配置
const defaultConfig = {
  repo_id: null,
  repo_tag: null,
  namespace: 'default',
  replicas: 1,
  cpu_limit: 1,
  memory_limit_str: '2Gi',
  storage: 1,
  gpu_limit: 1,
  gpu_memory_limit: 2048
}

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

const repoOptions = computed(() => {
  return repos.value.map(r => ({ label: r.name, value: r.id }))
})

const tagOptions = computed(() => {
  return tags.value.map(t => ({ label: t.name, value: t.name }))
})

const formData = ref({
  name: '',
  description: '',
  config: { ...defaultConfig }
})

const rules = {
  name: [{ required: true, message: '请输入模板名称' }],
  'config.repo_id': [{ required: true, message: '请选择镜像仓库' }],
  'config.repo_tag': [{ required: true, message: '请选择镜像标签' }],
  'config.namespace': [{ required: true, message: '请输入命名空间' }],
  'config.container_port': [{ required: true, message: '请输入容器端口', type: 'number' }],
  'config.cpu_limit': [{ required: true, message: '请选择CPU' }],
  'config.memory_limit_str': [{ required: true, message: '请选择内存' }],
  'config.storage': [{ required: true, message: '请选择存储大小' }],
  'config.gpu_limit': [{ required: true, message: '请选择GPU数量' }],
  'config.gpu_memory_limit': [{ required: true, message: '请选择单卡显存' }]
}

function getRowNumber(index) {
  const page = pagination.value.page || 1
  const pageSize = pagination.value.pageSize || 10
  return (page - 1) * pageSize + index + 1
}

const columns = [
  {
    title: '序号',
    key: 'index',
    width: 80,
    render: (_, index) => getRowNumber(index)
  },
  { title: '模板名称', key: 'name', width: 150 },
  { title: '描述', key: 'description', ellipsis: { tooltip: true } },
  { title: '镜像', key: 'image', ellipsis: { tooltip: true }, width: 200 },
  { title: '副本数', key: 'replicas', width: 80 },
  { title: '创建时间', key: 'created_at', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 320,
    render: (row) => h(NSpace, {}, () => [
      h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => '编辑'),
      h(NButton, { size: 'small', type: 'success', onClick: () => handleStartClick(row) }, () => '启动实例'),
      h(NButton, { size: 'small', type: 'primary', onClick: () => handleBatchStartClick(row) }, () => '批量启动'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
]

const filteredBatchUsers = computed(() => {
  const keyword = (batchUserKeyword.value || '').trim().toLowerCase()

  return users.value.filter((u) => {
    const userRoleId = String(u.role_id || u.role?.id || '')
    const selectedRoleId = String(batchRoleId.value || '')
    const roleMatched = !batchRoleId.value || userRoleId === selectedRoleId
    if (!roleMatched) return false

    if (!keyword) return true

    const username = (u.username || '').toLowerCase()
    const email = (u.email || '').toLowerCase()
    const roleName = (u.role?.name || '').toLowerCase()
    return username.includes(keyword) || email.includes(keyword) || roleName.includes(keyword)
  })
})

const batchRoleOptions = computed(() => {
  return roles.value.map((r) => ({
    label: r.name,
    value: r.id,
  }))
})

const batchUserColumns = [
  {
    title: '序号',
    key: 'index',
    width: 80,
    render: (_, index) => index + 1
  },
  { title: '用户名', key: 'username', width: 140 },
  { title: '邮箱', key: 'email', width: 220, render: (row) => row.email || '-' },
  { title: '角色', key: 'role', width: 140, render: (row) => row.role?.name || '-' },
  {
    title: '状态',
    key: 'is_active',
    width: 100,
    render: (row) => (row.is_active ? '激活' : '禁用')
  }
]

const batchUserTableColumns = computed(() => {
  const selectionColumn = {
    title: '',
    key: 'selection',
    width: 56,
    render: (row) => h(NCheckbox, {
      checked: batchUserIds.value.includes(row.id),
      'onUpdate:checked': (checked) => toggleBatchUserChecked(row.id, checked),
    })
  }
  return [selectionColumn, ...batchUserColumns]
})

async function loadTemplates() {
  try {
    const page = pagination.value.page || 1
    const pageSize = pagination.value.pageSize || 10
    const params = {
      offset: (page - 1) * pageSize,
      limit: pageSize
    }
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }
    const res = await templateApi.getList(params)
    // 解析配置以显示关键信息
    templates.value = (res.items || []).map(t => {
      let config = {}
      try {
        config = t.config ? JSON.parse(t.config) : {}
      } catch (e) {
        config = {}
      }
      console.log('Parsed config for template', t.id, config)
      return {
        ...t,
        image: buildImageString(config) || '-',
        replicas: config.replicas || 1
      }
    })
    pagination.value.itemCount = res.total || 0

    const maxPage = Math.max(1, Math.ceil((pagination.value.itemCount || 0) / pageSize))
    if (pagination.value.page > maxPage) {
      pagination.value.page = maxPage
      return loadTemplates()
    }
  } catch (err) {
    console.error(err)
  }
}

function handleSearch() {
  pagination.value.page = 1
  loadTemplates()
}

function parseConfig(configStr) {
  if (!configStr) return { ...defaultConfig }
  try {
    const parsed = JSON.parse(configStr)
    return {
      ...defaultConfig,
      ...parsed
    }
  } catch (e) {
    return { ...defaultConfig }
  }
}

function getRepoNameById(repoId) {
  if (repoId === null || repoId === undefined) return ''
  const repo = repos.value.find(r => r.id === repoId)
  return repo?.name || ''
}

function buildImageString(config) {
  if (!config) return ''
  if (typeof config.image === 'string' && config.image.trim()) {
    return config.image.trim()
  }

  const repoTag = typeof config.repo_tag === 'string' ? config.repo_tag.trim() : ''
  if (!repoTag) return ''

  if (repoTag.includes(':')) {
    return repoTag
  }

  const repoName = config.repo_name || getRepoNameById(config.repo_id)
  return repoName ? `${repoName}:${repoTag}` : ''
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

async function loadUsers() {
  try {
    const res = await authApi.getUsers({ offset: 0, limit: 1000 })
    users.value = res.items || []
  } catch (err) {
    console.error(err)
    users.value = []
  }
}

async function loadRoles() {
  try {
    const res = await roleApi.getList({ offset: 0, limit: 200 })
    roles.value = res.items || []
  } catch (err) {
    console.error(err)
    roles.value = []
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
    const res = await imageApi.getAllTags(repoId)
    if (formData.value.config.repo_id !== repoId) return
    tags.value = res.items || []
  } catch (err) {
    console.error(err)
    message.error('加载镜像标签失败')
  } finally {
    tagsLoading.value = false
  }
}

function handleRepoChange(repoId) {
  formData.value.config.repo_tag = null
  loadTags(repoId)
}

async function openModal() {
  await loadRepos()
  if (formData.value.config.repo_id) {
    await loadTags(formData.value.config.repo_id)
  }
}

function handleAddClick() {
  isEdit.value = false
  formData.value = {
    name: '',
    description: '',
    config: { ...defaultConfig }
  }
  openModal()
  showModal.value = true
}

function handleEdit(row) {
  isEdit.value = true
  formData.value = {
    ...row,
    config: parseConfig(row.config)
  }
  openModal()
  showModal.value = true
}

function handleStartClick(row) {
  selectedTemplate.value = row
  showStartModal.value = true
}

function handleBatchStartClick(row) {
  batchTemplate.value = row
  batchUserIds.value = []
  batchUserKeyword.value = ''
  batchRoleId.value = null
  showBatchStartModal.value = true
}

const batchUserRowKey = (row) => row.id

function toggleBatchUserChecked(userId, checked) {
  if (checked) {
    if (!batchUserIds.value.includes(userId)) {
      batchUserIds.value = [...batchUserIds.value, userId]
    }
    return
  }
  batchUserIds.value = batchUserIds.value.filter((id) => id !== userId)
}

function selectAllFilteredUsers() {
  batchUserIds.value = filteredBatchUsers.value.map((u) => u.id)
}

function clearBatchUserSelection() {
  batchUserIds.value = []
}

async function handleBatchStartSubmit() {
  if (!batchTemplate.value?.id) {
    message.warning('模板信息无效')
    return
  }
  if (!batchUserIds.value.length) {
    message.warning('请至少选择一个用户')
    return
  }

  batchLoading.value = true
  try {
    const res = await instanceApi.startBatchByTemplate({
      template_id: batchTemplate.value.id,
      user_ids: batchUserIds.value,
    })
    message.success(res.message || `已提交 ${res.queued_count || 0} 个创建任务`)
    showBatchStartModal.value = false
    window.dispatchEvent(new CustomEvent('queue-updated'))
  } catch (err) {
    console.error(err)
    const errorMsg = err.response?.data?.message || err.message || '批量启动失败'
    message.error(errorMsg)
  } finally {
    batchLoading.value = false
  }
}

watch(() => formData.value.config.gpu_limit, (val) => {
  if (val === 0) {
    formData.value.config.gpu_memory_limit = 0
  }
})

watch([batchUserKeyword, batchRoleId], () => {
  batchUserIds.value = []
})

async function handleSubmit() {
  const configToSubmit = {
    ...formData.value.config,
    repo_name: getRepoNameById(formData.value.config.repo_id)
  }
  configToSubmit.image = buildImageString(configToSubmit)

  // 构建提交数据
  const submitData = {
    name: formData.value.name,
    description: formData.value.description,
    config: JSON.stringify(configToSubmit)
  }

  if (isEdit.value) {
    submitData.id = formData.value.id
  }

  loading.value = true
  try {
    if (isEdit.value) {
      await templateApi.update(formData.value.id, submitData)
      message.success('更新成功')
    } else {
      await templateApi.create(submitData)
      message.success('添加成功')
    }
    showModal.value = false
    loadTemplates()
  } catch (err) {
    console.error(err)
    message.error(isEdit.value ? '更新失败' : '添加失败')
  } finally {
    loading.value = false
  }
}

function onInstanceStarted() {
  message.success('实例启动成功')
}

async function handleDelete(id) {
  try {
    await templateApi.delete(id)
    message.success('删除成功')
    loadTemplates()
  } catch (err) {
    console.error(err)
    message.error('删除失败')
  }
}

onMounted(() => {
  loadTemplates()
  loadUsers()
  loadRoles()
})
</script>

<style scoped>
.templates {
  padding: 0;
}

:deep(.n-modal-body) {
  max-height: 60vh;
}

.batch-user-table-wrap {
  height: 420px;
}
</style>
