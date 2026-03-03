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
          <n-button type="primary" @click="showAddModal = true">添加用户</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="users"
        :pagination="pagination"
        :bordered="false"
      />
    </n-card>
    
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
      </n-form>
      <template #action>
        <n-button @click="showEditModal = false">取消</n-button>
        <n-button type="primary" :loading="editLoading" @click="handleEdit">保存</n-button>
        <n-button type="info" @click="showPwdModal = true; initPwdForm()">修改密码</n-button>
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
import { NCard, NInput, NButton, NDataTable, NSpace, NModal, NForm, NFormItem, NSwitch, useMessage } from 'naive-ui'
import { authApi } from '@/api'

const message = useMessage()
const searchKey = ref('')
const showAddModal = ref(false)
const showEditModal = ref(false)
const showPwdModal = ref(false)
const addLoading = ref(false)
const editLoading = ref(false)
const pwdLoading = ref(false)
const users = ref([])
const pagination = ref({ pageSize: 10 })

const addFormData = ref({ username: '', email: '', phone: '', password: '' })
const addRules = {
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, message: '请输入密码' }]
}

const editFormData = ref({ id: null, username: '', email: '', phone: '', is_active: true })
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

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '用户名', key: 'username', width: 120 },
  { title: '邮箱', key: 'email', width: 180 },
  { title: '手机号', key: 'phone', width: 120 },
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

async function loadUsers() {
  try {
    const res = await authApi.getUsers({ 
      skip: 0, 
      limit: pagination.value.pageSize,
      keyword: searchKey.value || undefined
    })
    users.value = res.users || []
    pagination.value.pageSize = 10
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
    addFormData.value = { username: '', email: '', phone: '', password: '' }
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
    is_active: row.is_active
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
      is_active: editFormData.value.is_active
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

loadUsers()
</script>
