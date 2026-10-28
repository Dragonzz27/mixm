import { API_BASE_URL } from './config'
import { unwrapEnvelope, resolveFileURLs } from './protocol'
import axios from 'axios'
import { useUserStore } from '@/stores/user'
import { createDiscreteApi } from 'naive-ui'
import router, { resetRouter } from '@/router'
import { themeOverrides } from '@/styles/theme'

let message
try {
  const msgApi = createDiscreteApi(['message'], {
    configProviderProps: {
      themeOverrides
    }
  })
  message = msgApi.message
} catch (e) {
  message = {
    error: (msg) => console.error(msg),
    success: (msg) => console.log(msg),
    warning: (msg) => console.warn(msg),
    info: (msg) => console.info(msg)
  }
}

let isRefreshing = false
let requestQueue = []

function processQueue(error, token = null) {
  requestQueue.forEach(promise => {
    if (error) {
      promise.reject(error)
    } else {
      promise.resolve(token)
    }
  })
  requestQueue = []
}

const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    if (config.data instanceof FormData) {
      delete config.headers['Content-Type']
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    return resolveFileURLs(unwrapEnvelope(response.data), API_BASE_URL)
  },
  async (error) => {
    const { response, config } = error
    const userStore = useUserStore()

    if (response) {
      if (config.url?.includes('/auth/login') || config.url?.includes('/auth/refresh')) {
        return Promise.reject(error)
      }

      if (response.status === 401 && !config._retry) {
        config._retry = true
        if (!isRefreshing) {
          isRefreshing = true

          const refreshed = await userStore.refreshAccessToken()

          if (refreshed) {
            const newToken = userStore.token
            config.headers.Authorization = `Bearer ${newToken}`
            processQueue(null, newToken)
            isRefreshing = false
            return request(config)
          }

          processQueue(error, null)
          isRefreshing = false
          message.error('登录已过期，请重新登录')
          userStore.logout()
          resetRouter()
          router.replace('/login')
          return Promise.reject(error)
        }

        return new Promise((resolve, reject) => {
          requestQueue.push({ resolve, reject })
        }).then(token => {
          config.headers.Authorization = `Bearer ${token}`
          return request(config)
        }).catch(err => {
          return Promise.reject(err)
        })
      }

      switch (response.status) {
        case 403:
          message.error('没有权限访问')
          break
        case 404:
          message.error('请求的资源不存在')
          break
        case 500:
          message.error('服务器错误')
          break
        default:
          message.error(response.data?.message || '请求失败')
      }
    } else {
      message.error('网络连接失败')
    }
    return Promise.reject(error)
  }
)

export default request
