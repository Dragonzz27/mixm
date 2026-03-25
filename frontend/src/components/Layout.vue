<template>
  <div class="layout-container">
    <header class="header">
      <div class="header-left">
        <router-link to="/app" class="logo">
          <n-icon :size="24" :component="UnlinkSharp" />
          <span>MutiM</span>
        </router-link>
      </div>
      <div class="header-right">
        <n-input placeholder="Search content..." style="width: 240px;">
          <template #prefix>
            <n-icon><SearchOutline /></n-icon>
          </template>
        </n-input>
        <n-badge :value="queueBadge" :max="99" :show="queueBadge > 0">
          <n-button quaternary circle @click="showQueueModal = true" title="创建队列">
            <template #icon>
              <n-icon><ListOutline /></n-icon>
            </template>
          </n-button>
        </n-badge>
        <n-dropdown :options="userMenuOptions" @select="handleUserMenu">
          <div class="user-info">
            <!-- <n-avatar round size="small" src="/300.img" /> -->
            <span>{{ userStore.userInfo?.username || 'User' }}</span>
            <n-icon><ChevronDownOutline /></n-icon>
          </div>
        </n-dropdown>
        <!-- <n-button type="primary" size="small" @click="goToAdmin">Admin</n-button> -->
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
import { h, ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { NIcon, NBadge } from 'naive-ui'
import {
  UnlinkSharp,
  SearchOutline,
  ChevronDownOutline,
  ChevronBackOutline,
  ChevronForwardOutline,
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
  ListOutline,
  ImageOutline
} from '@vicons/ionicons5'
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
  ImageOutline
}

const defaultMenuItems = [
  { key: 'ClientHome', label: '首页', path: '/app', icon: 'HomeOutline' },
  { key: 'ClientContents', label: '内容浏览', path: '/app/contents', icon: 'BookOutline' },
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
  { key: 'Clusters', label: '集群管理', path: '/admin/clusters', icon: 'LayersOutline' }
]

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const isSidebarCollapsed = ref(false)
const showQueueModal = ref(false)
const queueBadge = ref(0)
let queueBadgeTimer = null

async function updateQueueBadge() {
  try {
    const res = await instanceApi.getQueueStatus()
    queueBadge.value = (res.pending || 0) + (res.processing || 0)
  } catch (err) {
    console.error('Failed to get queue status:', err)
  }
}

onMounted(() => {
  updateQueueBadge()
  queueBadgeTimer = setInterval(updateQueueBadge, 5000)
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
    return userStore.menuConfig
  }
  return defaultMenuItems
})

const menuOptions = computed(() => {
  return menuItems.value.map((item) => {
    if (item.type === 'divider') return item
    const iconComponent = iconMap[item.icon]
    return {
      label: item.label,
      key: item.key,
      icon: iconComponent ? () => h(NIcon, null, { default: () => h(iconComponent) }) : undefined
    }
  })
})

const activeMenu = computed(() => {
  const currentRouteName = route.name
  // 匹配子级路由到父级菜单项
  if (['ClientContents', 'ClientContent'].includes(currentRouteName)) {
    return 'ClientContents'
  }
  if (['ClientMyInstances', 'ClientInstanceDetail'].includes(currentRouteName)) {
    return 'ClientMyInstances'
  }
  
  // 管理端路由匹配
  const found = menuItems.value.find(item => item.key === currentRouteName)
  if (found) return found.key

  return 'ClientHome'
})

const userMenuOptions = [
  { label: 'Profile', key: 'profile' },
  { label: 'Settings', key: 'settings' },
  { type: 'divider', key: 'd1' },
  { label: 'Logout', key: 'logout' }
]

function handleUserMenu(key) {
  switch (key) {
    case 'logout':
      userStore.logout()
      router.push('/login')
      break
    case 'profile':
      router.push('/app/profile')
      break
    case 'settings':
      router.push('/app/settings')
      break
  }
}

function handleMenuSelect(key) {
  const target = menuItems.value.find((item) => item.key === key)
  if (target && route.path !== target.path) {
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
  overflow: hidden; /* 切断边缘溢出，保持画幅纯粹 */
  background-color: #ffffff;
}

.header {
  height: 64px;
  flex-shrink: 0; /* 关键：拒绝被压缩 */
  background: white;
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

/* 3. 画面的下半部分（主体承载区） */
/* 占据 Header 剩下的所有空间，同时锁死不让内容撑破这个区域 */
.layout-body {
  flex: 1;
  display: flex;
  overflow: hidden; /* 关键：这就像遮罩层，确保 Sidebar 和 Main 都在此高度内 */
}

/* 4. 左侧前景框架（Sidebar） */
/* 它的高度会自动撑满 layout-body，达到完美的 100% 占比 */
.sidebar {
  width: 220px;
  height: 100%; /* 严丝合缝地填满高度 */
  border-right: 1px solid #e0e0e6;
  background: #fff;
  display: flex;
  flex-direction: column;
  transition: width 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  /* overflow: hidden; 这里移除，交给内部的 sidebar-menu 去处理滚动 */
}

.sidebar.collapsed {
  width: 64px;
}

/* 如果菜单项极多，允许左侧这部分区域自己拥有微小的内部滚动，不破坏框架 */
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

/* 5. 核心流动主体（Main Content） */
/* 让右侧内容像胶卷一样，在固定的画框中自由上下滑动 */
.main-content {
  flex: 1;
  min-width: 0;
  height: 100%; /* 高度强制约束在父级框架内 */
  display: flex;
  flex-direction: column;
  padding: 24px;
  overflow-y: auto; /* 关键：开启右侧独立的垂直滚动条！ */
}

.main-content:has(.fullscreen-page) {
  max-width: 100%;
  padding: 0;
}
</style>