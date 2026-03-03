<template>
  <n-modal v-model:show="showModal" preset="dialog" title="实例配置" style="width: 500px;">
    <n-descriptions :column="1" label-placement="left" bordered>
      <n-descriptions-item label="实例ID">
        <n-tag type="info">{{ config.id }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item label="实例名称">
        {{ config.name }}
      </n-descriptions-item>
      <n-descriptions-item label="镜像版本">
        <n-tag>{{ config.image }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item label="访问地址">
        <n-input :value="config.accessUrl" readonly>
          <template #suffix>
            <n-button quaternary size="small" @click="copyAccessUrl">复制</n-button>
          </template>
        </n-input>
      </n-descriptions-item>
      <n-descriptions-item label="访问密码">
        <n-input :value="config.password" readonly type="password" show-password-on="click">
          <template #suffix>
            <n-button quaternary size="small" @click="copyPassword">复制</n-button>
          </template>
        </n-input>
      </n-descriptions-item>
      <n-descriptions-item label="CPU限制">
        {{ config.cpuLimit }}
      </n-descriptions-item>
      <n-descriptions-item label="内存限制">
        {{ config.memoryLimit }}
      </n-descriptions-item>
      <n-descriptions-item label="命名空间">
        {{ config.namespace }}
      </n-descriptions-item>
      <n-descriptions-item label="端口映射">
        <n-tag type="warning">{{ config.portMapping }}</n-tag>
      </n-descriptions-item>
    </n-descriptions>
  </n-modal>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
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

function copyAccessUrl() {
  if (props.config?.accessUrl) {
    navigator.clipboard.writeText(props.config.accessUrl)
    message.success('访问地址已复制到剪贴板')
  }
}

function copyPassword() {
  if (props.config?.password) {
    navigator.clipboard.writeText(props.config.password)
    message.success('密码已复制到剪贴板')
  }
}
</script>
