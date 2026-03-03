<template>
  <div class="categories">
    <n-card title="分类管理">
      <template #header-extra>
        <n-button type="primary" @click="handleAddRoot">添加顶级分类</n-button>
      </template>
      <n-spin :show="loading">
        <div class="tree-wrapper">
          <CategoryTreeNode
            v-for="category in categoryTree"
            :key="category.id"
            :category="category"
            :expanded-keys="expandedKeys"
            :depth="0"
            @toggle="toggleExpand"
            @add-child="handleAddChild"
            @edit="handleEdit"
            @delete="handleDelete"
            @bind="openBindModal"
          />
          <n-empty v-if="categoryTree.length === 0" description="暂无分类" />
        </div>
      </n-spin>
    </n-card>

    <n-modal v-model:show="showAddModal" preset="dialog" :title="isEdit ? '编辑分类' : (isChild ? '添加子分类' : '添加顶级分类')">
      <n-form :model="formData" :rules="rules" ref="formRef">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formData.name" />
        </n-form-item>
        <n-form-item label="父分类" path="parent_id">
          <n-input :value="parentName" disabled />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formData.description" type="textarea" />
        </n-form-item>
        <n-form-item label="排序" path="sort_order">
          <n-input-number v-model:value="formData.sort_order" :min="0" />
        </n-form-item>
        <n-form-item label="状态" path="is_active">
          <n-switch v-model:value="formData.is_active">
            <template #checked>启用</template>
            <template #unchecked>禁用</template>
          </n-switch>
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showAddModal = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleSubmit">确定</n-button>
      </template>
    </n-modal>

    <n-modal v-model:show="showBindModal" preset="card" title="绑定内容" style="width: 900px;">
      <div class="bind-transfer">
        <div class="transfer-panel">
          <div class="panel-header">
            <span>未绑定内容</span>
            <n-space>
              <n-input v-model:value="unboundSearchKey" placeholder="搜索标题" size="small" style="width: 150px;" @keyup.enter="loadUnboundContents(1)" />
            </n-space>
          </div>
          <n-spin :show="loadingUnbound">
            <n-data-table
              :columns="unboundColumns"
              :data="unboundContents"
              :bordered="false"
              :max-height="350"
              :pagination="unboundPagination"
              :row-key="row => row.id"
              :checked-row-keys="selectedUnbound"
              @update:checked-row-keys="val => selectedUnbound = val"
              @update:page="page => loadUnboundContents(page)"
              @update:page-size="pageSize => { unboundPagination.pageSize = pageSize; loadUnboundContents(1); }"
            />
          </n-spin>
        </div>
        <div class="transfer-actions">
          <n-button type="primary" :disabled="selectedUnbound.length === 0" @click="handleBindSelected">
            添加 >
          </n-button>
          <n-button type="warning" :disabled="selectedBound.length === 0" @click="handleUnbindSelected">
            < 移除
          </n-button>
        </div>
        <div class="transfer-panel">
          <div class="panel-header">
            <span>已绑定内容</span>
          </div>
          <n-spin :show="loadingBound">
            <n-data-table
              :columns="boundColumns"
              :data="boundContents"
              :bordered="false"
              :max-height="350"
              :pagination="{ pageSize: 10 }"
              :row-key="row => row.id"
              :checked-row-keys="selectedBound"
              @update:checked-row-keys="val => selectedBound = val"
            />
          </n-spin>
        </div>
      </div>
      <template #action>
        <n-button @click="async () => { showBindModal = false; await loadCategories(); }">关闭</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NCard, NButton, NInput, NSpace, NModal, NForm, NFormItem, NInputNumber, NSwitch, NSpin, NEmpty, NDataTable, useMessage, NIcon } from 'naive-ui'
import { categoryApi, bindingApi } from '@/api'
import { useContentTypeStore } from '@/stores/contentType'
import { ChevronForwardOutline } from '@vicons/ionicons5'
import CategoryTreeNode from '@/components/CategoryTreeNode.vue'

const message = useMessage()
const contentTypeStore = useContentTypeStore()

const loading = ref(false)
const categoryTree = ref([])
const expandedKeys = ref([])
const showAddModal = ref(false)
const isEdit = ref(false)
const isChild = ref(false)
const parentName = ref('')

const formData = ref({ name: '', parent_id: null, description: '', sort_order: 0, is_active: true })
const rules = {
  name: [{ required: true, message: '请输入名称' }]
}

const showBindModal = ref(false)
const currentCategoryId = ref(null)
const unboundContents = ref([])
const unboundPagination = ref({ page: 1, pageSize: 10, total: 0 })
const boundContents = ref([])
const selectedUnbound = ref([])
const selectedBound = ref([])
const loadingUnbound = ref(false)
const loadingBound = ref(false)
const unboundSearchKey = ref('')

const unboundColumns = [
  { type: 'selection', width: 50 },
  { title: 'ID', key: 'id', width: 60 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '类型', key: 'content_type', width: 100, render: (row) => contentTypeStore.getTypeName(row.content_type) },
  { title: '状态', key: 'status', width: 80, render: (row) => h('span', row.status === 'published' ? '已发布' : row.status === 'archived' ? '已归档' : '草稿') }
]

const boundColumns = [
  { type: 'selection', width: 50 },
  { title: 'ID', key: 'id', width: 60 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '类型', key: 'content_type', width: 100, render: (row) => contentTypeStore.getTypeName(row.content_type) },
  { title: '状态', key: 'status', width: 80, render: (row) => h('span', row.status === 'published' ? '已发布' : row.status === 'archived' ? '已归档' : '草稿') }
]

async function loadCategories() {
  loading.value = true
  try {
    const res = await categoryApi.getTree({max_depth:100})
    categoryTree.value = res || []
    await loadBoundCounts()
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function loadBoundCounts() {
  const allIds = getAllCategoryIds(categoryTree.value)
  if (allIds.length === 0) return
  
  try {
    const res = await bindingApi.getCounts({ category_ids: allIds })
    const setCounts = (list) => {
      for (const cat of list) {
        cat.bound_count = res[cat.id] || 0
        if (cat.children?.length) {
          setCounts(cat.children)
        }
      }
    }
    setCounts(categoryTree.value)
  } catch (err) {
    console.error(err)
  }
}

function getAllCategoryIds(list) {
  const ids = []
  for (const cat of list) {
    ids.push(cat.id)
    if (cat.children?.length) {
      ids.push(...getAllCategoryIds(cat.children))
    }
  }
  return ids
}

function toggleExpand(categoryId) {
  const index = expandedKeys.value.indexOf(categoryId)
  if (index > -1) {
    expandedKeys.value.splice(index, 1)
  } else {
    expandedKeys.value.push(categoryId)
  }
}

function handleAddRoot() {
  isEdit.value = false
  isChild.value = false
  parentName.value = '顶级分类'
  formData.value = { name: '', parent_id: null, description: '', sort_order: 0, is_active: true }
  showAddModal.value = true
}

function handleAddChild(category) {
  isEdit.value = false
  isChild.value = true
  parentName.value = category.name
  formData.value = { name: '', parent_id: category.id, description: '', sort_order: 0, is_active: true }
  showAddModal.value = true
}

function handleEdit(category) {
  isEdit.value = true
  isChild.value = !!category.parent_id
  parentName.value = isChild.value ? getParentName(category.parent_id) : '顶级分类'
  formData.value = { 
    id: category.id,
    name: category.name, 
    parent_id: category.parent_id, 
    description: category.description || '', 
    sort_order: category.sort_order || 0, 
    is_active: category.is_active 
  }
  showAddModal.value = true
}

function getParentName(parentId) {
  const findParent = (list, id) => {
    for (const cat of list) {
      if (cat.id === id) return cat.name
      if (cat.children?.length) {
        const found = findParent(cat.children, id)
        if (found) return found
      }
    }
    return ''
  }
  return findParent(categoryTree.value, parentId)
}

async function handleSubmit() {
  loading.value = true
  try {
    if (isEdit.value) {
      await categoryApi.update(formData.value.id, formData.value)
      message.success('更新成功')
    } else {
      await categoryApi.create(formData.value)
      message.success('添加成功')
    }
    showAddModal.value = false
    await loadCategories()
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function handleDelete(id) {
  try {
    await categoryApi.delete(id)
    message.success('删除成功')
    await loadCategories()
  } catch (err) {
    console.error(err)
  }
}

async function openBindModal(category) {
  currentCategoryId.value = category.id
  showBindModal.value = true
  unboundSearchKey.value = ''
  unboundPagination.value = { page: 1, pageSize: 10, total: 0 }
  selectedUnbound.value = []
  selectedBound.value = []
  await Promise.all([
    loadUnboundContents(1),
    loadBoundContents()
  ])
}

async function loadUnboundContents(page = 1) {
  loadingUnbound.value = true
  try {
    const skip = (page - 1) * unboundPagination.value.pageSize
    const params = {
      keyword: unboundSearchKey.value || undefined,
      skip: skip,
      limit: unboundPagination.value.pageSize
    }
    const res = await bindingApi.getUnboundContents(currentCategoryId.value, params)
    unboundContents.value = res.contents || []
    unboundPagination.value.total = res.total || 0
    unboundPagination.value.page = page
  } catch (err) {
    console.error(err)
    unboundContents.value = []
  } finally {
    loadingUnbound.value = false
  }
}

async function loadBoundContents() {
  loadingBound.value = true
  try {
    const res = await bindingApi.getCategoryDetail(currentCategoryId.value)
    if (res && res.contents) {
      boundContents.value = res.contents.map(c => ({
        id: c.id,
        title: c.title,
        content_type: c.content_type,
        status: c.status
      }))
    } else {
      boundContents.value = []
    }
  } catch (err) {
    console.error(err)
    boundContents.value = []
  } finally {
    loadingBound.value = false
  }
}

async function handleBindSelected() {
  if (selectedUnbound.value.length === 0) return
  
  try {
    await bindingApi.batchBind(currentCategoryId.value, { content_ids: selectedUnbound.value })
    message.success(`已添加 ${selectedUnbound.value.length} 个内容`)
    selectedUnbound.value = []
    await Promise.all([loadUnboundContents(), loadBoundContents()])
    await loadCategories()
  } catch (err) {
    console.error(err)
  }
}

async function handleUnbindSelected() {
  if (selectedBound.value.length === 0) return
  
  try {
    await bindingApi.batchUnbind(currentCategoryId.value, { content_ids: selectedBound.value })
    message.success(`已移除 ${selectedBound.value.length} 个内容`)
    selectedBound.value = []
    await Promise.all([loadUnboundContents(), loadBoundContents()])
    await loadCategories()
  } catch (err) {
    console.error(err)
  }
}

onMounted(() => {
  loadCategories()
})
</script>

<style scoped>
.tree-wrapper {
  padding: 8px 0;
}

.bind-transfer {
  display: flex;
  gap: 16px;
}

.transfer-panel {
  flex: 1;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #f5f5f5;
  border-bottom: 1px solid #e0e0e0;
  font-weight: 500;
}

.transfer-actions {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
}
</style>
