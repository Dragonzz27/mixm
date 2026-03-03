<template>
  <div class="templates">
    <n-card title="模板管理">
      <template #header-extra>
        <n-space>
          <n-input v-model:value="searchKeyword" placeholder="搜索模板名称" clearable style="width: 200px;" @keyup.enter="loadTemplates" />
          <n-button type="primary" @click="loadTemplates">搜索</n-button>
          <n-button @click="handleAddClick">添加模板</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="templates"
        :pagination="pagination"
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
          <n-form-item label="镜像" path="config.image">
            <n-input v-model:value="formData.config.image" placeholder="例如：codercom/code-server:latest" />
          </n-form-item>
          <n-form-item label="命名空间" path="config.namespace">
            <n-input v-model:value="formData.config.namespace" placeholder="例如：default" />
          </n-form-item>
          <n-form-item label="副本数" path="config.replicas">
            <n-input-number v-model:value="formData.config.replicas" :min="1" :max="10" />
          </n-form-item>

          <n-row :gutter="24">
            <n-col :span="24">
              <n-form-item label="容器端口" path="config.container_port">
                <n-input-number v-model:value="formData.config.container_port" :min="1" :max="65535" />
              </n-form-item>
            </n-col>
          </n-row>

          <n-divider>资源限制</n-divider>
          <n-row :gutter="24">
            <n-col :span="12">
              <n-form-item label="CPU限制" path="config.cpu_limit">
                <n-input v-model:value="formData.config.cpu_limit" placeholder="例如：1 或 500m" />
              </n-form-item>
            </n-col>
            <n-col :span="12">
              <n-form-item label="内存限制" path="config.memory_limit_str">
                <n-input v-model:value="formData.config.memory_limit_str" placeholder="例如：1Gi 或 512Mi" />
              </n-form-item>
            </n-col>
          </n-row>
          <n-row :gutter="24">
            <n-col :span="12">
              <n-form-item label="CPU请求" path="config.cpu_request">
                <n-input v-model:value="formData.config.cpu_request" placeholder="例如：0.5 或 250m" />
              </n-form-item>
            </n-col>
            <n-col :span="12">
              <n-form-item label="内存请求" path="config.memory_request">
                <n-input v-model:value="formData.config.memory_request" placeholder="例如：512Mi 或 256Mi" />
              </n-form-item>
            </n-col>
          </n-row>

          <n-form-item label="存储大小" path="config.storage">
            <n-input v-model:value="formData.config.storage" placeholder="例如：1Gi 或 10Gi" />
          </n-form-item>

          <n-divider>环境变量</n-divider>
          <n-form-item label="环境变量">
            <n-space vertical style="width: 100%;">
              <n-row v-for="(env, index) in formData.config.env" :key="index" :gutter="12" align="center">
                <n-col :span="10">
                  <n-input v-model:value="env.name" placeholder="变量名" />
                </n-col>
                <n-col :span="10">
                  <n-input v-model:value="env.value" placeholder="变量值" />
                </n-col>
                <n-col :span="4">
                  <n-button type="error" size="small" @click="removeEnv(index)">删除</n-button>
                </n-col>
              </n-row>
              <n-button type="primary" size="small" @click="addEnv" style="width: 100%;">
                <template #icon>
                  <n-icon><add-outline /></n-icon>
                </template>
                添加环境变量
              </n-button>
            </n-space>
          </n-form-item>

          <n-divider>配置预览</n-divider>
          <n-form-item label="JSON配置">
            <n-input
              v-model:value="configPreview"
              type="textarea"
              :rows="8"
              readonly
              placeholder="配置JSON预览"
            />
          </n-form-item>
        </n-form>
      </n-scrollbar>
      <template #action>
        <n-button @click="showModal = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleSubmit">确定</n-button>
      </template>
    </n-modal>
    <StartInstanceModal
      v-model:show="showStartModal"
      :template="selectedTemplate"
      @success="onInstanceStarted"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h, watch } from 'vue'
import { NCard, NButton, NDataTable, NSpace, NModal, NForm, NFormItem, NInput, NInputNumber, NDivider, NRow, NCol, NIcon, NScrollbar, useMessage } from 'naive-ui'
import { AddOutline } from '@vicons/ionicons5'
import { templateApi } from '@/api'
import StartInstanceModal from '@/components/StartInstanceModal.vue'

const message = useMessage()
const templates = ref([])
const showModal = ref(false)
const showStartModal = ref(false)
const selectedTemplate = ref(null)
const isEdit = ref(false)
const loading = ref(false)
const searchKeyword = ref('')
const pagination = ref({ pageSize: 10 })
const formRef = ref(null)

// 默认配置
const defaultConfig = {
  image: 'codercom/code-server:latest',
  namespace: 'default',
  replicas: 1,
  container_port: 8080,
  cpu_limit: '1',
  memory_limit_str: '1Gi',
  cpu_request: '0.5',
  memory_request: '512Mi',
  storage: '1Gi',
  memory_limit: 1024,
  env: []
}

const formData = ref({
  name: '',
  description: '',
  config: { ...defaultConfig }
})

const rules = {
  name: [{ required: true, message: '请输入模板名称' }],
  'config.image': [{ required: true, message: '请输入镜像地址' }],
  'config.namespace': [{ required: true, message: '请输入命名空间' }],
  'config.replicas': [{ required: true, message: '请输入副本数', type: 'number' }],
  'config.container_port': [{ required: true, message: '请输入容器端口', type: 'number' }],
  'config.cpu_limit': [{ required: true, message: '请输入CPU限制' }],
  'config.memory_limit_str': [{ required: true, message: '请输入内存限制' }],
  'config.storage': [{ required: true, message: '请输入存储大小' }]
}

// 配置预览
const configPreview = computed(() => {
  return JSON.stringify(formData.value.config, null, 2)
})

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '模板名称', key: 'name', width: 150 },
  { title: '描述', key: 'description', ellipsis: { tooltip: true } },
  { title: '镜像', key: 'image', ellipsis: { tooltip: true }, width: 200 },
  { title: '副本数', key: 'replicas', width: 80 },
  { title: '创建时间', key: 'created_at', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render: (row) => h(NSpace, {}, () => [
      h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => '编辑'),
      h(NButton, { size: 'small', type: 'success', onClick: () => handleStartClick(row) }, () => '启动实例'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
]

async function loadTemplates() {
  try {
    const params = { limit: pagination.value.pageSize }
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }
    const res = await templateApi.getList(params)
    // 解析配置以显示关键信息
    templates.value = (res.templates || []).map(t => {
      let config = {}
      try {
        config = t.config ? JSON.parse(t.config) : {}
      } catch (e) {
        config = {}
      }
      return {
        ...t,
        image: config.image || '-',
        replicas: config.replicas || 1
      }
    })
    pagination.value.pageCount = Math.ceil((res.total || 0) / pagination.value.pageSize)
  } catch (err) {
    console.error(err)
  }
}

function parseConfig(configStr) {
  if (!configStr) return { ...defaultConfig }
  try {
    const parsed = JSON.parse(configStr)
    return {
      ...defaultConfig,
      ...parsed,
      env: parsed.env || []
    }
  } catch (e) {
    return { ...defaultConfig }
  }
}

function handleAddClick() {
  isEdit.value = false
  formData.value = {
    name: '',
    description: '',
    config: { ...defaultConfig, env: [] }
  }
  showModal.value = true
}

function handleEdit(row) {
  isEdit.value = true
  formData.value = {
    ...row,
    config: parseConfig(row.config)
  }
  showModal.value = true
}

function handleStartClick(row) {
  selectedTemplate.value = row
  showStartModal.value = true
}

function addEnv() {
  formData.value.config.env.push({ name: '', value: '' })
}

function removeEnv(index) {
  formData.value.config.env.splice(index, 1)
}

async function handleSubmit() {
  // 过滤掉空的环境变量
  const validEnv = formData.value.config.env.filter(e => e.name.trim() !== '')
  formData.value.config.env = validEnv

  // 构建提交数据
  const submitData = {
    name: formData.value.name,
    description: formData.value.description,
    config: JSON.stringify(formData.value.config)
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

onMounted(loadTemplates)
</script>

<style scoped>
.templates {
  padding: 0;
}

:deep(.n-modal-body) {
  max-height: 60vh;
}
</style>
