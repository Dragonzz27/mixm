<template>
  <div class="users">
    <n-card title="用户管理">
      <template #header-extra>
        <n-space>
          <n-input 
            v-model:value="searchKey" 
            placeholder="搜索用户名/邮箱/手机号" 
            clearable 
            style="width: 250px"
            @keyup.enter="loadUsers"
          />
          <n-button type="info" @click="showImportModal = true">批量导入</n-button>
          <n-button type="primary" @click="showAddModal = true">添加用户</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="users"
        :pagination="pagination"
        remote
        :bordered="false"
      />
    </n-card>
    
    <n-modal v-model:show="showImportModal" preset="dialog" title="批量导入学生">
      <n-space vertical>
        <n-alert type="info">
          默认密码123456。角色需先在角色管理中创建。
        </n-alert>
        <n-upload
          accept=".xlsx,.xls"
          :max="1"
          :custom-request="handleImportRequest"
        >
          <n-button>选择Excel文件</n-button>
        </n-upload>
        <n-button text type="primary" @click="downloadTemplate">下载模板</n-button>
      </n-space>
      <template #action>
        <n-button @click="showImportModal = false">关闭</n-button>
      </template>
    </n-modal>
    
    <n-modal v-model:show="showAddModal" preset="dialog" title="添加用户">
      <n-form :model="addFormData" :rules="addRules" ref="formRef">
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="addFormData.username" />
        </n-form-item>
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="addFormData.email" />
        </n-form-item>
        <n-form-item label="手机号" path="phone">
          <n-input v-model:value="addFormData.phone" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input v-model:value="addFormData.password" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item label="角色" path="role_id">
          <n-select v-model:value="addFormData.role_id" :options="roleOptions" placeholder="请选择角色" clearable />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showAddModal = false">取消</n-button>
        <n-button type="primary" :loading="addLoading" @click="handleAdd">确定</n-button>
      </template>
    </n-modal>
    
    <n-modal v-model:show="showEditModal" preset="dialog" title="编辑用户">
      <n-form :model="editFormData" :rules="editRules" ref="formRef">
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="editFormData.username" />
        </n-form-item>
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="editFormData.email" />
        </n-form-item>
        <n-form-item label="手机号" path="phone">
          <n-input v-model:value="editFormData.phone" />
        </n-form-item>
        <n-form-item label="状态" path="is_active">
          <n-switch v-model:value="editFormData.is_active" />
        </n-form-item>
        <n-form-item label="角色" path="role_id">
          <n-select v-model:value="editFormData.role_id" :options="roleOptions" placeholder="请选择角色" clearable />
        </n-form-item>
      </n-form>
      <template #action>
        <div style="display:flex; justify-content:space-between; align-items:center; width:100%;">
          <div>
            <n-button type="info" @click="showPwdModal = true; initPwdForm()">修改密码</n-button>
          </div>
          <div>
            <n-space>
              <n-button @click="showEditModal = false">取消</n-button>
              <n-button type="primary" :loading="editLoading" @click="handleEdit">保存</n-button>
            </n-space>
          </div>
        </div>
      </template>
    </n-modal>
    
    <n-modal v-model:show="showPwdModal" preset="dialog" title="修改密码">
      <n-form :model="pwdFormData" :rules="pwdRules" ref="formRef">
        <n-form-item label="新密码" path="new_password">
          <n-input v-model:value="pwdFormData.new_password" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item label="确认密码" path="confirm_password">
          <n-input v-model:value="pwdFormData.confirm_password" type="password" show-password-on="click" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showPwdModal = false">取消</n-button>
        <n-button type="primary" :loading="pwdLoading" @click="handleChangePassword">确认修改</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, h, watch } from 'vue'
import { NCard, NInput, NButton, NDataTable, NSpace, NModal, NForm, NFormItem, NSwitch, useMessage, NUpload, NAlert, NButtonGroup, NSelect } from 'naive-ui'
import { authApi, roleApi } from '@/api'

const message = useMessage()
const searchKey = ref('')
const showAddModal = ref(false)
const showEditModal = ref(false)
const showPwdModal = ref(false)
const showImportModal = ref(false)
const addLoading = ref(false)
const editLoading = ref(false)
const pwdLoading = ref(false)
const importLoading = ref(false)
const users = ref([])
const roleOptions = ref([])
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  onChange: (page) => {
    pagination.value.page = page
    loadUsers()
  },
  onUpdatePageSize: (pageSize) => {
    pagination.value.pageSize = pageSize
    pagination.value.page = 1
    loadUsers()
  }
})

const addFormData = ref({ username: '', email: '', phone: '', password: '', role_id: null })
const addRules = {
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, message: '请输入密码' }]
}

const editFormData = ref({ id: null, username: '', email: '', phone: '', is_active: true, role_id: null })
const editRules = {
  username: [{ required: true, message: '请输入用户名' }]
}

const pwdFormData = ref({ new_password: '', confirm_password: '' })
const pwdRules = {
  new_password: [{ required: true, message: '请输入新密码' }],
  confirm_password: [
    { required: true, message: '请确认新密码' },
    {
      validator: (rule, value) => value === pwdFormData.value.new_password,
      message: '两次密码不一致',
      trigger: 'blur'
    }
  ]
}

function getRowNumber(index) {
  const page = pagination.value.page || 1
  const pageSize = pagination.value.pageSize || 10
  return (page - 1) * pageSize + index + 1
}

const columns = [
  { title: '序号', key: 'index', width: 80, render: (_, index) => getRowNumber(index) },
  { title: '用户名', key: 'username', width: 120 },
  { title: '邮箱', key: 'email', width: 180 },
  { title: '手机号', key: 'phone', width: 120 },
  {
    title: '角色',
    key: 'role',
    width: 220,
    render: (row) => row.role?.name || '-'
  },
  { title: '状态', key: 'is_active', width: 80, render: (row) => h('span', row.is_active ? '激活' : '禁用') },
  {
    title: '操作',
    key: 'actions',
    render: (row) => h(NSpace, {}, () => [
      h(NButton, { size: 'small', onClick: () => handleEditClick(row) }, () => '编辑'),
      h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, () => '删除')
    ])
  }
]

let searchTimer = null

watch(searchKey, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    pagination.value.page = 1
    loadUsers()
  }, 300)
})

async function loadUsers() {
  try {
    const page = pagination.value.page || 1
    const pageSize = pagination.value.pageSize || 10
    const offset = (page - 1) * pageSize

    const res = await authApi.getUsers({ 
      offset,
      limit: pageSize,
      keyword: searchKey.value || undefined
    })

    users.value = res.items || []
    pagination.value.itemCount = res.total || 0

    const maxPage = Math.max(1, Math.ceil((pagination.value.itemCount || 0) / pageSize))
    if (pagination.value.page > maxPage) {
      pagination.value.page = maxPage
      return loadUsers()
    }
  } catch (err) {
    console.error(err)
  }
}

async function loadRoles() {
  try {
    const res = await roleApi.getList({ offset: 0, limit: 200 })
    roleOptions.value = (res.items || []).map((item) => ({
      label: item.name,
      value: item.id
    }))
  } catch (err) {
    console.error(err)
  }
}



async function handleAdd() {
  addLoading.value = true
  try {
    await authApi.register(addFormData.value)
    message.success('添加成功')
    showAddModal.value = false
    addFormData.value = { username: '', email: '', phone: '', password: '', role_id: null }
    loadUsers()
  } catch (err) {
    console.error(err)
  } finally {
    addLoading.value = false
  }
}

function handleEditClick(row) {
  editFormData.value = {
    id: row.id,
    username: row.username || '',
    email: row.email || '',
    phone: row.phone || '',
    is_active: row.is_active,
    role_id: row.role_id || null
  }
  showEditModal.value = true
}

async function handleEdit() {
  editLoading.value = true
  try {
    await authApi.updateUser(editFormData.value.id, {
      username: editFormData.value.username,
      email: editFormData.value.email,
      phone: editFormData.value.phone,
      is_active: editFormData.value.is_active,
      role_id: editFormData.value.role_id
    })
    message.success('更新成功')
    showEditModal.value = false
    loadUsers()
  } catch (err) {
    console.error(err)
  } finally {
    editLoading.value = false
  }
}

function initPwdForm() {
  pwdFormData.value = { new_password: '', confirm_password: '' }
}

async function handleChangePassword() {
  if (pwdFormData.value.new_password !== pwdFormData.value.confirm_password) {
    message.warning('两次密码不一致')
    return
  }
  pwdLoading.value = true
  try {
    await authApi.changePassword(editFormData.value.id, {
      new_password: pwdFormData.value.new_password
    })
    message.success('密码修改成功')
    showPwdModal.value = false
  } catch (err) {
    console.error(err)
  } finally {
    pwdLoading.value = false
  }
}

async function handleDelete(id) {
  try {
    await authApi.deleteUser(id)
    message.success('删除成功')
    loadUsers()
  } catch (err) {
    console.error(err)
  }
}

function downloadTemplate() {
  import('xlsx').then(XLSX => {
    const data = [
      ['username', 'email', 'phone', 'role_name'],
      ['test001', 'test001@example.com', '13800000001', '学生']
    ]
    const ws = XLSX.utils.aoa_to_sheet(data)
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, 'Users')
    XLSX.writeFile(wb, 'user_template.xlsx')
  })
}

function handleImportRequest({ file }) {
  importLoading.value = true
  authApi.importUsers(file.file).then(res => {
    message.success(`导入完成：成功${res.success}条，失败${res.failed}条`)
    if (res.errors && res.errors.length > 0) {
      console.error('导入错误:', res.errors)
    }
    showImportModal.value = false
    loadUsers()
  }).catch(err => {
    message.error('导入失败')
    console.error(err)
  }).finally(() => {
    importLoading.value = false
  })
}

loadRoles()
loadUsers()
</script>
