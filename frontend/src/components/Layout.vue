<template>
  <div class="layout-container" @keydown.esc="mobileSidebarOpen = false">
    <a class="skip-link" href="#main-content">跳转到主要内容</a>
    <header class="header">
      <div class="header-left">
        <button
          type="button"
          class="mobile-nav-toggle"
          :aria-expanded="mobileSidebarOpen"
          aria-controls="workspace-sidebar"
          :aria-label="mobileSidebarOpen ? '关闭导航' : '打开导航'"
          @click="toggleMobileSidebar"
        >
          <span></span><span></span>
        </button>
        <router-link :to="homePath" class="logo" aria-label="mixm 首页">
          <BrandMark />
          <span>mixm<span class="brand-period">.</span></span>
        </router-link>
        <div class="header-context">
          <span class="context-workspace">{{ workspaceLabel }}</span>
          <span class="context-divider" aria-hidden="true">/</span>
          <span class="context-page">{{ currentPageTitle }}</span>
        </div>
      </div>
      <div class="header-right">
        <n-badge :value="queueBadge" :max="99" :show="queueBadge > 0">
          <n-button
            quaternary
            circle
            @click="showQueueModal = true"
            title="创建队列"
            aria-label="查看实例创建队列"
          >
            <template #icon
              ><n-icon><ListOutline /></n-icon
            ></template>
          </n-button>
        </n-badge>
        <span class="header-rule" aria-hidden="true"></span>
        <n-dropdown :options="userMenuOptions" trigger="click" @select="handleUserMenu">
          <button type="button" class="user-info" aria-label="账户菜单">
            <span class="user-avatar" aria-hidden="true">{{ userInitial }}</span>
            <span class="username">{{ userStore.userInfo?.username || '用户' }}</span>
            <n-icon :size="13"><ChevronDownOutline /></n-icon>
          </button>
        </n-dropdown>
      </div>
    </header>

    <QueueStatusModal v-model:show="showQueueModal" />

    <div class="layout-body">
      <button
        v-if="mobileSidebarOpen"
        type="button"
        class="sidebar-backdrop"
        aria-label="关闭导航"
        @click="mobileSidebarOpen = false"
      ></button>
      <aside
        id="workspace-sidebar"
        :class="['sidebar', { collapsed: isSidebarCollapsed, 'mobile-open': mobileSidebarOpen }]"
        aria-label="工作空间导航"
      >
        <div v-if="!isSidebarCollapsed" class="sidebar-intro">
          <p class="eyebrow">YOUR WORKSPACE</p>
          <span>知识，始于探索。</span>
        </div>
        <nav class="sidebar-menu" aria-label="主导航">
          <n-menu
            :value="activeMenu"
            :options="menuOptions"
            :collapsed="isSidebarCollapsed"
            :collapsed-width="64"
            :collapsed-icon-size="20"
            :default-expand-all="true"
            @update:value="handleMenuSelect"
          />
        </nav>
        <div class="sidebar-footer">
          <span v-if="!isSidebarCollapsed" class="sidebar-footer-copy">保持好奇，持续生长。</span>
          <n-button
            quaternary
            circle
            @click="toggleSidebar"
            :aria-label="isSidebarCollapsed ? '展开侧栏' : '收起侧栏'"
          >
            <template #icon
              ><n-icon
                ><ChevronBackOutline v-if="!isSidebarCollapsed" /><ChevronForwardOutline
                  v-else /></n-icon
            ></template>
          </n-button>
        </div>
      </aside>
      <main id="main-content" class="main-content" tabindex="-1"><router-view /></main>
    </div>
  </div>
</template>

<script setup>
import { computed, h, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NBadge, NIcon } from 'naive-ui'
import {
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
  SpeedometerOutline
} from '@vicons/ionicons5'
import { useUserStore } from '@/stores/user'
import { resetRouter } from '@/router'
import QueueStatusModal from './QueueStatusModal.vue'
import BrandMark from './BrandMark.vue'
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
  SearchOutline
}

const defaultMenuItems = [
  { key: 'ClientHome', label: '首页', path: '/app', icon: 'HomeOutline' },
  { key: 'ClientContents', label: '内容浏览', path: '/app/contents', icon: 'BookOutline' },
  { key: 'ClientAssignments', label: '我的作业', path: '/app/assignments', icon: 'CreateOutline' },
  {
    key: 'ClientSimulationTest',
    label: '模拟测试',
    path: '/app/simulation-test',
    icon: 'DocumentTextOutline'
  },
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
  { key: 'Practices', label: '练习管理', path: '/admin/practices', icon: 'CreateOutline' }
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
const mobileSidebarOpen = ref(false)
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
      children: instanceItems
    })
  }

  if (practiceItems.length > 0) {
    result.push({
      key: 'AdminPracticeGroup',
      label: '练习管理',
      icon: 'CreateOutline',
      hideChildrenIcon: true,
      children: practiceItems
    })
  }

  return result
}

function normalizeMenuItems(items) {
  const source = Array.isArray(items) ? items : []
  const hasAdminItems = source.some((item) => item?.path?.startsWith?.('/admin'))
  const hasNestedChildren = source.some(
    (item) => Array.isArray(item?.children) && item.children.length > 0
  )

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
      : undefined
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
  if (['ClientMyInstances', 'ClientInstanceDetail'].includes(currentRouteName))
    return 'ClientMyInstances'
  if (['ClientAssignments'].includes(currentRouteName)) return 'ClientAssignments'
  if (['ClientSimulationTest'].includes(currentRouteName)) return 'ClientSimulationTest'
  const foundNested = findMenuItemByKey(menuItems.value, currentRouteName)
  if (foundNested) return foundNested.key

  const found = menuItems.value.find((item) => item.key === currentRouteName)
  if (found) return found.key
  return 'ClientHome'
})

const workspaceLabel = computed(() => (route.path.startsWith('/admin') ? '管理工作台' : '学习空间'))
const currentPageTitle = computed(
  () =>
    findMenuItemByKey(menuItems.value, activeMenu.value)?.label ||
    route.meta.title ||
    workspaceLabel.value
)
const homePath = computed(() => (userStore.userInfo?.role === 'admin' ? '/admin' : '/app'))
const userInitial = computed(() => (userStore.userInfo?.username || 'M').slice(0, 1).toUpperCase())

watch(
  () => route.fullPath,
  () => {
    mobileSidebarOpen.value = false
  }
)

function toggleMobileSidebar() {
  isSidebarCollapsed.value = false
  mobileSidebarOpen.value = !mobileSidebarOpen.value
}

const userMenuOptions = [{ label: '退出登录', key: 'logout' }]

async function handleUserMenu(key) {
  if (key === 'logout') {
    userStore.logout()
    resetRouter()
    await router.replace('/login')
  }
}

function handleMenuSelect(key) {
  mobileSidebarOpen.value = false
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
  height: 100svh;
  overflow: hidden;
  background: var(--mixm-paper);
}
.skip-link {
  position: fixed;
  top: -80px;
  left: 24px;
  z-index: 300;
  padding: 12px 18px;
  background: var(--mixm-ink);
  color: var(--mixm-surface);
  border-radius: 4px;
}
.skip-link:focus {
  top: 12px;
}
.header {
  height: 76px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 0 32px;
  background: var(--mixm-paper);
  border-bottom: 1px solid var(--mixm-border);
  z-index: 100;
}
.header-left,
.header-right {
  display: flex;
  align-items: center;
  min-width: 0;
}
.header-left {
  gap: 26px;
}
.header-right {
  gap: 16px;
  flex-shrink: 0;
}
.logo {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  min-width: 173px;
}
.logo > span {
  color: var(--mixm-ink);
  font-size: 31px;
  font-weight: 600;
  letter-spacing: -0.065em;
  line-height: 1;
}
.brand-period {
  color: var(--mixm-accent);
}
.header-context {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
  font-size: 12px;
}
.context-workspace {
  color: var(--mixm-muted);
  white-space: nowrap;
}
.context-divider {
  color: #b7b2a4;
  font-size: 16px;
}
.context-page {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.header-rule {
  height: 20px;
  width: 1px;
  background: var(--mixm-border);
}
.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 0;
  border: 0;
  background: transparent;
  color: var(--mixm-ink);
  cursor: pointer;
  font-size: 12px;
}
.user-avatar {
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border: 1px solid var(--mixm-border);
  border-radius: 50%;
  background: var(--mixm-surface-muted);
  font-family: var(--mixm-font-display);
  font-size: 15px;
}
.layout-body {
  position: relative;
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}
.sidebar {
  width: 232px;
  flex-shrink: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--mixm-border);
  background: var(--mixm-paper);
  transition: width 0.2s ease;
}
.sidebar.collapsed {
  width: 64px;
}
.sidebar-intro {
  padding: 25px 24px 16px;
}
.sidebar-intro .eyebrow {
  font-size: 9px;
  letter-spacing: 0.14em;
  margin-bottom: 8px;
  color: var(--mixm-faint);
}
.sidebar-intro > span {
  font-family: var(--mixm-font-display);
  font-size: 17px;
}
.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  padding: 0 8px 16px;
}
.collapsed .sidebar-menu {
  padding: 16px 0;
}
.sidebar-menu :deep(.n-menu-item-content) {
  font-size: 13px;
}
.sidebar-footer {
  min-height: 68px;
  padding: 12px 16px;
  border-top: 1px solid var(--mixm-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.collapsed .sidebar-footer {
  padding-inline: 0;
  justify-content: center;
}
.sidebar-footer-copy {
  font-size: 10px;
  color: var(--mixm-muted);
}
.main-content {
  flex: 1;
  min-width: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 36px 40px;
  overflow-y: auto;
  overflow-x: hidden;
}
.main-content:has(.fullscreen-page) {
  max-width: 100%;
  padding: 0;
}
.main-content:has(.fullscreen-page) > * {
  max-width: none;
}
.mobile-nav-toggle,
.sidebar-backdrop {
  display: none;
}
@media (max-width: 1100px) {
  .main-content {
    padding: 28px;
  }
  .header {
    padding: 0 24px;
  }
  .logo {
    min-width: 181px;
  }
}
@media (max-width: 760px) {
  .header {
    height: 68px;
    padding: 0 18px;
    gap: 16px;
  }
  .header-left {
    gap: 12px;
  }
  .logo {
    min-width: 0;
    gap: 6px;
  }
  .logo > span {
    font-size: 28px;
  }
  .logo :deep(.brand-mark) {
    width: 28px;
    height: 28px;
  }
  .header-context,
  .header-rule,
  .username {
    display: none;
  }
  .header-right {
    gap: 12px;
  }
  .mobile-nav-toggle {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 6px;
    width: 28px;
    height: 36px;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .mobile-nav-toggle span {
    width: 18px;
    height: 1px;
    background: var(--mixm-ink);
  }
  .sidebar {
    position: absolute;
    inset: 0 auto 0 0;
    width: 248px;
    transform: translateX(-100%);
    visibility: hidden;
    z-index: 30;
    transition:
      transform 0.2s ease,
      visibility 0.2s;
  }
  .sidebar.mobile-open {
    transform: translateX(0);
    visibility: visible;
  }
  .sidebar-backdrop {
    display: block;
    position: absolute;
    inset: 0;
    z-index: 20;
    border: 0;
    background: rgb(41 41 35 / 25%);
  }
  .sidebar-footer :deep(.n-button) {
    display: none;
  }
  .main-content {
    padding: 24px 18px;
  }
}
</style>
