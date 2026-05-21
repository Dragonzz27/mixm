export const DEFAULT_ASSIGNMENT_ATTACHMENT_NAMES = [
  '实验指导',
  '实验报告模板',
  '学习视频'
]

export const ASSIGNMENT_ATTACHMENT_ACCEPT = '.mp4,.webm,.ogg,.mov,.m4v,.pdf,.doc,.docx,.txt,.md'

export const ASSIGNMENT_ATTACHMENT_ALLOWED_EXTENSIONS = [
  '.mp4',
  '.webm',
  '.ogg',
  '.mov',
  '.m4v',
  '.pdf',
  '.doc',
  '.docx',
  '.txt',
  '.md'
]

export function createDefaultAssignmentAttachments() {
  return DEFAULT_ASSIGNMENT_ATTACHMENT_NAMES.map(name => ({
    name,
    file_path: '',
    file_name: '',
    file_size: null,
    file_url: ''
  }))
}

export function normalizeAssignmentAttachments(attachments, { keepEmpty = true } = {}) {
  const list = Array.isArray(attachments) ? attachments : []
  const normalized = list
    .map(item => ({
      name: (item?.name || '').trim(),
      file_path: item?.file_path || '',
      file_name: item?.file_name || '',
      file_size: item?.file_size ?? null,
      file_url: item?.file_url || ''
    }))
    .filter(item => item.name)

  return keepEmpty ? normalized : normalized.filter(item => item.file_path || item.file_url)
}

export function getAssignmentAttachmentExtension(fileName = '') {
  if (!fileName.includes('.')) return ''
  return `.${fileName.split('.').pop().toLowerCase()}`
}

export function isAssignmentAttachmentFileValid(fileName = '') {
  return ASSIGNMENT_ATTACHMENT_ALLOWED_EXTENSIONS.includes(
    getAssignmentAttachmentExtension(fileName)
  )
}

export function formatAttachmentSize(bytes) {
  if (!bytes && bytes !== 0) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}
