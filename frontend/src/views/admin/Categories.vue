<template>
  <div class="categories">
    <n-card title="分类管理">
      <template #header-extra>
        <n-space>
          <n-button type="warning" @click="showAiModal = true">智能体生成</n-button>
          <n-button type="primary" @click="handleAddRoot">添加顶级分类</n-button>
        </n-space>
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
              remote
              :row-key="row => row.id"
              :checked-row-keys="selectedUnbound"
              @update:checked-row-keys="val => selectedUnbound = val"
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

    <n-modal v-model:show="showAiModal" preset="dialog" title="智能体生成知识树" style="width: 90%; max-width: 1200px;">
      <div class="ai-modal-content">
        <div class="ai-input-panel">
          <n-form label-placement="top">
            <n-form-item label="课程名称">
              <n-input v-model:value="aiForm.course_name" placeholder="请输入课程名称" />
            </n-form-item>
            <n-form-item label="考试大纲">
              <n-input
                v-model:value="aiForm.outline_text"
                type="textarea"
                :rows="15"
                placeholder="请输入考试大纲内容"
              />
            </n-form-item>
            <n-button type="primary" :loading="aiLoading" @click="handleAiGenerate" :disabled="!aiForm.outline_text">
              生成知识树
            </n-button>
          </n-form>
        </div>
        <div class="ai-result-panel">
          <div class="result-header">
            <span>知识树结构 (可编辑)</span>
            <n-space>
              <n-button size="small" @click="showJsonModal = true" :disabled="!knowledgeTree">查看原始JSON</n-button>
              <n-button type="success" size="small" :disabled="!knowledgeTree" @click="handleSaveTree">
                保存到分类
              </n-button>
            </n-space>
          </div>
          <div class="tree-container">
            <div v-if="isStreaming" class="streaming-view">
              <div class="streaming-header">
                <n-spin size="small" />
                <span>AI 正在生成中...</span>
              </div>
              <pre class="streaming-content">{{ streamingContent }}</pre>
            </div>
            <div v-else-if="knowledgeTree" class="tree-list">
              <div v-for="(chapter, ci) in knowledgeTree.chapters" :key="ci" class="tree-chapter">
                <div class="tree-item-header">
                  <span class="level-badge">章节</span>
                  <n-input v-model:value="chapter.name" size="small" style="width: 200px;" />
                  <n-button size="small" quaternary @click="chapter.sections.push({ name: '', knowledge_points: [] })">+ 小节</n-button>
                  <n-button size="small" type="error" quaternary @click="knowledgeTree.chapters.splice(ci, 1)">删除</n-button>
                </div>

                <div v-for="(section, si) in chapter.sections" :key="si" class="tree-section">
                  <div class="tree-item-header">
                    <span class="level-badge">小节</span>
                    <n-input v-model:value="section.name" size="small" style="width: 180px;" />
                    <n-button size="small" quaternary @click="section.knowledge_points.push({ name: '', elements: [] })">+ 知识点</n-button>
                    <n-button size="small" type="error" quaternary @click="chapter.sections.splice(si, 1)">删除</n-button>
                  </div>

                  <div v-for="(kp, ki) in section.knowledge_points" :key="ki" class="tree-kp">
                    <div class="tree-item-header">
                      <span class="level-badge">知识点</span>
                      <n-input v-model:value="kp.name" size="small" style="width: 160px;" />
                      <n-button size="small" quaternary @click="kp.elements.push({ name: '' })">+ 知识元</n-button>
                      <n-button size="small" type="error" quaternary @click="section.knowledge_points.splice(ki, 1)">删除</n-button>
                    </div>

                    <div v-for="(elem, ei) in kp.elements" :key="ei" class="tree-element">
                      <span class="level-badge small">知识元</span>
                      <n-input v-model:value="elem.name" size="small" style="width: 200px;" />
                      <n-button size="small" type="error" quaternary @click="kp.elements.splice(ei, 1)">×</n-button>
                    </div>
                    <n-button size="small" quaternary style="margin-left: 80px;" @click="kp.elements.push({ name: '' })">+ 添加知识元</n-button>
                  </div>
                </div>
              </div>
              <n-button type="primary" quaternary @click="knowledgeTree.chapters.push({ name: '', sections: [] })" style="margin-top: 12px;">+ 添加章节</n-button>
            </div>
            <div v-else class="empty-result">
              <n-empty description="请输入考试大纲并点击生成" />
            </div>
          </div>
        </div>

        <n-modal v-model:show="showJsonModal" preset="dialog" title="原始JSON" style="width: 600px;">
          <pre class="json-viewer">{{ originalJson }}</pre>
          <template #action>
            <n-button @click="showJsonModal = false">关闭</n-button>
          </template>
        </n-modal>
      </div>
      <template #action>
        <n-button @click="showAiModal = false">关闭</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NCard, NButton, NInput, NSpace, NModal, NForm, NFormItem, NInputNumber, NSwitch, NSpin, NEmpty, NDataTable, useMessage, NIcon } from 'naive-ui'
import { categoryApi, bindingApi, knowledgeTreeApi } from '@/api'
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
const unboundPagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    loadUnboundContents(page)
  },
  onUpdatePageSize: (pageSize) => {
    unboundPagination.value.pageSize = pageSize
    loadUnboundContents(1)
  }
})
const boundContents = ref([])
const selectedUnbound = ref([])
const selectedBound = ref([])
const loadingUnbound = ref(false)
const loadingBound = ref(false)
const unboundSearchKey = ref('')

const showAiModal = ref(false)
const aiLoading = ref(false)
const aiForm = ref({ course_name: '', outline_text: '' })
const knowledgeTree = ref(null)
const streamingContent = ref('')
const isStreaming = ref(false)
const showJsonModal = ref(false)
const originalJson = ref('')

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
  unboundPagination.value.page = 1
  unboundPagination.value.pageSize = 10
  unboundPagination.value.itemCount = 0
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
    const pageSize = unboundPagination.value.pageSize
    const skip = (page - 1) * pageSize
    const params = {
      keyword: unboundSearchKey.value || undefined,
      skip: skip,
      limit: pageSize
    }
    const res = await bindingApi.getUnboundContents(currentCategoryId.value, params)
    unboundContents.value = res.contents || []
    unboundPagination.value.itemCount = res.total || 0
    unboundPagination.value.page = page

    const maxPage = Math.max(1, Math.ceil((unboundPagination.value.itemCount || 0) / pageSize))
    if (unboundPagination.value.page > maxPage) {
      return loadUnboundContents(maxPage)
    }
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

async function handleAiGenerate() {
  aiLoading.value = true
  isStreaming.value = true
  streamingContent.value = ''
  knowledgeTree.value = null
  try {
    knowledgeTreeApi.analyzeStream(
      {
        course_name: aiForm.value.course_name || undefined,
        outline_text: aiForm.value.outline_text
      },
      (chunk) => {
        streamingContent.value += chunk
      },
      (data) => {
        try {
          knowledgeTree.value = data
          originalJson.value = JSON.stringify(data, null, 2)
          message.success('知识树生成成功')
        } catch (e) {
          message.error('解析失败')
        }
        isStreaming.value = false
        aiLoading.value = false
      },
      (err) => {
        console.error(err)
        message.error(err.message || '调用失败')
        isStreaming.value = false
        aiLoading.value = false
      }
    )
  } catch (err) {
    console.error(err)
    message.error('调用失败')
    isStreaming.value = false
    aiLoading.value = false
  }
}

async function handleSaveTree() {
  if (!knowledgeTree.value) return
  try {
    const res = await knowledgeTreeApi.save({
      knowledge_tree: knowledgeTree.value
    })
    if (res.success) {
      message.success('保存成功')
      showAiModal.value = false
      loadCategories()
    } else {
      message.error(res.message || '保存失败')
    }
  } catch (err) {
    console.error(err)
    message.error('保存失败')
  }
}
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

.ai-modal-content {
  display: flex;
  gap: 24px;
  min-height: 60vh;
}

.ai-input-panel {
  width: 350px;
  flex-shrink: 0;
}

.ai-result-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 500;
}

.tree-container {
  flex: 1;
  overflow-y: auto;
  max-height: 55vh;
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;
}

.tree-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tree-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.level-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  background: #e0e0e0;
  color: #666;
  font-weight: 500;
}

.level-badge.small {
  padding: 1px 6px;
  font-size: 11px;
}

.tree-chapter {
  padding: 12px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
}

.tree-section {
  margin-left: 24px;
  padding: 10px;
  background: #f9f9f9;
  border-radius: 6px;
  margin-top: 8px;
}

.tree-kp {
  margin-left: 20px;
  padding: 8px;
  background: #f5f5f5;
  border-radius: 4px;
  margin-top: 8px;
}

.tree-element {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 6px 0 6px 20px;
}

.empty-result {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
}

.streaming-view {
  padding: 16px;
}

.streaming-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  color: #666;
  font-size: 14px;
}

.streaming-content {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 6px;
  max-height: 400px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  color: #333;
}

.json-viewer {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 6px;
  max-height: 500px;
  overflow: auto;
  font-family: monospace;
  font-size: 12px;
  white-space: pre-wrap;
}
</style>
