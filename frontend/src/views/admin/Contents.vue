<template>
  <div class="contents">
    <n-card title="内容管理">
      <template #header-extra>
        <n-space>
          <n-input 
            v-model:value="searchKey" 
            placeholder="搜索标题" 
            clearable 
            style="width: 200px"
            @keyup.enter="loadContents"
          />
          <n-button type="primary" @click="handleAddClick">添加内容</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="contents"
        :pagination="pagination"
        :bordered="false"
      />
    </n-card>
    <n-modal v-model:show="showAddModal" preset="dialog" :title="isEdit ? '编辑内容' : '添加内容'" style="width: 80%;">
      <div class="modal-content">
        <div class="left-panel">
          <n-form :model="formData" :rules="rules" ref="formRef" label-placement="left" label-width="80">
            <n-form-item label="标题" path="title">
              <n-input v-model:value="formData.title" />
            </n-form-item>
            <n-form-item label="类型" path="content_type">
              <n-select v-model:value="formData.content_type" :options="typeOptions" placeholder="选择类型" />
            </n-form-item>
            <n-form-item label="模板" path="template_id">
              <n-select v-model:value="formData.template_id" :options="templateOptions" clearable placeholder="选择模板" />
            </n-form-item>
            <n-form-item label="链接" path="url">
              <n-input v-model:value="formData.url" placeholder="资源链接" />
            </n-form-item>
            <n-form-item label="状态" path="status">
              <n-select v-model:value="formData.status" :options="statusOptions" />
            </n-form-item>
          </n-form>
        </div>
        <div class="right-panel">
          <div class="editor-label">内容 (超文本/Markdown)</div>
          <RichEditor v-model="formData.content" />
        </div>
      </div>
      <template #action>
        <n-button @click="showAddModal = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleSubmit">确定</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { NCard, NButton, NInput, NDataTable, NSpace, NModal, NForm, NFormItem, NSelect, useMessage } from 'naive-ui'
import RichEditor from '@/components/RichEditor.vue'
import { contentApi, templateApi } from '@/api'
import { useContentTypeStore } from '@/stores/contentType'

const message = useMessage()
const contentTypeStore = useContentTypeStore()
const formRef = ref(null)
const searchKey = ref('')
const contents = ref([])
const templates = ref([])
const showAddModal = ref(false)
const isEdit = ref(false)
const loading = ref(false)
const pagination = ref({ pageSize: 10 })

const formData = ref({ title: '', content: '', content_type: null, template_id: null, url: '', status: 'draft' })

const rules = {
  title: [{ required: true, message: '请输入标题' }],
  content_type: [{ required: true, message: '请选择类型' }],
  status: [{ required: true, message: '请选择状态' }]
}

const typeOptions = computed(() => {
  return contentTypeStore.contentTypes.map(t => ({ label: t.name, value: t.code }))
})

const templateOptions = computed(() => {
  return templates.value.map(t => ({ label: t.name, value: t.id }))
})

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '已发布', value: 'published' },
  { label: '已归档', value: 'archived' }
]

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '类型', key: 'content_type', width: 100, render: (row) => contentTypeStore.getTypeName(row.content_type) },
  { title: '模板ID', key: 'template_id', width: 80, render: (row) => row.template_id || '-' },
  { title: '状态', key: 'status', width: 100, render: (row) => h('span', row.status === 'published' ? '已发布' : row.status === 'archived' ? '已归档' : '草稿') },
  { title: '创建时间', key: 'created_at', width: 180 },
  {
    title: '操作',
    key: 'actions',
    render: (row) => h(NSpace, {}, () => [
      h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => '编辑'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
]

async function loadContents() {
  try {
    const res = await contentApi.getList({
      skip: 0,
      limit: 100,
      keyword: searchKey.value || undefined
    })
    contents.value = res.contents || []
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

function handleAddClick() {
  isEdit.value = false
  formData.value = { title: '', content: '', content_type: '', template_id: null, url: '', status: 'draft' }
  showAddModal.value = true
}

async function handleEdit(row) {
  isEdit.value = true
  formData.value = {
    ...row,
    content: row.content || '',
    content_type: row.content_type || null,
    template_id: row.template_id || null,
    status: row.status || 'draft'
  }
  loading.value = true
  try {
    const detail = await contentApi.getDetail(row.id)
    formData.value = {
      ...row,
      ...detail,
      content: detail?.content || formData.value.content || '',
      content_type: detail?.content_type || formData.value.content_type || null,
      template_id: detail?.template_id || formData.value.template_id || null,
      status: detail?.status || formData.value.status || 'draft'
    }
  } catch (err) {
    console.error(err)
    message.warning('未能加载完整内容，将使用当前列表数据进行编辑')
  } finally {
    loading.value = false
  }

  showAddModal.value = true
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  
  loading.value = true
  try {
    const data = {
      ...formData.value,
      content: formData.value.content || '',
      content_type: formData.value.content_type || 'article',
      template_id: formData.value.template_id || null
    }
    if (isEdit.value) {
      await contentApi.update(formData.value.id, data)
      message.success('更新成功')
    } else {
      await contentApi.create(data)
      message.success('添加成功')
    }
    showAddModal.value = false
    loadContents()
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function handleDelete(id) {
  try {
    await contentApi.delete(id)
    message.success('删除成功')
    loadContents()
  } catch (err) {
    console.error(err)
  }
}

onMounted(() => {
  loadContents()
  loadTemplates()
})
</script>

<style scoped>
.modal-content {
  display: flex;
  gap: 24px;
  min-height: 70vh;
}

.left-panel {
  width: 320px;
  flex-shrink: 0;
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.editor-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}
</style>