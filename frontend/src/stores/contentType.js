import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { contentTypeApi } from '@/api'

export const useContentTypeStore = defineStore('contentType', () => {
  const contentTypes = ref([])
  const typeMap = computed(() => {
    const map = {}
    contentTypes.value.forEach(t => {
      map[t.code] = t.name
    })
    return map
  })

  async function loadContentTypes() {
    try {
      const res = await contentTypeApi.getList({})
      contentTypes.value = res.items || []
    } catch (err) {
      console.error('Failed to load content types:', err)
    }
  }

  function getTypeName(code) {
    return typeMap.value[code] || code || '-'
  }

  return {
    contentTypes,
    typeMap,
    loadContentTypes,
    getTypeName
  }
})
