<template>
  <div class="client-contents">
    <div class="page-header">
      <h1>内容列表</h1>
    </div>

    <div class="content-layout">
      <aside class="category-sidebar">
        <div class="sidebar-header">
          <span>选择分类</span>
        </div>
        <div class="tree-wrapper">
          <n-input 
            v-model:value="searchKey" 
            placeholder="搜索分类..." 
            clearable 
            size="small"
            style="margin-bottom: 12px"
          />
          <n-tree
            :data="treeData"
            :selected-keys="selectedKeys"
            :expanded-keys="expandedKeys"
            :pattern="searchKey"
            :show-search="false"
            block-line
            selectable
            @update:selected-keys="handleSelect"
            @update:expanded-keys="handleExpand"
          />
        </div>
      </aside>

      <main class="content-main">
        <div class="section-cards" v-if="level3Categories.length > 0">
          <h3 class="section-title">{{ selectedLevel2Name || selectedLevel1Name || '请选择分类' }}</h3>
          <n-grid :cols="3" :x-gap="20" :y-gap="20" :item-responsive="true">
            <n-gi :span="1" :xs="24" :sm="12" :md="8" :lg="6" v-for="cat in level3Categories" :key="cat.id">
              <div class="category-card" @click="goToDetail(cat.id)">
                <div class="card-cover" :style="{ background: getCoverColor(cat.id) }">
                  <span class="card-level">L{{ cat.level }}</span>
                </div>
                <div class="card-body">
                  <h4 class="card-title">{{ cat.name }}</h4>
                  <p class="card-desc">{{ cat.description || '点击进入章节详情' }}</p>
                </div>
              </div>
            </n-gi>
          </n-grid>
        </div>

        <div class="empty-state" v-else>
          <n-empty v-if="!selectedLevel1Id" description="请从左侧选择分类" />
          <n-empty v-else-if="!selectedLevel2Id" description="该分类下暂无子章节" />
          <n-empty v-else description="暂无子章节" />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NGrid, NGi, NSpin, NEmpty, NTree, NInput } from 'naive-ui'
import { categoryApi } from '@/api'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const categoryTree = ref([])
const selectedLevel1Id = ref(null)
const selectedLevel2Id = ref(null)
const expandedKeys = ref([])
const searchKey = ref('')

const selectedKeys = computed(() => {
  if (selectedLevel2Id.value) {
    return [selectedLevel2Id.value]
  }
  if (selectedLevel1Id.value) {
    return [selectedLevel1Id.value]
  }
  return []
})

const colors = ['#18a058', '#2080f0', '#f0a020', '#d03050', '#9c27b0', '#00bcd4']

function getCoverColor(id) {
  return colors[id % colors.length]
}

const selectedLevel1Name = computed(() => {
  if (!selectedLevel1Id.value) return ''
  const cat = categoryTree.value.find(c => c.id === selectedLevel1Id.value)
  return cat?.name || ''
})

const selectedLevel2Name = computed(() => {
  if (!selectedLevel2Id.value) return ''
  for (const cat of categoryTree.value) {
    if (cat.children?.length) {
      const child = cat.children.find(c => c.id === selectedLevel2Id.value)
      if (child) return child.name
    }
  }
  return ''
})

const level3Categories = computed(() => {
  if (!selectedLevel2Id.value) return []
  for (const cat of categoryTree.value) {
    if (cat.children?.length) {
      const child = cat.children.find(c => c.id === selectedLevel2Id.value)
      if (child && child.children) {
        return child.children
      }
    }
  }
  return []
})

const treeData = computed(() => {
  return categoryTree.value.map(cat => ({
    key: cat.id,
    label: cat.name,
    children: cat.children?.map(child => ({
      key: child.id,
      label: child.name
    }))
  }))
})

function handleSelect(keys) {
  if (keys.length === 0) return
  
  const selectedId = keys[0]
  
  for (const cat of categoryTree.value) {
    if (cat.id === selectedId) {
      selectedLevel1Id.value = cat.id
      if (cat.children?.length) {
        selectedLevel2Id.value = cat.children[0].id
        expandedKeys.value = [cat.id]
      } else {
        selectedLevel2Id.value = null
      }
      return
    }
    
    if (cat.children?.length) {
      for (const child of cat.children) {
        if (child.id === selectedId) {
          selectedLevel1Id.value = cat.id
          selectedLevel2Id.value = child.id
          return
        }
      }
    }
  }
}

function handleExpand(keys) {
  expandedKeys.value = keys
}

function goToDetail(categoryId) {
  router.push(`/app/content/${categoryId}?type=category`)
}

async function loadCategoryTree() {
  loading.value = true
  try {
    const res = await categoryApi.getTree({})
    categoryTree.value = res || []
    
    if (categoryTree.value.length > 0) {
      const firstCategory = categoryTree.value[0]
      selectedLevel1Id.value = firstCategory.id
      expandedKeys.value = [firstCategory.id]
      
      if (firstCategory.children?.length) {
        selectedLevel2Id.value = firstCategory.children[0].id
      }
    }
  } catch (err) {
    console.error('Failed to load category tree:', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCategoryTree()
})
</script>

<style scoped>
.client-contents {
  padding-bottom: 24px;
}

.page-header {
  margin-bottom: 16px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.content-layout {
  display: flex;
  gap: 20px;
}

.category-sidebar {
  width: 260px;
  flex-shrink: 0;
  background: white;
  border-radius: 8px;
  border: 1px solid #e0e0e6;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

.sidebar-header {
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 600;
  color: #333;
  border-bottom: 1px solid #e0e0e6;
}

.tree-wrapper {
  padding: 12px;
}

.content-main {
  flex: 1;
  min-width: 0;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin: 0 0 16px;
}

.category-card {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #e0e0e6;
}

.category-card:hover {
  border-color: #18a058;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.card-cover {
  height: 100px;
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  padding: 12px;
}

.card-level {
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.card-body {
  padding: 16px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0 0 8px;
  line-height: 1.4;
}

.card-desc {
  font-size: 13px;
  color: #666;
  margin: 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.empty-state {
  padding: 60px 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

@media (max-width: 768px) {
  .content-layout {
    flex-direction: column;
  }
  
  .category-sidebar {
    width: 100%;
    max-height: 300px;
  }
}
</style>
