import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

export const useUserStore = defineStore('user', () => {
  const getStoredItem = (key) => {
    const val = localStorage.getItem(key)
    return (val === 'undefined' || val === 'null') ? '' : (val || '')
  }

  const token = ref(getStoredItem('token'))
  const refreshToken = ref(getStoredItem('refreshToken'))
  const userInfo = ref(null)
  const routeConfig = ref([])
  const menuConfig = ref([])
  try {
    const storedUserInfo = localStorage.getItem('userInfo')
    if (storedUserInfo && storedUserInfo !== 'undefined') {
      userInfo.value = JSON.parse(storedUserInfo)
    }
  } catch (err) {
    console.error('Failed to parse userInfo from localStorage:', err)
  }

  try {
    const storedRouteConfig = localStorage.getItem('routeConfig')
    if (storedRouteConfig && storedRouteConfig !== 'undefined') {
      routeConfig.value = JSON.parse(storedRouteConfig)
    }
  } catch (err) {
    console.error('Failed to parse routeConfig from localStorage:', err)
  }

  try {
    const storedMenuConfig = localStorage.getItem('menuConfig')
    if (storedMenuConfig && storedMenuConfig !== 'undefined') {
      menuConfig.value = JSON.parse(storedMenuConfig)
    }
  } catch (err) {
    console.error('Failed to parse menuConfig from localStorage:', err)
  }
  const refreshTokenPending = ref(false)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(value) {
    token.value = value || ''
    if (value) {
      localStorage.setItem('token', value)
    } else {
      localStorage.removeItem('token')
    }
  }

  function setRefreshToken(value) {
    refreshToken.value = value || ''
    if (value) {
      localStorage.setItem('refreshToken', value)
    } else {
      localStorage.removeItem('refreshToken')
    }
  }

  async function refreshAccessToken() {
    if (refreshTokenPending.value) {
      return false
    }
    if (!refreshToken.value) {
      return false
    }
    refreshTokenPending.value = true
    try {
      const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
        refresh_token: refreshToken.value
      })
      if (response.data.access_token && response.data.refresh_token) {
        setToken(response.data.access_token)
        setRefreshToken(response.data.refresh_token)
        return true
      }
      return false
    } catch (error) {
      return false
    } finally {
      refreshTokenPending.value = false
    }
  }

  function setUserInfo(info) {
    userInfo.value = info
    if (info) {
      localStorage.setItem('userInfo', JSON.stringify(info))
    } else {
      localStorage.removeItem('userInfo')
    }
  }

  function setRouteConfig(routes) {
    routeConfig.value = Array.isArray(routes) ? routes : []
    if (routeConfig.value.length > 0) {
      localStorage.setItem('routeConfig', JSON.stringify(routeConfig.value))
    } else {
      localStorage.removeItem('routeConfig')
    }
  }

  function setMenuConfig(menu) {
    menuConfig.value = Array.isArray(menu) ? menu : []
    if (menuConfig.value.length > 0) {
      localStorage.setItem('menuConfig', JSON.stringify(menuConfig.value))
    } else {
      localStorage.removeItem('menuConfig')
    }
  }

  function logout() {
    token.value = ''
    refreshToken.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('userInfo')
    routeConfig.value = []
    menuConfig.value = []
    localStorage.removeItem('routeConfig')
    localStorage.removeItem('menuConfig')
  }

  function updateRole(role) {
    if (userInfo.value) {
      userInfo.value.role = role
      localStorage.setItem('userInfo', JSON.stringify(userInfo.value))
    }
  }

  return {
    token,
    refreshToken,
    userInfo,
    routeConfig,
    menuConfig,
    isLoggedIn,
    setToken,
    setRefreshToken,
    refreshAccessToken,
    setUserInfo,
    setRouteConfig,
    setMenuConfig,
    logout,
    updateRole
  }
})
