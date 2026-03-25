<template>
  <div class="auth-layout">
    <div class="auth-sidebar">
      <div class="sidebar-bg"></div>
      <div class="sidebar-content">
        <div class="brand">
          <div class="brand-icon">
            <n-icon :size="24" :component="UnlinkSharp" />
          </div>
          <span class="brand-name">MutiM</span>
        </div>
        <div class="testimonial">
          <p class="quote">"Efficiency is doing things right; effectiveness is doing the right things."</p>
          <div class="author-info">
            <p class="author-title">Power By</p>
            <p class="author-name">Ecust CCIOT & Yoobit.cn</p>
            
          </div>
        </div>
      </div>
    </div>
    <div class="auth-main">
      <div class="auth-card-wrapper">
        <div class="auth-card">
          <n-tabs type="segment" animated :default-value="activeTab" @update:value="handleTabChange" class="auth-tabs">
            <n-tab-pane name="login" tab="Login">
              <div class="auth-header">
                <h1 class="auth-title">Welcome back</h1>
                <p class="auth-description">Enter your email below to login to your account</p>
              </div>
              <n-form ref="loginFormRef" :model="loginFormData" :rules="loginRules" class="auth-form">
                <div class="form-field">
                  <label class="form-label">Email</label>
                  <n-input v-model:value="loginFormData.account" placeholder="m@example.com" />
                </div>
                <div class="form-field">
                  <div class="form-label-row">
                    <label class="form-label">Password</label>
                    <a href="#" class="forgot-link">Forgot your password?</a>
                  </div>
                  <n-input v-model:value="loginFormData.password" type="password" show-password-on="click"
                    placeholder="Enter your password" @keydown.enter="handleLogin">
                    <template #password-visible-icon>
                      <n-icon :size="16" :component="GlassesOutline" />
                    </template>
                    <template #password-invisible-icon>
                      <n-icon :size="16" :component="Glasses" />
                    </template>
                  </n-input>
                </div>
              </n-form>
              <n-button type="primary" block size="large" :loading="loginLoading" @click="handleLogin" class="submit-btn">
                Login
              </n-button>
            </n-tab-pane>
            <n-tab-pane name="register" tab="Sign up">
              <div class="auth-header">
                <h1 class="auth-title">Create an account</h1>
                <p class="auth-description">Enter your email below to create your account</p>
              </div>
              <n-form ref="registerFormRef" :model="registerFormData" :rules="registerRules" class="auth-form">
                <div class="form-field">
                  <label class="form-label">Username</label>
                  <n-input v-model:value="registerFormData.username" placeholder="Enter your username" />
                </div>
                <div class="form-field">
                  <label class="form-label">Email</label>
                  <n-input v-model:value="registerFormData.email" placeholder="m@example.com" />
                </div>
                <div class="form-field">
                  <label class="form-label">Phone</label>
                  <n-input v-model:value="registerFormData.phone" placeholder="Enter your phone number" />
                </div>
                <div class="form-field">
                  <label class="form-label">Password</label>
                  <n-input v-model:value="registerFormData.password" type="password" show-password-on="click"
                    placeholder="Create a password">
                    <template #password-visible-icon>
                      <n-icon :size="16" :component="GlassesOutline" />
                    </template>
                    <template #password-invisible-icon>
                      <n-icon :size="16" :component="Glasses" />
                    </template>
                  </n-input>
                </div>
                <div class="form-field">
                  <label class="form-label">Confirm Password</label>
                  <n-input v-model:value="registerFormData.confirmPassword" type="password" show-password-on="click"
                    placeholder="Confirm your password" @keydown.enter="handleRegister">
                    <template #password-visible-icon>
                      <n-icon :size="16" :component="GlassesOutline" />
                    </template>
                    <template #password-invisible-icon>
                      <n-icon :size="16" :component="Glasses" />
                    </template>
                  </n-input>
                </div>
              </n-form>
              <n-button type="primary" block size="large" :loading="registerLoading" @click="handleRegister" class="submit-btn">
                Create Account
              </n-button>
            </n-tab-pane>
          </n-tabs>
          <div class="divider">
            <span>Or continue with</span>
          </div>
          <n-button block size="large" class="social-btn">
            <n-icon :size="20" :component="LogoGithub" class="social-icon" />
            GitHub
          </n-button>
          <div class="auth-footer">
            <p v-if="activeTab === 'login'">Don't have an account? <a href="javascript:;" @click="activeTab = 'register'">Sign up</a></p>
            <p v-else>Already have an account? <a href="javascript:;" @click="activeTab = 'login'">Sign in</a></p>
          </div>
          <p v-if="activeTab === 'register'" class="terms">By clicking continue, you agree to our Terms of Service and Privacy Policy.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage, useDialog } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api'
import { Glasses, GlassesOutline, LogoGithub, UnlinkSharp } from '@vicons/ionicons5'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const dialog = useDialog()
const userStore = useUserStore()

const activeTab = ref('login')
const loginFormRef = ref(null)
const registerFormRef = ref(null)
const loginLoading = ref(false)
const registerLoading = ref(false)

const loginFormData = reactive({
  account: 'admin',
  password: 'password123'
})

const registerFormData = reactive({
  username: '',
  email: '',
  phone: '',
  password: '',
  confirmPassword: ''
})

const loginRules = {
  account: [{ required: true, message: 'Please enter your email', trigger: 'blur' }],
  password: [{ required: true, message: 'Please enter your password', trigger: 'blur' }]
}

const registerRules = {
  username: [{ required: true, message: 'Please enter your username', trigger: 'blur' }],
  email: [{ required: true, message: 'Please enter your email', trigger: 'blur' }],
  password: [{ required: true, message: 'Please enter your password', trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: 'Please confirm your password', trigger: 'blur' },
    {
      validator: (rule, value) => value === registerFormData.password,
      message: 'Passwords do not match',
      trigger: 'blur'
    }
  ]
}

function handleTabChange(value) {
  activeTab.value = value
}

async function handleLogin() {
  try {
    loginLoading.value = true
    const res = await authApi.login(loginFormData)
    userStore.setToken(res.access_token)
    userStore.setRefreshToken(res.refresh_token)
    userStore.setUserInfo(res.user)
    userStore.setRouteConfig(res.routes)
    userStore.setMenuConfig(res.menu)
    message.success('Login successful')
    
    const userRole = res.user?.role || 'client'
    const redirect = route.query.redirect || (userRole === 'admin' ? '/admin' : '/app')
    router.push(redirect)
  } catch (err) {
    if (err.response?.status === 401) {
      dialog.error({
        title: 'Login Failed',
        content: err.response.data?.detail || 'Invalid email or password',
        positiveText: 'OK'
      })
    } else {
      dialog.error({
        title: 'Login Failed',
        content: 'Network error, please try again later',
        positiveText: 'OK'
      })
    }
  } finally {
    loginLoading.value = false
  }
}

async function handleRegister() {
  try {
    registerLoading.value = true
    const { confirmPassword, ...data } = registerFormData
    await authApi.register(data)
    message.success('Account created successfully')
    activeTab.value = 'login'
    message.info('Please login with your new account')
  } catch (err) {
    console.error(err)
  } finally {
    registerLoading.value = false
  }
}
</script>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: flex;
  width: 100%;
}

.auth-sidebar {
  display: none;
  position: relative;
  flex: 1;
  background-color: #0f172a;
  overflow: hidden;
}

.auth-sidebar::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.9) 0%, rgba(15, 23, 42, 0.7) 50%, rgba(15, 23, 42, 0.8) 100%);
}

@media (min-width: 1024px) {
  .auth-sidebar {
    display: flex;
    flex-direction: column;
  }
}

.sidebar-bg {
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg width='100%25' height='100%25' xmlns='http://www.w3.org/2000/svg'%3E%3Cdefs%3E%3Cpattern id='grid' width='40' height='40' patternUnits='userSpaceOnUse'%3E%3Cpath d='M 40 0 L 0 0 0 40' fill='none' stroke='rgba(255,255,255,0.03)' stroke-width='1'/%3E%3C/pattern%3E%3C/defs%3E%3Crect width='100%25' height='100%25' fill='url(%23grid)'/%3E%3C/svg%3E");
}

.sidebar-content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 100%;
  padding: 2.5rem;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.brand-icon {
  width: 2.25rem;
  height: 2.25rem;
  background: linear-gradient(135deg, #ffffff 0%, #6d6d6d 100%);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.brand-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: white;
}

.testimonial {
  max-width: 320px;
}

.quote {
  font-size: 1rem;
  font-style: italic;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 1.25rem;
}

.author-info {
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding-top: 1rem;
}

.author-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: white;
  margin-bottom: 0.25rem;
}

.author-title {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.5);
}

.auth-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background-color: #f8fafc;
}

.auth-card-wrapper {
  width: 100%;
  max-width: 380px;
}

.auth-card {
  background: white;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
  padding: 2rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1);
}

.auth-tabs {
  margin-bottom: 1.5rem;
}

.auth-header {
  margin-bottom: 1.5rem;
}

.auth-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 0.375rem;
}

.auth-description {
  color: #64748b;
  font-size: 0.875rem;
}

.auth-form {
  margin-bottom: 1.25rem;
}

.form-field {
  margin-bottom: 1.25rem;
}

.form-label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #0f172a;
  margin-bottom: 0.375rem;
}

.forgot-link {
  font-size: 0.75rem;
  color: #18a058;
  text-decoration: none;
  font-weight: 500;
}

.forgot-link:hover {
  text-decoration: underline;
}

.submit-btn {
  margin-bottom: 1.25rem;
  height: 2.5rem;
}

.divider {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #e2e8f0;
}

.divider span {
  font-size: 0.75rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.social-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
  height: 2.5rem;
  background: white;
  border: 1px solid #e2e8f0;
  color: #0f172a;
}

.social-btn:hover {
  background: #f8fafc;
}

.social-icon {
  width: 1.25rem;
  height: 1.25rem;
  margin-right: 0.5rem;
}

.auth-footer {
  text-align: center;
  font-size: 0.875rem;
  color: #64748b;
}

.auth-footer a {
  color: #18a058;
  text-decoration: none;
  font-weight: 500;
}

.auth-footer a:hover {
  text-decoration: underline;
}

.terms {
  text-align: center;
  font-size: 0.75rem;
  color: #94a3b8;
  margin-top: 1rem;
  line-height: 1.5;
}
</style>
