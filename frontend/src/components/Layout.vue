<template>
  <div class="layout-container">
    <header class="header">
      <div class="header-left">
        <a href="https://mixm.yoobit.cn" class="logo" target="_blank" rel="noopener noreferrer">
          <n-icon :size="24" :component="UnlinkSharp" />
          <span>mixm</span>
        </a>
      </div>
      <div class="header-right">
        <n-badge :value="queueBadge" :max="99" :show="queueBadge > 0">
          <n-button quaternary circle @click="showQueueModal = true" title="鍒涘缓闃熷垪">
            <template #icon>
              <n-icon><ListOutline /></n-icon>
            </template>
          </n-button>
        </n-badge>
        <n-dropdown :options="userMenuOptions" @select="handleUserMenu">
          <div class="user-info">
            <span>{{ userStore.userInfo?.username || '用户' }}</span>
            <n-icon><ChevronDownOutline /></n-icon>
          </div>
        </n-dropdown>
      </div>
    </header>

    <QueueStatusModal v-model:show="showQueueModal" />

    <div class="layout-body">
      <aside :class="['sidebar', { collapsed: isSidebarCollapsed }]">
        <div class="sidebar-menu">
          <n-menu
            :value="activeMenu"
            :options="menuOptions"
            :collapsed="isSidebarCollapsed"
            :collapsed-width="64"
            :collapsed-icon-size="22"
            :default-expand-all="true"
            @update:value="handleMenuSelect"
          />
        </div>
        <div class="sidebar-footer">
          <n-button quaternary circle @click="toggleSidebar">
            <template #icon>
              <n-icon>
                <ChevronBackOutline v-if="!isSidebarCollapsed" />
                <ChevronForwardOutline v-else />
              </n-icon>
            </template>
          </n-button>
        </div>
      </aside>

      <main class="main-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed, h, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NBadge, NIcon } from 'naive-ui'
import {
  UnlinkSharp,
  ChevronDownOutline,
  ChevronBackOutline,
  ChevronForwardOutline,
  HomeOutline,
  BookOutline,
  CreateOutline,
  CubeOutline,
  DocumentTextOutline,
  FileTrayFullOutline,
  GridOutline,
  ImageOutline,
  LayersOutline,
  ListOutline,
  PeopleOutline,
  SearchOutline,
  ServerOutline,
  ShieldCheckmarkOutline,
  SpeedometerOutline,
} from '@vicons/ionicons5'
import { useUserStore } from '@/stores/user'
import { resetRouter } from '@/router'
import QueueStatusModal from './QueueStatusModal.vue'
import { instanceApi } from '@/api'

const iconMap = {
  HomeOutline,
  BookOutline,
  CubeOutline,
  SpeedometerOutline,
  PeopleOutline,
  ShieldCheckmarkOutline,
  GridOutline,
  DocumentTextOutline,
  FileTrayFullOutline,
  ServerOutline,
  LayersOutline,
  ImageOutline,
  CreateOutline,
  SearchOutline,
}

const defaultMenuItems = [
  { key: 'ClientHome', label: '首页', path: '/app', icon: 'HomeOutline' },
  { key: 'ClientContents', label: '内容浏览', path: '/app/contents', icon: 'BookOutline' },
  { key: 'ClientAssignments', label: '我的作业', path: '/app/assignments', icon: 'CreateOutline' },
  { key: 'ClientSimulationTest', label: '模拟测试', path: '/app/simulation-test', icon: 'DocumentTextOutline' },
  { key: 'ClientMyInstances', label: '我的实例', path: '/app/my-instances', icon: 'CubeOutline' },
  { type: 'divider', key: 'd1' },
  { key: 'Dashboard', label: '仪表盘', path: '/admin', icon: 'SpeedometerOutline' },
  { key: 'Users', label: '用户管理', path: '/admin/users', icon: 'PeopleOutline' },
  { key: 'Roles', label: '角色管理', path: '/admin/roles', icon: 'ShieldCheckmarkOutline' },
  { key: 'Categories', label: '分类管理', path: '/admin/categories', icon: 'GridOutline' },
  { key: 'Contents', label: '内容管理', path: '/admin/contents', icon: 'DocumentTextOutline' },
  { key: 'Templates', label: '模板管理', path: '/admin/templates', icon: 'FileTrayFullOutline' },
  { key: 'Instances', label: '实例管理', path: '/admin/instances', icon: 'ServerOutline' },
  { key: 'AdminImageRepositories', label: '镜像仓库', path: '/admin/images', icon: 'ImageOutline' },
  { key: 'Clusters', label: '集群管理', path: '/admin/clusters', icon: 'LayersOutline' },
  { key: 'Assignments', label: '作业管理', path: '/admin/assignments', icon: 'CreateOutline' },
  { key: 'Questions', label: '题库管理', path: '/admin/questions', icon: 'DocumentTextOutline' },
  { key: 'Practices', label: '练习管理', path: '/admin/practices', icon: 'CreateOutline' },
]

const requiredClientKeys = new Set([
  'ClientHome',
  'ClientContents',
  'ClientAssignments',
  'ClientSimulationTest',
  'ClientMyInstances'
])
const instanceMenuKeys = new Set(['Templates', 'Instances', 'AdminImageRepositories', 'Clusters'])
const practiceMenuKeys = new Set(['Assignments', 'Questions', 'Practices', 'ClientSimulationTest'])

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isSidebarCollapsed = ref(false)
const showQueueModal = ref(false)
const queueBadge = ref(0)
let queueBadgeTimer = null

function buildGroupedAdminMenu(items) {
  const result = []
  const instanceItems = []
  const practiceItems = []

  for (const item of items) {
    if (item?.type === 'divider') {
      result.push(item)
      continue
    }

    if (instanceMenuKeys.has(item.key)) {
      instanceItems.push(item)
      continue
    }

    if (practiceMenuKeys.has(item.key)) {
      practiceItems.push(item)
      continue
    }

    result.push(item)
  }

  if (instanceItems.length > 0) {
    result.push({
      key: 'AdminInstanceGroup',
      label: '实例管理',
      icon: 'ServerOutline',
      hideChildrenIcon: true,
      children: instanceItems,
    })
  }

  if (practiceItems.length > 0) {
    result.push({
      key: 'AdminPracticeGroup',
      label: '练习管理',
      icon: 'CreateOutline',
      hideChildrenIcon: true,
      children: practiceItems,
    })
  }

  return result
}

function normalizeMenuItems(items) {
  const source = Array.isArray(items) ? items : []
  const hasAdminItems = source.some((item) => item?.path?.startsWith?.('/admin'))
  const hasNestedChildren = source.some((item) => Array.isArray(item?.children) && item.children.length > 0)

  if (!hasAdminItems || hasNestedChildren) {
    return source
  }

  return buildGroupedAdminMenu(source)
}

const groupedParentKeys = new Set(['AdminInstanceGroup', 'AdminPracticeGroup'])

function renderMenuOption(item, parentKey = null) {
  if (item.type === 'divider') return item

  const shouldHideIcon = groupedParentKeys.has(parentKey)
  const iconComponent = shouldHideIcon ? null : iconMap[item.icon]
  return {
    label: item.label,
    key: item.key,
    icon: iconComponent ? () => h(NIcon, null, { default: () => h(iconComponent) }) : undefined,
    children: Array.isArray(item.children)
      ? item.children.map((child) => renderMenuOption(child, item.key))
      : undefined,
  }
}

function findMenuItemByKey(items, key) {
  for (const item of items || []) {
    if (item?.key === key) {
      return item
    }
    if (Array.isArray(item?.children)) {
      const found = findMenuItemByKey(item.children, key)
      if (found) return found
    }
  }
  return null
}

async function updateQueueBadge() {
  try {
    const res = await instanceApi.getQueueStatus()
    queueBadge.value = (res?.pending || 0) + (res?.processing || 0)
  } catch (err) {
    console.error('Failed to get queue status:', err)
  }
}

onMounted(() => {
  updateQueueBadge()
  queueBadgeTimer = setInterval(updateQueueBadge, 30000)
  window.addEventListener('queue-updated', updateQueueBadge)
})

onUnmounted(() => {
  if (queueBadgeTimer) {
    clearInterval(queueBadgeTimer)
    queueBadgeTimer = null
  }
  window.removeEventListener('queue-updated', updateQueueBadge)
})

const menuItems = computed(() => {
  if (Array.isArray(userStore.menuConfig) && userStore.menuConfig.length > 0) {
    const fromStore = [...userStore.menuConfig]
    const existingKeys = new Set(fromStore.map((item) => item?.key).filter(Boolean))
    defaultMenuItems.forEach((item) => {
      if (requiredClientKeys.has(item.key) && !existingKeys.has(item.key)) {
        fromStore.push(item)
      }
    })
    return normalizeMenuItems(fromStore)
  }
  return normalizeMenuItems(defaultMenuItems)
})

const menuOptions = computed(() => menuItems.value.map(renderMenuOption))

const activeMenu = computed(() => {
  const currentRouteName = route.name
  if (['ClientContents', 'ClientContent'].includes(currentRouteName)) return 'ClientContents'
  if (['ClientMyInstances', 'ClientInstanceDetail'].includes(currentRouteName)) return 'ClientMyInstances'
  if (['ClientAssignments'].includes(currentRouteName)) return 'ClientAssignments'
  if (['ClientSimulationTest'].includes(currentRouteName)) return 'ClientSimulationTest'
  const foundNested = findMenuItemByKey(menuItems.value, currentRouteName)
  if (foundNested) return foundNested.key

  const found = menuItems.value.find((item) => item.key === currentRouteName)
  if (found) return found.key
  return 'ClientHome'
})

const userMenuOptions = [{ label: '退出登录', key: 'logout' }]

async function handleUserMenu(key) {
  if (key === 'logout') {
    userStore.logout()
    resetRouter()
    await router.replace('/login')
  }
}

function handleMenuSelect(key) {
  const target = findMenuItemByKey(menuItems.value, key)
  if (target?.path && route.path !== target.path) {
    router.push(target.path)
  }
}

function toggleSidebar() {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}
</script>

<style scoped>
.layout-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background-color: #ffffff;
}

.header {
  height: 64px;
  flex-shrink: 0;
  background: #ffffff;
  border-bottom: 1px solid #e0e0e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 32px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 20px;
  font-weight: 600;
  color: #000;
  text-decoration: none;
}

.logo:hover {
  color: #36ad6a;
}

.layout-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: 220px;
  height: 100%;
  border-right: 1px solid #e0e0e6;
  background: #fff;
  display: flex;
  flex-direction: column;
  transition: width 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar.collapsed {
  width: 64px;
}

.sidebar-menu {
  flex: 1;
  padding: 16px 0;
  overflow-y: auto;
}

.sidebar-footer {
  padding: 12px 0;
  border-top: 1px solid #f0f0f5;
  display: flex;
  justify-content: center;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-info:hover {
  background: #f5f5f5;
}

.user-info span {
  font-size: 14px;
  color: #333;
}

.main-content {
  flex: 1;
  min-width: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 24px;
  overflow-y: auto;
}

.main-content:has(.fullscreen-page) {
  max-width: 100%;
  padding: 0;
}
</style>
