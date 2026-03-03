<template>
  <div class="image-repositories">
    <div class="page-header">
      <div class="page-title">
        <div class="page-main-title">镜像</div>
        <span class="title-divider" aria-hidden="true"></span>
        <n-tabs 
          v-model:value="activeTab" 
          type="line" 
          class="image-tabs"
          @update:value="handleTabChange"
        >
          <!-- <n-tab name="my">我的</n-tab>   EDU版本不开放 -->
          <n-tab name="public">公开</n-tab>
        </n-tabs>
      </div>
      <div class="page-actions">
        <n-input
          v-model:value="searchText"
          placeholder="搜索镜像"
          clearable
          style="width: 200px;"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <n-button type="primary" @click="showCreateModal = true">创建镜像仓库</n-button>
      </div>
    </div>

    <n-spin :show="loading">
      <n-card :bordered="false" v-if="repositories.length > 0">
        <n-table :bordered="false" :single-line="false">
          <thead>
            <tr>
              <th>镜像仓库名称</th>
              <th>个数</th>
              <th>创建者</th>
              <th>创建时间</th>
              <th>公开性</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="repo in repositories" :key="repo.id">
              <td>
                <a href="javascript:void(0)" @click="goToDetail(repo.id)" class="repo-name">
                  {{ repo.name }}
                </a>
              </td>
              <td>{{ repo.tags_count }}</td>
              <td>
                <div class="creator-cell">
                  <n-avatar round size="small" class="creator-avatar">
                    {{ repo.user_info?.username?.charAt(0) || '?' }}
                  </n-avatar>
                  <span>{{ repo.user_info?.username || '-' }}</span>
                </div>
              </td>
              <td>{{ formatTime(repo.created_at) }}</td>
              <td>
                <n-tag :type="repo.is_public ? 'success' : 'warning'" size="small">
                  {{ repo.is_public ? '公开' : '私有' }}
                </n-tag>
              </td>
              <td>
                <n-space>
                  <a href="javascript:void(0)" class="action-link" @click="goToDetail(repo.id)">详情</a>
                  <n-dropdown
                    :options="getMoreOptions(repo)"
                    @select="(key) => handleMoreAction(key, repo)"
                  >
                    <a href="javascript:void(0)" class="action-link">更多</a>
                  </n-dropdown>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-card>
      <n-empty v-else description="暂无镜像仓库" style="padding: 80px 0;">
        <template #extra>
          <n-button type="primary" @click="showCreateModal = true">创建第一个仓库</n-button>
        </template>
      </n-empty>

      <div class="pagination" v-if="total > 0">
        <span>共 {{ total }} 条</span>
        <n-pagination
          v-model:page="page"
          :page-size="pageSize"
          :page-slot="7"
          @update:page="handlePageChange"
        />
        <n-select
          v-model:value="pageSize"
          :options="pageSizeOptions"
          style="width: 90px;"
          @update:value="handlePageSizeChange"
        />
      </div>
    </n-spin>

    <CreateImageModal
      v-model:show="showCreateModal"
      @success="fetchRepositories"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { imageApi } from '@/api'
import { useUserStore } from '@/stores/user'
import CreateImageModal from '@/components/CreateImageModal.vue'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const loading = ref(false)
const repositories = ref([])
const showCreateModal = ref(false)
const activeTab = ref('public')
const searchText = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const pageSizeOptions = [
  { label: '20 条/页', value: 20 },
  { label: '50 条/页', value: 50 },
  { label: '100 条/页', value: 100 }
]

const currentUserId = computed(() => userStore.userInfo?.id)

function getMoreOptions(repo) {
  return [
    // { 
    //   label: repo.is_public ? '设为私有' : '设为公开', 
    //   key: 'toggle_public' 
    // },
    { label: '删除', key: 'delete' }
  ]
}

function handleMoreAction(key, repo) {
  if (key === 'toggle_public') {
    togglePublic(repo)
  } else if (key === 'delete') {
    deleteRepo(repo.id)
  }
}

async function togglePublic(repo) {
  try {
    await imageApi.update(repo.id, { is_public: !repo.is_public })
    message.success(repo.is_public ? '已设为私有' : '已设为公开')
    fetchRepositories()
  } catch (err) {
    message.error('操作失败')
  }
}

async function fetchRepositories() {
  try {
    loading.value = true
    const skip = (page.value - 1) * pageSize.value
    const res = await imageApi.getList({
      skip,
      limit: pageSize.value,
      type: activeTab.value,
      search: searchText.value || undefined
    })
    repositories.value = res.repositories || []
    total.value = res.total || 0
  } catch (err) {
    console.error('Failed to fetch repositories:', err)
    message.error('加载镜像仓库失败')
  } finally {
    loading.value = false
  }
}

function handleTabChange() {
  page.value = 1
  fetchRepositories()
}

function handleSearch() {
  page.value = 1
  fetchRepositories()
}

function handlePageChange() {
  fetchRepositories()
}

function handlePageSizeChange() {
  page.value = 1
  fetchRepositories()
}

function formatTime(timestamp) {
  if (!timestamp) return 'N/A'
  return new Date(timestamp).toLocaleString()
}

function goToDetail(id) {
  router.push(`/admin/images/${id}`)
}

async function deleteRepo(id) {
  try {
    await imageApi.delete(id)
    message.success('删除成功')
    fetchRepositories()
  } catch (err) {
    message.error('删除失败')
  }
}

onMounted(() => {
  fetchRepositories()
})
</script>

<style scoped>
.image-repositories {
  padding-bottom: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0;
  border-bottom: 1px solid #e0e0e6;
  margin-bottom: 16px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 16px;
}

.page-main-title {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin: 0;
  white-space: nowrap;
  flex-shrink: 0;
  line-height: 1;
}

.title-divider {
  width: 1px;
  height: 16px;
  background: #e0e0e6;
  flex-shrink: 0;
}

.image-tabs {
  margin-bottom: 0;
}

.image-tabs :deep(.n-tabs-nav) {
  margin-bottom: 0;
}

.image-tabs :deep(.n-tabs-tab) {
  padding: 12px 0;
}

.page-title :deep(.n-tabs .n-tabs-tab--active) {
  font-weight: 600;
}

.page-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 8px;
}

.repo-name {
  color: #18a058;
  text-decoration: none;
  font-weight: 500;
}

.repo-name:hover {
  text-decoration: underline;
}

.creator-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.creator-avatar {
  background-color: #18a058;
  color: #fff;
  font-size: 12px;
}

.action-link {
  color: #18a058;
  text-decoration: none;
  font-size: 14px;
}

.action-link:hover {
  text-decoration: underline;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  margin-top: 16px;
}

.pagination span {
  color: #666;
  font-size: 14px;
}
</style>
