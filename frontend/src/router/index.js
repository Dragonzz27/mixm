import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/register',
    redirect: '/login'
  },
  {
    path: '/app',
    component: () => import('@/components/Layout.vue'),
    meta: { requiresAuth: true, isClient: true },
    children: [
      {
        path: '',
        name: 'ClientHome',
        meta: { title: 'Home' },
        component: () => import('@/views/client/Home.vue')
      },
      {
        path: 'contents',
        name: 'ClientContents',
        meta: { title: '内容列表' },
        component: () => import('@/views/client/Contents.vue')
      },
      {
        path: 'levels',
        name: 'ClientLevels',
        meta: { title: 'Levels' },
        component: () => import('@/views/client/Home.vue')
      },
      {
        path: 'levels/:id',
        name: 'ClientLevelDetail',
        meta: { title: 'Level Detail' },
        component: () => import('@/views/client/Home.vue')
      },
      {
        path: 'my-instances',
        name: 'ClientMyInstances',
        meta: { title: '我启动的实例' },
        component: () => import('@/views/client/MyInstances.vue')
      },
      {
        path: 'instances/:id',
        name: 'ClientInstanceDetail',
        meta: { title: 'Instance Detail' },
        component: () => import('@/views/client/InstanceDetail.vue')
      },
      {
        path: 'content/:id',
        name: 'ClientContent',
        meta: { title: '内容详情' },
        component: () => import('@/views/client/ContentDetail.vue')
      },
      {
        path: 'images',
        name: 'ClientImageRepositories',
        meta: { title: '镜像仓库' },
        component: () => import('@/views/client/ImageRepositories.vue')
      },
      {
        path: 'images/:id',
        name: 'ClientImageDetail',
        meta: { title: '镜像详情' },
        component: () => import('@/views/client/ImageDetail.vue')
      }
    ]
  },
  {
    path: '/admin',
    component: () => import('@/components/Layout.vue'),
    meta: { requiresAuth: true, isAdmin: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/admin/Dashboard.vue')
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/admin/Users.vue')
      },
      {
        path: 'categories',
        name: 'Categories',
        component: () => import('@/views/admin/Categories.vue')
      },
      {
        path: 'contents',
        name: 'Contents',
        component: () => import('@/views/admin/Contents.vue')
      },
      {
        path: 'templates',
        name: 'Templates',
        component: () => import('@/views/admin/Templates.vue')
      },
      {
        path: 'bindings',
        name: 'Bindings',
        component: () => import('@/views/admin/Bindings.vue')
      },
      {
        path: 'clusters',
        name: 'Clusters',
        component: () => import('@/views/admin/Clusters.vue')
      },
      {
        path: 'instances',
        name: 'Instances',
        component: () => import('@/views/admin/Instances.vue')
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/app'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth && !userStore.token) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.meta.guest && userStore.token) {
    const role = userStore.userInfo?.role
    if (role === 'admin') {
      next({ path: '/admin' })
    } else {
      next({ path: '/app' })
    }
  } else {
    next()
  }
})

export default router
