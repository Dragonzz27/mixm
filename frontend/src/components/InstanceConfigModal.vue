<template>
  <n-modal v-model:show="showModal" preset="dialog" title="实例配置" style="width: 550px;">
    <n-descriptions :column="1" label-placement="left" bordered>
      <n-descriptions-item label="实例ID">
        <n-tag type="info">{{ config.id }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item label="实例名称">
        {{ config.name }}
      </n-descriptions-item>
      <n-descriptions-item label="镜像">
        <n-tag>{{ config.image }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item label="CPU限制">
        {{ config.cpuLimit }} 核
      </n-descriptions-item>
      <n-descriptions-item label="内存限制">
        {{ config.memoryLimit }}
      </n-descriptions-item>
      <n-descriptions-item label="临时存储">
        {{ config.storageLimit }} Gi
      </n-descriptions-item>
      <n-descriptions-item label="GPU数量">
        {{ config.gpuLimit || 0 }} 张
      </n-descriptions-item>
      <n-descriptions-item label="单卡显存">
        {{ formatGpuMemory(config.gpuMemoryLimit) }}
      </n-descriptions-item>
      <n-descriptions-item label="命名空间">
        {{ config.namespace }}
      </n-descriptions-item>
      <n-descriptions-item label="SSH地址">
        <n-input :value="config.sshConnection" readonly>
          <template #suffix>
            <n-button quaternary size="small" @click="copySshConnection">复制</n-button>
          </template>
        </n-input>
      </n-descriptions-item>
      <n-descriptions-item label="SSH用户名">
        <n-tag type="success">{{ config.sshUsername }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item label="SSH端口">
        <n-tag type="warning">{{ config.sshPort }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item label="SSH密码">
        <n-tag type="warning">{{ config.sshPassword }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item v-if="showCodeServerPath" label="Code-Server访问路径">
        <n-input :value="codeServerUrl" readonly>
          <template #suffix>
            <n-button quaternary size="small" @click="copyCodeServerUrl">复制</n-button>
          </template>
        </n-input>
      </n-descriptions-item>
    </n-descriptions>
  </n-modal>
</template>

<script setup>
import { computed } from 'vue'
import { NModal, NDescriptions, NDescriptionsItem, NTag, NInput, NButton, useMessage } from 'naive-ui'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  config: {
    type: Object,
    default: () => null
  }
})

const emit = defineEmits(['update:show'])

const message = useMessage()

const showModal = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})

const showCodeServerPath = computed(() => {
  const imageName = (props.config?.image || '').toLowerCase()
  return imageName.includes('codeserver') || imageName.includes('code-server')
})

const codeServerUrl = computed(() => {
  if (!showCodeServerPath.value) return ''
  const hostIp = props.config?.sshHost || ''
  const endpoint = props.config?.endpoint || `/${props.config?.id || ''}`
  const port = props.config?.accessPort || 80
  const password = props.config?.codeServerPassword || ''
  if (!hostIp) return ''
  return `http://${hostIp}:${port}${endpoint}/`
})

function copySshConnection() {
  if (props.config?.sshConnection) {
    navigator.clipboard.writeText(props.config.sshConnection)
    message.success('SSH连接命令已复制到剪贴板')
  }
}

function copyCodeServerUrl() {
  if (!codeServerUrl.value) return
  navigator.clipboard.writeText(codeServerUrl.value)
  message.success('Code-Server访问路径已复制到剪贴板')
}

function formatGpuMemory(mb) {
  const memoryMb = Number(mb) || 0
  if (memoryMb <= 0) return '-'
  if (memoryMb % 1024 === 0) {
    return `${memoryMb / 1024}Gi`
  }
  return `${memoryMb}Mi`
}
</script>
