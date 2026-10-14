<template>
  <div class="roles">
    <n-card title="角色管理">
      <template #header-extra>
        <n-space>
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索角色名"
            clearable
            style="width: 220px"
            @keyup.enter="handleSearch"
          />
          <n-button type="primary" @click="openCreateModal">新增角色</n-button>
        </n-space>
      </template>

      <n-data-table
        :columns="columns"
        :data="roles"
        :pagination="pagination"
        remote
        :bordered="false"
      />
    </n-card>

    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑角色' : '新增角色'"
      style="width: 680px"
    >
      <n-form :model="formData" :rules="rules" ref="formRef" label-placement="left" label-width="100px">
        <n-form-item label="角色名" path="name">
          <n-input v-model:value="formData.name" placeholder="请输入角色名" />
        </n-form-item>
        <n-form-item label="权限" path="route_keys">
          <n-checkbox-group v-model:value="formData.route_keys">
            <n-space>
              <n-checkbox
                v-for="option in permissionOptions"
                :key="option.value"
                :value="option.value"
                :label="option.label"
              />
            </n-space>
          </n-checkbox-group>
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showModal = false">取消</n-button>
        <n-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { h, onMounted, ref } from 'vue'
import {
  NButton,
  NCard,
  NCheckbox,
  NCheckboxGroup,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSpace,
  useDialog,
  useMessage
} from 'naive-ui'
import { roleApi } from '@/api'

const message = useMessage()
const dialog = useDialog()

const roles = ref([])
const showModal = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const searchKeyword = ref('')
const permissionOptions = ref([])
const formRef = ref(null)

const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    pagination.value.page = page
    loadRoles()
  },
  onUpdatePageSize: (pageSize) => {
    pagination.value.pageSize = pageSize
    pagination.value.page = 1
    loadRoles()
  }
})

const formData = ref({
  id: '',
  name: '',
  route_keys: []
})

const rules = {
  name: [{ required: true, message: '请输入角色名' }]
}

function getRowNumber(index) {
  const page = pagination.value.page || 1
  const pageSize = pagination.value.pageSize || 10
  return (page - 1) * pageSize + index + 1
}

const columns = [
  {
    title: '序号',
    key: 'index',
    width: 80,
    render: (_, index) => getRowNumber(index)
  },
  { title: '角色名', key: 'name', width: 180 },
  {
    title: '权限',
    key: 'route_keys',
    render: (row) => {
      const labels = (row.route_keys || []).map((key) => {
        const found = permissionOptions.value.find((item) => item.value === key)
        return found ? found.label : key
      })
      return labels.join('、') || '-'
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    render: (row) => h(NSpace, {}, () => [
      h(
        NButton,
        {
          size: 'small',
          onClick: () => openEditModal(row)
        },
        () => '编辑'
      ),
      h(
        NButton,
        {
          size: 'small',
          type: 'error',
          onClick: () => handleDelete(row)
        },
        () => '删除'
      )
    ])
  }
]

async function loadPermissionOptions() {
  try {
    const res = await roleApi.getPermissionOptions()
    permissionOptions.value = (res || []).map((item) => ({
      label: item.label,
      value: item.key
    }))
  } catch (err) {
    console.error(err)
    message.error('读取权限选项失败')
  }
}

async function loadRoles() {
  try {
    const page = pagination.value.page || 1
    const pageSize = pagination.value.pageSize || 10
    const params = {
      offset: (page - 1) * pageSize,
      limit: pageSize,
      keyword: searchKeyword.value || undefined
    }
    const res = await roleApi.getList(params)
    roles.value = res.items || []
    pagination.value.itemCount = res.total || 0

    const maxPage = Math.max(1, Math.ceil((pagination.value.itemCount || 0) / pageSize))
    if (pagination.value.page > maxPage) {
      pagination.value.page = maxPage
      return loadRoles()
    }
  } catch (err) {
    console.error(err)
    message.error('读取角色列表失败')
  }
}

function handleSearch() {
  pagination.value.page = 1
  loadRoles()
}

function openCreateModal() {
  isEdit.value = false
  formData.value = {
    id: '',
    name: '',
    route_keys: []
  }
  showModal.value = true
}

function openEditModal(row) {
  isEdit.value = true
  formData.value = {
    id: row.id,
    name: row.name,
    route_keys: [...(row.route_keys || [])]
  }
  showModal.value = true
}

async function handleSubmit() {
  submitLoading.value = true
  try {
    const payload = {
      name: formData.value.name,
      route_keys: formData.value.route_keys
    }

    if (isEdit.value) {
      await roleApi.update(formData.value.id, payload)
      message.success('角色更新成功')
    } else {
      await roleApi.create(payload)
      message.success('角色创建成功')
    }

    showModal.value = false
    loadRoles()
  } catch (err) {
    console.error(err)
    message.error(err?.response?.data?.message || '提交失败')
  } finally {
    submitLoading.value = false
  }
}

function handleDelete(row) {
  dialog.warning({
    title: '删除确认',
    content: `确认删除角色“${row.name}”吗？`,
    positiveText: '确认',
    negativeText: '取消',
    async onPositiveClick() {
      try {
        await roleApi.delete(row.id)
        message.success('删除成功')
        loadRoles()
      } catch (err) {
        console.error(err)
        message.error(err?.response?.data?.message || '删除失败')
      }
    }
  })
}

onMounted(async () => {
  await loadPermissionOptions()
  await loadRoles()
})
</script>
