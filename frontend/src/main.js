import { createApp } from 'vue'
import { createPinia } from 'pinia'
import naive from 'naive-ui'
import App from './App.vue'
import router from './router'
import { useContentTypeStore } from '@/stores/contentType'
import './styles/common.css'
import { cssVariables } from './styles/theme'

Object.entries(cssVariables).forEach(([property, value]) => {
  document.documentElement.style.setProperty(property, value)
})

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(naive)

const contentTypeStore = useContentTypeStore()
contentTypeStore.loadContentTypes()

app.mount('#app')
