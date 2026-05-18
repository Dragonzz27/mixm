<template>
  <div class="bindings">
    <n-card title="内容-分类绑定管理">
      <template #header-extra>
        <n-button type="primary" @click="showAddModal = true">添加绑定</n-button>
      </template>
      <n-tabs v-model:value="activeTab" type="card">
        <n-tab-pane name="byContent" tab="按内容查看">
          <n-space style="margin-bottom: 16px;">
            <n-select v-model:value="selectedContentId" :options="contentOptions" filterable placeholder="选择内容" style="width: 300px;" @update:value="loadBindingsByContent" />
            <n-button @click="selectedContentId = null; loadAllBindings()">显示全部</n-button>
          </n-space>
          <n-data-table :columns="columns" :data="bindings" :bordered="false" />
        </n-tab-pane>
        <n-tab-pane name="byCategory" tab="按分类查看">
          <n-space style="margin-bottom: 16px;">
            <n-select v-model:value="selectedCategoryId" :options="categoryOptions" filterable placeholder="选择分类" style="width: 300px;" @update:value="loadBindingsByCategory" />
            <n-button @click="selectedCategoryId = null; loadAllBindings()">显示全部</n-button>
          </n-space>
          <n-data-table :columns="columns" :data="bindings" :bordered="false" />
        </n-tab-pane>
      </n-tabs>
    </n-card>
    <n-modal v-model:show="showAddModal" preset="dialog" title="添加绑定">
      <n-form :model="formData" :rules="rules">
        <!-- 此处搜索为客户端搜索，内网访问带宽足够，为MVP方案，后续可迭代升级为服务端搜索 -->
        <n-form-item label="内容" path="content_id">
          <n-select v-model:value="formData.content_id" :options="contentOptions" filterable clearable placeholder="搜索并选择内容" />
        </n-form-item>
        <n-form-item label="分类" path="category_id">
          <n-select v-model:value="formData.category_id" :options="categoryOptions" filterable clearable placeholder="搜索并选择分类" />
        </n-form-item>
        <n-form-item label="排序" path="sort_order">
          <n-input-number v-model:value="formData.sort_order" :min="0" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showAddModal = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleAdd">确定</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { NCard, NButton, NDataTable, NSpace, NModal, NForm, NFormItem, NSelect, NInputNumber, NTabs, NTabPane, useMessage } from 'naive-ui'
import { bindingApi, contentApi, categoryApi } from '@/api'
import { useContentTypeStore } from '@/stores/contentType'

const message = useMessage()
const contentTypeStore = useContentTypeStore()
const bindings = ref([])
const contents = ref([])
const categories = ref([])
const showAddModal = ref(false)
const loading = ref(false)
const activeTab = ref('byContent')
const selectedContentId = ref(null)
const selectedCategoryId = ref(null)

const formData = ref({ content_id: null, category_id: null, sort_order: 0 })
const rules = {
  content_id: [{ required: true, message: '请选择内容' }],
  category_id: [{ required: true, message: '请选择分类' }]
}

const contentOptions = computed(() => {
  return contents.value.map(c => ({ label: `${c.id} - ${c.title}`, value: c.id }))
})

const categoryOptions = computed(() => {
  return categories.value.map(c => ({ label: `${c.id} - ${c.name}`, value: c.id }))
})

const columns = [
  { title: '序号', key: 'index', width: 70, render: (_, index) => index + 1 },
  { title: '内容ID', key: 'content_id', width: 80 },
  { title: '分类ID', key: 'category_id', width: 80 },
  { title: '内容类型', key: 'content_type', width: 100, render: (row) => contentTypeStore.getTypeName(row.content_type) },
  { title: '排序', key: 'sort_order', width: 80 },
  { title: '创建时间', key: 'created_at', width: 180 },
  {
    title: '操作',
    key: 'actions',
    render: (row) => h(NSpace, {}, () => [
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
]

async function loadContents() {
  try {
    const res = await contentApi.getList({ limit: 1000 })
    contents.value = res.contents || []
  } catch (err) {
    console.error(err)
  }
}

async function loadCategories() {
  try {
    const res = await categoryApi.getList({})
    categories.value = res || []
  } catch (err) {
    console.error(err)
  }
}

async function loadAllBindings() {
  selectedContentId.value = null
  selectedCategoryId.value = null
  if (contents.value.length > 0) {
    const allBindings = []
    for (const content of contents.value.slice(0, 50)) {
      try {
        const res = await bindingApi.getByContent(content.id)
        if (res && res.length > 0) {
          for (const binding of res) {
            const fullContent = contents.value.find(c => c.id === binding.content_id)
            allBindings.push({
              ...binding,
              content_type: fullContent?.content_type || binding.content_type
            })
          }
        }
      } catch (err) {
        console.error(err)
      }
    }
    bindings.value = allBindings
  }
}

async function loadBindingsByContent(contentId) {
  selectedCategoryId.value = null
  if (contentId) {
    try {
      const res = await bindingApi.getByContent(contentId)
      const fullContent = contents.value.find(c => c.id === contentId)
      bindings.value = (res || []).map(b => ({
        ...b,
        content_type: fullContent?.content_type || b.content_type
      }))
    } catch (err) {
      console.error(err)
    }
  } else {
    loadAllBindings()
  }
}

async function loadBindingsByCategory(categoryId) {
  selectedContentId.value = null
  if (categoryId) {
    try {
      const res = await bindingApi.getByCategory(categoryId)
      const bindingContents = []
      for (const binding of res || []) {
        const fullContent = contents.value.find(c => c.id === binding.content_id)
        if (fullContent) {
          bindingContents.push({
            ...binding,
            content_type: fullContent.content_type
          })
        }
      }
      bindings.value = bindingContents
    } catch (err) {
      console.error(err)
    }
  } else {
    loadAllBindings()
  }
}

async function handleAdd() {
  loading.value = true
  try {
    await bindingApi.create(formData.value)
    message.success('绑定成功')
    showAddModal.value = false
    formData.value = { content_id: null, category_id: null, sort_order: 0 }
    if (selectedContentId.value) {
      loadBindingsByContent(selectedContentId.value)
    } else if (selectedCategoryId.value) {
      loadBindingsByCategory(selectedCategoryId.value)
    } else {
      loadAllBindings()
    }
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function handleDelete(id) {
  try {
    await bindingApi.delete(id)
    message.success('解除绑定成功')
    if (selectedContentId.value) {
      loadBindingsByContent(selectedContentId.value)
    } else if (selectedCategoryId.value) {
      loadBindingsByCategory(selectedCategoryId.value)
    } else {
      loadAllBindings()
    }
  } catch (err) {
    console.error(err)
  }
}

onMounted(() => {
  loadContents()
  loadCategories()
  loadAllBindings()
})
</script>
