<template>
  <n-modal
    v-model:show="showModal"
    preset="card"
    title="拉取镜像"
    style="width: 600px;"
    :mask-closable="false"
    @close="handleClose"
  >
    <n-alert type="info" style="margin-bottom: 16px;">
      请在本地执行以下命令拉取镜像
    </n-alert>

    <n-space vertical :size="16">
      <div class="step-item">
        <div class="step-label">1. 登录仓库（如需要）</div>
        <n-code :code="loginCommand" language="bash" />
      </div>

      <div class="step-item">
        <div class="step-label">2. 拉取镜像</div>
        <n-code :code="pullCommand" language="bash" />
      </div>
    </n-space>

    <template #footer>
      <div style="display: flex; justify-content: flex-end;">
        <n-button @click="handleClose">关闭</n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: Boolean,
  repository: Object
})

const emit = defineEmits(['update:show'])

const registryHost = 'localhost:5000'

const showModal = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})

const loginCommand = computed(() => {
  return `docker login ${registryHost}`
})

const pullCommand = computed(() => {
  const repoName = props.repository?.name || '<repository-name>'
  return `docker pull ${registryHost}/${repoName}:<tag>`
})

function handleClose() {
  emit('update:show', false)
}
</script>

<style scoped>
.step-item {
  background: var(--mixm-paper);
  padding: 12px;
  border-radius: 6px;
}

.step-label {
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--mixm-ink);
}
</style>
