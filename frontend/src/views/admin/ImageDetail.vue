<template>
  <div class="image-detail">
    <div class="page-header">
      <div class="header-left">
        <n-button quaternary @click="goBack">
          <template #icon>
            <n-icon><ArrowBackOutline /></n-icon>
          </template>
        </n-button>
        <div>
          <h1>{{ repository?.name }}</h1>
          <div class="repo-meta">
            <span class="meta-item">
              <span class="meta-label">创建者:</span>
              <span class="meta-value">{{ repository?.user_info?.username || '-' }}</span>
            </span>
            <span class="meta-item">
              <span class="meta-label">创建时间:</span>
              <span class="meta-value">{{ formatTime(repository?.created_at) }}</span>
            </span>
            <span class="meta-item">
              <span class="meta-label">最后更新:</span>
              <span class="meta-value">{{ formatTime(repository?.updated_at) }}</span>
            </span>
            <span class="meta-item">
              <span class="meta-label">公开性:</span>
              <n-tag :type="repository?.is_public ? 'success' : 'warning'" size="small">
                {{ repository?.is_public ? '公开' : '私有' }}
              </n-tag>
            </span>
          </div>
        </div>
      </div>
      <div class="header-actions">
        <n-button type="primary" @click="showPushModal = true">推送镜像</n-button>
        <n-button v-if="isOwner" @click="showEditModal = true">设置</n-button>
      </div>
    </div>

    <n-card :bordered="false" class="tags-card">
      <n-spin :show="loadingTags">
        <n-table :bordered="false" :single-line="false" class="tags-table">
          <thead>
            <tr>
              <th>Tag</th>
              <th>大小</th>
              <th v-if="isOwner">操作</th>
            </tr>
          </thead>
          <tbody v-if="tags.length > 0">
            <tr v-for="tag in tags" :key="tag.name">
              <td>
                <div>{{ tag.name }}</div>
              </td>
              <td>{{ formatSize(tag.size) }}</td>
              <td v-if="isOwner">
                <n-button size="small" type="info" @click="showPullModal = true">
                  拉取
                </n-button>
                <n-button size="small" type="error" style="margin-left: 8px;" @click="deleteTag(tag.name)">
                  删除
                </n-button>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr class="table-empty-row">
              <td :colspan="isOwner ? 3 : 2">
                <n-empty description="暂无镜像Tag" />
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <EditImageModal
      v-model:show="showEditModal"
      :repository="repository"
      @success="fetchDetail"
    />

    <PushImageModal
      v-model:show="showPushModal"
      :repository="repository"
    />

    <PullImageModal
      v-model:show="showPullModal"
      :repository="repository"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { ArrowBackOutline } from '@vicons/ionicons5'
import { imageApi } from '@/api'
import { useUserStore } from '@/stores/user'
import EditImageModal from '@/components/EditImageModal.vue'
import PushImageModal from '@/components/PushImageModal.vue'
import PullImageModal from '@/components/PullImageModal.vue'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const userStore = useUserStore()

const loadingTags = ref(false)
const repository = ref(null)
const tags = ref([])
const showEditModal = ref(false)
const showPushModal = ref(false)
const showPullModal = ref(false)

const currentUserId = computed(() => userStore.userInfo?.id)

const isOwner = computed(() => {
  return repository.value?.user_id === currentUserId.value
})

async function fetchDetail() {
  try {
    const repoId = route.params.id
    const res = await imageApi.getDetail(repoId)
    repository.value = res
    tags.value = res.tags || []
  } catch (err) {
    console.error('Failed to fetch repository:', err)
    message.error('加载仓库详情失败')
  }
}

async function fetchTags() {
  try {
    loadingTags.value = true
    const repoId = route.params.id
    const res = await imageApi.getTags(repoId)
    tags.value = res.items || []
  } catch (err) {
    console.error('Failed to fetch tags:', err)
  } finally {
    loadingTags.value = false
  }
}

function formatSize(bytes) {
  if (!bytes || bytes === 0) return 'N/A'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(2)} ${units[unitIndex]}`
}

function formatTime(timestamp) {
  if (!timestamp) return 'N/A'
  return new Date(timestamp).toLocaleString()
}

function goBack() {
  router.push('/admin/images')
}

async function deleteTag(tagName) {
  try {
    const repoId = route.params.id
    await imageApi.deleteTag(repoId, tagName)
    message.success('删除成功')
    fetchTags()
  } catch (err) {
    message.error('删除失败')
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
.image-detail {
  padding-bottom: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h1 {
  font-size: 24px;
  font-weight: 600;
  color: var(--mixm-ink);
  margin-bottom: 4px;
}

.header-left p {
  color: var(--mixm-muted);
  font-size: 14px;
}

.repo-meta {
  margin-top: 10px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 18px;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.meta-label {
  color: var(--mixm-muted);
}

.meta-value {
  color: var(--mixm-ink);
}

.header-actions {
  display: flex;
  gap: 12px;
}

.tags-card :deep(.n-card__content) {
  padding-top: 8px;
}

.tags-table :deep(th) {
  padding-top: 12px;
  padding-bottom: 12px;
}

.tags-table :deep(td) {
  padding-top: 12px;
  padding-bottom: 12px;
}

.table-empty-row :deep(td) {
  padding: 40px 0;
}
</style>
