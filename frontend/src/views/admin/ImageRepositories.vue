<template>
  <div class="image-repositories">
    <div class="page-header">
      <div class="page-title">
        <div class="page-main-title">镜像仓库</div>
        <span class="sync-note">自动同步 Registry · 每 30 秒更新</span>
      </div>
      <div class="page-actions">
        <n-input
          v-model:value="searchText"
          placeholder="搜索镜像"
          clearable
          style="width: 200px"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <n-button :loading="syncing" @click="syncRepositories">刷新</n-button>
      </div>
    </div>

    <n-alert v-if="syncError" type="warning" title="Registry 同步失败" class="sync-alert">
      {{ syncError }}。下方显示已登记的仓库，可点击刷新重试。
    </n-alert>

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
              <td>{{ repo.tags_count ?? '—' }}</td>
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
                  <a href="javascript:void(0)" class="action-link" @click="goToDetail(repo.id)">
                    详情
                  </a>
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
      <n-empty
        v-else
        :description="searchText ? '没有匹配的镜像仓库' : 'Registry 中暂无镜像仓库，请先推送镜像'"
        style="padding: 80px 0"
      />

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
          style="width: 90px"
          @update:value="handlePageSizeChange"
        />
      </div>
    </n-spin>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { imageApi } from '@/api'

const router = useRouter()
const message = useMessage()

const loading = ref(true)
const syncing = ref(false)
const syncError = ref('')
let syncTimer = null
const repositories = ref([])
const searchText = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const pageSizeOptions = [
  { label: '20 条/页', value: 20 },
  { label: '50 条/页', value: 50 },
  { label: '100 条/页', value: 100 }
]

function getMoreOptions() {
  return [{ label: '删除', key: 'delete' }]
}

function handleMoreAction(key, repo) {
  if (key === 'delete') deleteRepo(repo.id)
}

// 打开页面和定时刷新时先发现新仓库，再查询当前页；同步失败仍可查看已有记录。
async function syncRepositories() {
  if (syncing.value) return
  syncing.value = true
  try {
    await imageApi.sync()
    syncError.value = ''
  } catch (err) {
    syncError.value = err.response?.data?.message || '无法连接镜像仓库'
  } finally {
    await fetchRepositories()
    syncing.value = false
  }
}

async function fetchRepositories() {
  try {
    loading.value = true
    const offset = (page.value - 1) * pageSize.value
    const res = await imageApi.getList({
      offset,
      limit: pageSize.value,
      search: searchText.value || undefined
    })
    repositories.value = res.items || []
    total.value = res.total || 0
  } catch (err) {
    console.error('Failed to fetch repositories:', err)
    message.error('加载镜像仓库失败')
  } finally {
    loading.value = false
  }
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
  syncRepositories()
  syncTimer = setInterval(syncRepositories, 30000)
})

onUnmounted(() => {
  clearInterval(syncTimer)
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
  border-bottom: 1px solid var(--mixm-border);
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
  color: var(--mixm-ink);
  margin: 0;
  white-space: nowrap;
  flex-shrink: 0;
  line-height: 1;
}

.sync-note {
  color: var(--mixm-muted);
  font-size: 12px;
}

.sync-alert {
  margin-bottom: 16px;
}

.page-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 8px;
}

.repo-name {
  color: var(--mixm-accent);
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
  background-color: var(--mixm-accent);
  color: #fff;
  font-size: 12px;
}

.action-link {
  color: var(--mixm-accent);
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
  color: var(--mixm-muted);
  font-size: 14px;
}
@media (max-width: 1000px) {
  .page-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 16px;
  }
  .page-title {
    flex-wrap: wrap;
    gap: 8px 16px;
  }
  .page-actions {
    flex-wrap: wrap;
    width: 100%;
  }
}
</style>
