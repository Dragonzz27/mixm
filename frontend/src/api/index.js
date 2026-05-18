import request from './request'

export const authApi = {
  login(data) {
    return request.post('/auth/login', data)
  },
  register(data) {
    return request.post('/auth/register', data)
  },
  refreshToken(data) {
    return request.post('/auth/refresh', data)
  },
  getUsers(params) {
    return request.get('/auth/users', { params })
  },
  getUser(id) {
    return request.get(`/auth/users/${id}`)
  },
  updateUser(id, data) {
    return request.put(`/auth/users/${id}`, data)
  },
  changePassword(id, data) {
    return request.post(`/auth/users/${id}/change-password`, data)
  },
  deleteUser(id) {
    return request.delete(`/auth/users/${id}`)
  },
  importUsers(file) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/auth/users/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  }
}

export const categoryApi = {
  getList(params) {
    return request.get('/categories/', { params })
  },
  getTree(params) {
    return request.get('/categories/tree', { params })
  },
  getDetail(id) {
    return request.get(`/categories/${id}`)
  },
  create(data) {
    return request.post('/categories/', data)
  },
  update(id, data) {
    return request.put(`/categories/${id}`, data)
  },
  delete(id) {
    return request.delete(`/categories/${id}`)
  }
}

export const contentApi = {
  getList(params) {
    return request.get('/contents/', { params })
  },
  getById(id) {
    return request.get(`/contents/${id}`)
  },
  getDetail(id) {
    return request.get(`/contents/${id}`)
  },
  getByCategory(categoryId, params) {
    return request.get(`/contents/by-category/${categoryId}`, { params })
  },
  getByCategories(categoryIds, params) {
    return request.get('/contents/by-categories', { params: { ...params, category_ids: categoryIds.join(',') } })
  },
  getCategoriesWithContents(params) {
    return request.get('/contents/categories-with-contents', { params })
  },
  getCategoryTree(categoryId) {
    return request.get(`/contents/category-tree/${categoryId}`)
  },
  create(data) {
    return request.post('/contents/', data)
  },
  update(id, data) {
    return request.put(`/contents/${id}`, data)
  },
  delete(id) {
    return request.delete(`/contents/${id}`)
  },
  uploadFile(file) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/contents/upload', formData)
  }
}

export const bindingApi = {
  getCounts(data) {
    return request.post('/bindings/counts', data)
  },
  getAll(params) {
    return request.get('/bindings/', { params })
  },
  getByContent(contentId) {
    return request.get(`/bindings/content/${contentId}`)
  },
  getByCategory(categoryId) {
    return request.get(`/bindings/category/${categoryId}`)
  },
  getUnboundContents(categoryId, params) {
    return request.get(`/bindings/category/${categoryId}/unbound`, { params })
  },
  getCategoryDetail(categoryId) {
    return request.get(`/bindings/detail/category/${categoryId}`)
  },
  batchBind(categoryId, data) {
    return request.post(`/bindings/category/${categoryId}/batch-bind`, data)
  },
  batchUnbind(categoryId, data) {
    return request.delete(`/bindings/category/${categoryId}/batch-unbind`, { data })
  },
  getContentDetail(contentId) {
    return request.get(`/bindings/detail/content/${contentId}`)
  },
  getCategoryDetail(categoryId) {
    return request.get(`/bindings/detail/category/${categoryId}`)
  },
  create(data) {
    return request.post('/bindings/', data)
  },
  update(id, data) {
    return request.put(`/bindings/${id}`, data)
  },
  delete(id) {
    return request.delete(`/bindings/${id}`)
  }
}

export const contentTypeApi = {
  getList(params) {
    return request.get('/content-types/', { params })
  },
  getDetail(id) {
    return request.get(`/content-types/${id}`)
  },
  create(data) {
    return request.post('/content-types/', data)
  },
  delete(id) {
    return request.delete(`/content-types/${id}`)
  }
}

export const templateApi = {
  getList(params) {
    return request.get('/templates/', { params })
  },
  getDetail(id) {
    return request.get(`/templates/${id}`)
  },
  create(data) {
    return request.post('/templates/', data)
  },
  update(id, data) {
    return request.put(`/templates/${id}`, data)
  },
  delete(id) {
    return request.delete(`/templates/${id}`)
  }
}

export const systemApi = {
  getStats() {
    return request.get('/system/stats')
  },
  getLoad() {
    return request.get('/system/load')
  }
}

export const clusterApi = {
  getPods() {
    return request.get('/clusters/pods')
  },
  getNodes() {
    return request.get('/clusters/nodes')
  },
  getDeployments() {
    return request.get('/clusters/deployments')
  },
  deletePod(namespace, podName) {
    return request.delete(`/clusters/pods/${namespace}/${podName}`)
  },
  deleteDeployment(namespace, deploymentName) {
    return request.delete(`/clusters/deployments/${namespace}/${deploymentName}`)
  },
  getPodLogs(namespace, podName, container, tailLines) {
    const params = {}
    if (container) params.container = container
    if (tailLines) params.tail_lines = tailLines
    return request.get(`/clusters/pods/${namespace}/${podName}/logs`, { params })
  },
  describePod(namespace, podName) {
    return request.get(`/clusters/pods/${namespace}/${podName}/describe`)
  },
  getPV() {
    return request.get('/clusters/pv')
  },
  deletePV(pvName) {
    return request.delete(`/clusters/pv/${pvName}`)
  },
  getPVC(namespace) {
    const params = namespace ? { namespace } : {}
    return request.get('/clusters/pvc', { params })
  },
  deletePVC(namespace, pvcName) {
    return request.delete(`/clusters/pvc/${namespace}/${pvcName}`)
  }
}

export const instanceApi = {
  getList(params) {
    return request.get('/instances/', { params })
  },
  getDetail(id) {
    return request.get(`/instances/${id}`)
  },
  create(data) {
    return request.post('/instances/', data)
  },
  startByTemplate(data) {
    return request.post('/instances/start', data)
  },
  startBatchByTemplate(data) {
    return request.post('/instances/start/batch-template', data)
  },
  startQueued(data) {
    return request.post('/instances/start', data)
  },
  update(id, data) {
    return request.put(`/instances/${id}`, data)
  },
  start(id) {
    return request.post(`/instances/${id}/start`)
  },
  stop(id) {
    return request.post(`/instances/${id}/stop`)
  },
  deleteAll() {
    return request.post('/instances/delete/all')
  },
  delete(id) {
    return request.delete(`/instances/${id}`)
  },
  getByUser(userId) {
    return request.get(`/instances/user/${userId}`)
  },
  getMyInstances() {
    return request.get('/instances/me')
  },
  getQueueStatus() {
    return request.get('/instances/queue/status')
  },
  getQueueTasks() {
    return request.get('/instances/queue/tasks')
  },
  getTaskStatus(taskId) {
    return request.get(`/instances/queue/task/${taskId}`)
  }
}

export const imageApi = {
  getList(params) {
    return request.get('/images/', { params })
  },
  getMyList() {
    return request.get('/images/me')
  },
  getDetail(id) {
    return request.get(`/images/${id}`)
  },
  create(data) {
    return request.post('/images/', data)
  },
  update(id, data) {
    return request.put(`/images/${id}`, data)
  },
  delete(id) {
    return request.delete(`/images/${id}`)
  },
  getTags(id) {
    return request.get(`/images/${id}/tags`)
  },
  deleteTag(repoId, tagName) {
    return request.delete(`/images/${repoId}/tags/${tagName}`)
  }
}

export const roleApi = {
  getList(params) {
    return request.get('/roles/', { params })
  },
  getDetail(id) {
    return request.get(`/roles/${id}`)
  },
  create(data) {
    return request.post('/roles/', data)
  },
  update(id, data) {
    return request.put(`/roles/${id}`, data)
  },
  delete(id) {
    return request.delete(`/roles/${id}`)
  },
  getPermissionOptions() {
    return request.get('/roles/permissions')
  },
  getRoleOptions() {
    return request.get('/roles/').then(res => {
      return (res.roles || []).map(item => ({
        label: item.name,
        value: item.id
      }))
    })
  }
}

export const assignmentApi = {
  getList(params) {
    return request.get('/assignments/', { params })
  },
  getActiveList() {
    return request.get('/assignments/active')
  },
  getDetail(id) {
    return request.get(`/assignments/${id}`)
  },
  create(data) {
    return request.post('/assignments/', data)
  },
  update(id, data) {
    return request.put(`/assignments/${id}`, data)
  },
  delete(id) {
    return request.delete(`/assignments/${id}`)
  },
  getSubmissions(assignmentId, params) {
    return request.get(`/assignments/${assignmentId}/submissions`, { params })
  },
  exportSubmissions(assignmentId) {
    return request.get(`/assignments/${assignmentId}/submissions/export`)
  },
  uploadFile(file) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/assignments/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  submit(data) {
    return request.post('/assignments/submissions', data)
  },
  getMySubmissions(params) {
    return request.get('/assignments/my/submissions', { params })
  },
  updateSubmission(id, data) {
    return request.put(`/assignments/submissions/${id}`, data)
  },
  deleteSubmission(id) {
    return request.delete(`/assignments/submissions/${id}`)
  }
}

export const questionApi = {
  getList(params) {
    return request.get('/questions/', { params })
  },
  getDetail(id) {
    return request.get(`/questions/${id}`)
  },
  create(data) {
    return request.post('/questions/', data)
  },
  update(id, data) {
    return request.put(`/questions/${id}`, data)
  },
  delete(id) {
    return request.delete(`/questions/${id}`)
  },
  batchCreate(questions) {
    return request.post('/questions/batch', questions)
  },
  getAiGenerated(params) {
    return request.get('/questions/ai-generated', { params })
  },
  generateQuestion(data, onChunk, onDone, onError) {
    const token = localStorage.getItem('token')
    return fetch('/api/questions/generate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data)
    }).then(response => {
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      function read() {
        reader.read().then(({ done, value }) => {
          if (done) {
            if (buffer) {
              try {
                const data = JSON.parse(buffer)
                if (data.type === 'done') {
                  onDone(data.data)
                } else if (data.type === 'error') {
                  onError(data.message)
                }
              } catch (e) {}
            }
            return
          }
          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop()
          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6))
                if (data.type === 'chunk') {
                  onChunk(data.content)
                } else if (data.type === 'done') {
                  onDone(data.data)
                } else if (data.type === 'error') {
                  onError(data.message)
                }
              } catch (e) {}
            }
          }
          read()
        })
      }
      read()
    }).catch(onError)
  }
}

export const questionCategoryApi = {
  getList(params) {
    return request.get('/question-categories/', { params })
  },
  getTree() {
    return request.get('/question-categories/tree')
  },
  getDetail(id) {
    return request.get(`/question-categories/${id}`)
  },
  create(data) {
    return request.post('/question-categories/', data)
  },
  update(id, data) {
    return request.put(`/question-categories/${id}`, data)
  },
  delete(id) {
    return request.delete(`/question-categories/${id}`)
  }
}

export const knowledgeTreeApi = {
  analyze(data) {
    return request.post('/knowledge-tree/analyze', data)
  },
  analyzeStream(data, onChunk, onDone, onError) {
    const token = localStorage.getItem('token')
    return fetch('/api/knowledge-tree/analyze', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data)
    }).then(response => {
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      function read() {
        reader.read().then(({ done, value }) => {
          if (done) {
            if (buffer) {
              try {
                const data = JSON.parse(buffer)
                if (data.type === 'done') {
                  onDone(data.data)
                } else if (data.type === 'error') {
                  onError(data.message)
                }
              } catch (e) {}
            }
            return
          }
          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop()
          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6))
                if (data.type === 'chunk') {
                  onChunk(data.content)
                } else if (data.type === 'done') {
                  onDone(data.data)
                } else if (data.type === 'error') {
                  onError(data.message)
                }
              } catch (e) {}
            }
          }
          read()
        })
      }
      read()
    }).catch(onError)
  },
  save(data) {
    return request.post('/knowledge-tree/save', data)
  }
}

export const simulationApi = {
  generate(data) {
    return request.post('/simulation/generate', data)
  },
  submit(data) {
    return request.post('/simulation/submit', data)
  }
}
