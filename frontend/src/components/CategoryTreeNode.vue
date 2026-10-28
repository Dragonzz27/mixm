<template>
  <div class="tree-node-wrapper">
    <div class="tree-node" :style="{ paddingLeft: (depth * 24) + 'px' }" @click.stop="$emit('toggle', category.id)">
      <template v-if="category.children?.length">
        <n-icon class="expand-icon">
          <ChevronForwardOutline v-if="!expandedKeys.includes(category.id)" />
          <ChevronDownOutline v-else />
        </n-icon>
      </template>
      <span v-else class="expand-placeholder"></span>
      <!-- <span class="node-icon">📁</span> -->
      <span class="node-name">{{ category.name }}</span>
      <span class="node-level">({{ getLevelName(category.level) }})</span>
      <span class="node-count" v-if="category.bound_count > 0">({{ category.bound_count }})</span>
      <div class="node-actions">
        <n-button size="small" @click.stop="$emit('add-child', category)">添加子分类</n-button>
        <n-button size="small" @click.stop="$emit('edit', category)">编辑</n-button>
        <n-button size="small" type="info" @click.stop="$emit('bind', category)">绑定内容</n-button>
        <n-button size="small" type="error" @click.stop="$emit('delete', category.id)">删除</n-button>
      </div>
    </div>
    <div v-if="expandedKeys.includes(category.id) && category.children?.length" class="tree-children">
      <CategoryTreeNode
        v-for="child in category.children"
        :key="child.id"
        :category="child"
        :expanded-keys="expandedKeys"
        :depth="depth + 1"
        @toggle="$emit('toggle', $event)"
        @add-child="$emit('add-child', $event)"
        @edit="$emit('edit', $event)"
        @delete="$emit('delete', $event)"
        @bind="$emit('bind', $event)"
      />
    </div>
  </div>
</template>

<script setup>
import { NIcon, NButton } from 'naive-ui'
import { ChevronForwardOutline, ChevronDownOutline } from '@vicons/ionicons5'

const levelNames = {
  1: '课程',
  2: '章节',
  3: '小节',
  4: '知识点',
  5: '知识元',
}

function getLevelName(level) {
  return levelNames[level] || `L${level}`
}

defineProps({
  category: {
    type: Object,
    required: true
  },
  expandedKeys: {
    type: Array,
    default: () => []
  },
  depth: {
    type: Number,
    default: 0
  }
})

defineEmits(['toggle', 'add-child', 'edit', 'delete', 'bind'])
</script>

<style scoped>
.tree-node-wrapper {
  width: 100%;
}

.tree-node {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--mixm-border);
  transition: background 0.2s;
}

.tree-node:hover {
  background: var(--mixm-surface-muted);
}

.expand-icon {
  width: 20px;
  height: 20px;
  cursor: pointer;
  color: var(--mixm-muted);
  flex-shrink: 0;
}

.expand-placeholder {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.node-icon {
  margin-right: 8px;
  font-size: 14px;
}

.node-name {
  font-weight: 500;
  color: var(--mixm-ink);
  margin-right: 8px;
}

.node-level {
  font-size: 12px;
  color: var(--mixm-faint);
  margin-right: 8px;
}

.node-count {
  font-size: 12px;
  color: var(--mixm-accent);
  margin-right: 12px;
}

.node-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.tree-node:hover .node-actions {
  opacity: 1;
}

.tree-children {
  background: var(--mixm-surface-muted);
}
</style>
