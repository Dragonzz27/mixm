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
        meta: { title: '我的实例' },
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
        path: 'assignments',
        name: 'ClientAssignments',
        meta: { title: '我的作业' },
        component: () => import('@/views/client/Assignments.vue')
      },
      {
        path: 'simulation-test',
        name: 'ClientSimulationTest',
        meta: { title: '模拟测试' },
        component: () => import('@/views/client/SimulationTest.vue')
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
        path: 'roles',
        name: 'Roles',
        component: () => import('@/views/admin/Roles.vue')
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
      },
      {
        path: 'images',
        name: 'AdminImageRepositories',
        meta: { title: '镜像仓库' },
        component: () => import('@/views/admin/ImageRepositories.vue')
      },
      {
        path: 'images/:id',
        name: 'AdminImageDetail',
        meta: { title: '镜像详情' },
        component: () => import('@/views/admin/ImageDetail.vue')
      },
      {
        path: 'assignments',
        name: 'Assignments',
        meta: { title: '作业管理' },
        component: () => import('@/views/admin/Assignments.vue')
      },
      {
        path: 'questions',
        name: 'Questions',
        meta: { title: '题库管理' },
        component: () => import('@/views/admin/Questions.vue')
      },
      {
        path: 'practices',
        name: 'Practices',
        meta: { title: '练习管理' },
        component: () => import('@/views/admin/Practices.vue')
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/app'
  }
]

function collectRouteNames(routeTree) {
  const nameSet = new Set()
  const walk = (items) => {
    if (!Array.isArray(items)) return
    items.forEach((item) => {
      if (item?.name) {
        nameSet.add(item.name)
      }
      if (Array.isArray(item?.children) && item.children.length > 0) {
        walk(item.children)
      }
    })
  }
  walk(routeTree)
  return nameSet
}

const staticRouteNames = collectRouteNames(routes)

function collectAllowedNames(routeTree) {
  const nameSet = new Set()
  const walk = (items) => {
    if (!Array.isArray(items)) return
    items.forEach((item) => {
      if (item?.name) {
        nameSet.add(item.name)
      }
      if (Array.isArray(item?.children) && item.children.length > 0) {
        walk(item.children)
      }
    })
  }
  walk(routeTree)
  return nameSet
}

const router = createRouter({
  history: createWebHistory(),
  routes
})

export function resetRouter() {
  router.getRoutes().forEach((routeRecord) => {
    const routeName = routeRecord.name
    if (!routeName || staticRouteNames.has(routeName)) {
      return
    }
    if (router.hasRoute(routeName)) {
      router.removeRoute(routeName)
    }
  })
}

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const allowedNames = collectAllowedNames(userStore.routeConfig)

  if (to.meta.requiresAuth && !userStore.token) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.meta.requiresAuth && allowedNames.size > 0 && to.name && !allowedNames.has(to.name)) {
    const fallback = userStore.menuConfig?.find?.((item) => item?.path)?.path || '/app'
    next({ path: fallback })
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
