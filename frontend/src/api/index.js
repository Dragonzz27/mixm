import { stream } from './stream'
import request from './request'

export const authApi = {
  login(data) {
    return request.post('/auth/login', data)
  },
  register(data) {
    return request.post('/users', data)
  },
  refreshToken(data) {
    return request.post('/auth/refresh', data)
  },
  getUsers(params) {
    return request.get('/users', { params })
  },
  getUser(id) {
    return request.get(`/users/${id}`)
  },
  updateUser(id, data) {
    return request.patch(`/users/${id}`, data)
  },
  changePassword(id, data) {
    return request.put(`/users/${id}/password`, data)
  },
  deleteUser(id) {
    return request.delete(`/users/${id}`)
  },
  importUsers(file) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/users/imports', formData, {
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
    return request.patch(`/categories/${id}`, data)
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
    return request.get(`/categories/${categoryId}/contents`, { params })
  },
  getByCategories(categoryIds, params) {
    return request.get('/contents', { params: { ...params, category_ids: categoryIds.join(',') } })
  },
  getCategoriesWithContents(params) {
    return request.get('/categories', { params: { ...params, root_only: true, include: 'content-preview' } })
  },
  getCategoryTree(categoryId) {
    return request.get(`/categories/${categoryId}/content-tree`)
  },
  create(data) {
    return request.post('/contents/', data)
  },
  update(id, data) {
    return request.patch(`/contents/${id}`, data)
  },
  delete(id) {
    return request.delete(`/contents/${id}`)
  },
  uploadFile(file) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/contents/files', formData)
  }
}

export const bindingApi = {
  getCounts(data) {
    return request.get('/binding-counts', { params: { category_ids: data.category_ids.join(',') } })
  },
  getAll(params) {
    return request.get('/bindings/', { params })
  },
  getByContent(contentId) {
    return request.get('/bindings', { params: { content_id: contentId } })
  },
  getByCategory(categoryId) {
    return request.get('/bindings', { params: { category_id: categoryId } })
  },
  getUnboundContents(categoryId, params) {
    return request.get('/contents', { params: { ...params, unbound_category_id: categoryId } })
  },
  batchBind(categoryId, data) {
    return request.post(`/categories/${categoryId}/content-bindings`, data)
  },
  batchUnbind(categoryId, data) {
    return request.delete(`/categories/${categoryId}/content-bindings`, { data })
  },
  getContentDetail(contentId) {
    return request.get(`/contents/${contentId}`)
  },
  getCategoryDetail(categoryId) {
    return request.get(`/categories/${categoryId}`, { params: { include: 'contents' } })
  },
  create(data) {
    return request.post('/bindings/', data)
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
    return request.patch(`/templates/${id}`, data)
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
    return request.delete(`/clusters/namespaces/${namespace}/pods/${podName}`)
  },
  deleteDeployment(namespace, deploymentName) {
    return request.delete(`/clusters/namespaces/${namespace}/deployments/${deploymentName}`)
  },
  getPodLogs(namespace, podName, container, tailLines) {
    const params = {}
    if (container) params.container = container
    if (tailLines) params.tail_lines = tailLines
    return request.get(`/clusters/namespaces/${namespace}/pods/${podName}/logs`, { params })
  },
  describePod(namespace, podName) {
    return request.get(`/clusters/namespaces/${namespace}/pods/${podName}/describe`)
  },
  getPV() {
    return request.get('/clusters/persistent-volumes')
  },
  deletePV(pvName) {
    return request.delete(`/clusters/persistent-volumes/${pvName}`)
  },
  getPVC(namespace) {
    const params = namespace ? { namespace } : {}
    return request.get('/clusters/persistent-volume-claims', { params })
  },
  deletePVC(namespace, pvcName) {
    return request.delete(`/clusters/namespaces/${namespace}/persistent-volume-claims/${pvcName}`)
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
    return request.post('/instances', data)
  },
  startBatchByTemplate(data) {
    const { template_id, ...payload } = data
    return request.post(`/templates/${template_id}/instances`, payload)
  },
  startQueued(data) {
    return request.post('/instances', data)
  },
  update(id, data) {
    return request.patch(`/instances/${id}`, data)
  },
  start(id) {
    return request.patch(`/instances/${id}`, { status: 'running' })
  },
  stop(id) {
    return request.patch(`/instances/${id}`, { status: 'stopped' })
  },
  deleteAll() {
    return request.delete('/instances')
  },
  delete(id) {
    return request.delete(`/instances/${id}`)
  },
  getByUser(userId) {
    return request.get(`/users/${userId}/instances`)
  },
  getMyInstances() {
    return request.get('/users/me/instances')
  },
  getQueueStatus() {
    return request.get('/instance-tasks/summary')
  },
  getQueueTasks() {
    return request.get('/instance-tasks')
  },
  getTaskStatus(taskId) {
    return request.get(`/instance-tasks/${taskId}`)
  }
}

export const imageApi = {
  getList(params) {
    return request.get('/images/', { params })
  },
  getMyList() {
    return request.get('/users/me/images')
  },
  getDetail(id) {
    return request.get(`/images/${id}`)
  },
  create(data) {
    return request.post('/images/', data)
  },
  update(id, data) {
    return request.patch(`/images/${id}`, data)
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
    return request.patch(`/roles/${id}`, data)
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
    return request.get('/assignments', { params: { active: true } })
  },
  getDetail(id) {
    return request.get(`/assignments/${id}`)
  },
  create(data) {
    return request.post('/assignments/', data)
  },
  update(id, data) {
    return request.patch(`/assignments/${id}`, data)
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
    return request.post('/assignments/files', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  submit(data) {
    const { assignment_id, ...payload } = data
    return request.post(`/assignments/${assignment_id}/submissions`, payload)
  },
  getMySubmissions(params) {
    return request.get('/users/me/submissions', { params })
  },
  updateSubmission(id, data) {
    return request.patch(`/submissions/${id}`, data)
  },
  deleteSubmission(id) {
    return request.delete(`/submissions/${id}`)
  }
}

export const practiceApi = {
  getList(params) {
    return request.get('/practices/', { params })
  },
  getDetail(id) {
    return request.get(`/practices/${id}`)
  },
  create(data) {
    return request.post('/practices/', data)
  },
  update(id, data) {
    return request.patch(`/practices/${id}`, data)
  },
  delete(id) {
    return request.delete(`/practices/${id}`)
  },
  autoGenerate(data) {
    return request.post('/practices/generations', data)
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
    return request.patch(`/questions/${id}`, data)
  },
  review(id, data) {
    return request.put(`/questions/${id}/review`, data)
  },
  delete(id) {
    return request.delete(`/questions/${id}`)
  },
  batchCreate(questions) {
    return request.post('/questions/batches', questions)
  },
  getAiGenerated(params) {
    return request.get('/questions', { params: { ...params, is_ai_generated: true } })
  },
  generateQuestion(data, onChunk, onDone, onError) {
    return stream('/questions/generations', data, { onChunk, onDone, onError })
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
    return request.patch(`/question-categories/${id}`, data)
  },
  delete(id) {
    return request.delete(`/question-categories/${id}`)
  }
}

export const knowledgeTreeApi = {
  analyze(data) {
    return request.post('/knowledge-tree/analyses', data)
  },
  analyzeStream(data, onChunk, onDone, onError) {
    return stream('/knowledge-tree/analyses', data, { onChunk, onDone, onError })
  },
  save(data) {
    return request.post('/knowledge-tree', data)
  }
}

export const simulationApi = {
  generate(data) {
    return request.post('/simulation-attempts', data)
  },
  submit(data) {
    const { attempt_id, ...payload } = data
    return request.post(`/simulation-attempts/${attempt_id}/submissions`, payload)
  }
}
