<template>
  <div class="auth-layout">
    <header class="auth-topbar">
      <router-link to="/login" class="auth-brand" aria-label="mixm 首页">
        <BrandMark />
        <span>mixm<span class="brand-period">.</span></span>
      </router-link>
      <span class="topbar-caption">一个学习、实践与探索的空间</span>
      <span class="eyebrow topbar-label">A SPACE FOR CURIOSITY</span>
    </header>

    <div class="auth-body">
      <section class="auth-story" aria-labelledby="story-title">
        <div class="story-copy">
          <p class="eyebrow story-eyebrow"><span></span> LEARN. EXPERIMENT. GROW.</p>
          <h1 id="story-title" class="editorial-title">让好奇心，<br />成为你的下一步。</h1>
          <p class="story-description">
            从一个问题开始，在知识与实践之间，<br class="desktop-break" />发现更多可能。
          </p>
        </div>
        <ExplorationArt class="story-art" />
        <div class="story-bottom">
          <span>连接知识，让想法发生。</span>
          <span class="eyebrow">01 — 无限可能</span>
        </div>
      </section>

      <main class="auth-main">
        <div class="auth-card">
          <p class="eyebrow form-eyebrow">YOUR NEXT CHAPTER</p>
          <n-tabs :value="activeTab" type="line" @update:value="handleTabChange" class="auth-tabs">
            <n-tab-pane name="login" tab="登录">
              <div class="auth-heading">
                <h2 class="editorial-title">欢迎回来。</h2>
                <p>登录 mixm，继续你的学习与探索。</p>
              </div>
              <n-form
                ref="loginFormRef"
                :model="loginFormData"
                :rules="loginRules"
                :show-require-mark="false"
                @submit.prevent="handleLogin"
              >
                <n-form-item label="邮箱" path="email">
                  <n-input
                    v-model:value="loginFormData.email"
                    size="large"
                    placeholder="you@example.com"
                    :input-props="{ 'aria-label': '邮箱', autocomplete: 'username' }"
                  />
                </n-form-item>
                <n-form-item label="密码" path="password">
                  <n-input
                    v-model:value="loginFormData.password"
                    type="password"
                    size="large"
                    show-password-on="click"
                    placeholder="输入你的密码"
                    :input-props="{ 'aria-label': '密码', autocomplete: 'current-password' }"
                  />
                </n-form-item>
                <n-button
                  attr-type="submit"
                  type="primary"
                  block
                  size="large"
                  :loading="loginLoading"
                  class="submit-btn"
                >
                  进入学习空间 <span aria-hidden="true" class="button-arrow">↗</span>
                </n-button>
              </n-form>
              <p class="account-note">
                还没有账号？<button type="button" @click="handleTabChange('register')">
                  了解账号注册 <span aria-hidden="true">↗</span>
                </button>
              </p>
            </n-tab-pane>
            <n-tab-pane name="register" tab="创建账号">
              <div class="auth-heading">
                <h2 class="editorial-title">开启新一章。</h2>
                <p>自主注册暂未开放，请联系管理员获取账号。</p>
              </div>
              <n-form
                ref="registerFormRef"
                :model="registerFormData"
                :rules="registerRules"
                :show-require-mark="false"
                @submit.prevent="handleRegister"
              >
                <n-form-item label="用户名" path="username">
                  <n-input
                    v-model:value="registerFormData.username"
                    placeholder="你的名字"
                    :input-props="{ 'aria-label': '用户名', autocomplete: 'username' }"
                  />
                </n-form-item>
                <n-form-item label="邮箱" path="email">
                  <n-input
                    v-model:value="registerFormData.email"
                    placeholder="you@example.com"
                    :input-props="{ 'aria-label': '注册邮箱', autocomplete: 'email' }"
                  />
                </n-form-item>
                <n-form-item label="手机号" path="phone">
                  <n-input
                    v-model:value="registerFormData.phone"
                    placeholder="输入手机号"
                    :input-props="{ 'aria-label': '手机号', autocomplete: 'tel' }"
                  />
                </n-form-item>
                <n-form-item label="密码" path="password">
                  <n-input
                    v-model:value="registerFormData.password"
                    type="password"
                    show-password-on="click"
                    placeholder="设置密码"
                    :input-props="{ 'aria-label': '注册密码', autocomplete: 'new-password' }"
                  />
                </n-form-item>
                <n-form-item label="确认密码" path="confirmPassword">
                  <n-input
                    v-model:value="registerFormData.confirmPassword"
                    type="password"
                    show-password-on="click"
                    placeholder="再次输入密码"
                    :input-props="{ 'aria-label': '确认密码', autocomplete: 'new-password' }"
                  />
                </n-form-item>
                <n-button
                  attr-type="submit"
                  type="primary"
                  block
                  size="large"
                  :loading="registerLoading"
                  class="submit-btn"
                  >创建账号 <span aria-hidden="true" class="button-arrow">↗</span></n-button
                >
              </n-form>
            </n-tab-pane>
          </n-tabs>
          <div class="form-footnote">
            <span class="footnote-line"></span>为专注的学习，与自由的探索而设计。
          </div>
        </div>
      </main>
    </div>

    <footer class="auth-footer">
      <span>mixm · 教学与实验平台</span>
      <span>Ecust CCIOT &amp; Yoobit.cn</span>
    </footer>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage, useDialog } from 'naive-ui'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api'
import BrandMark from '@/components/BrandMark.vue'
import ExplorationArt from '@/components/ExplorationArt.vue'

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
  email: '',
  password: ''
})

const registerFormData = reactive({
  username: '',
  email: '',
  phone: '',
  password: '',
  confirmPassword: ''
})

const loginRules = {
  email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const registerRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    {
      validator: (rule, value) => value === registerFormData.password,
      message: '两次输入的密码不一致',
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
    message.success('登录成功')

    const userRole = res.user?.role || 'client'
    const redirect = route.query.redirect || (userRole === 'admin' ? '/admin' : '/app')
    router.push(redirect)
  } catch (err) {
    if (err.response?.status === 401) {
      dialog.error({
        title: '登录失败',
        content: err.response.data?.message || '邮箱或密码不正确',
        positiveText: '确定'
      })
    } else {
      dialog.error({
        title: '登录失败',
        content: '连接失败，请稍后重试',
        positiveText: '确定'
      })
    }
  } finally {
    loginLoading.value = false
  }
}

async function handleRegister() {
  message.warning('暂未开放自主注册，请联系管理员创建账号')
  return

  try {
    registerLoading.value = true
    const { confirmPassword, ...data } = registerFormData
    await authApi.register(data)
    message.success('账号创建成功')
    activeTab.value = 'login'
    message.info('请使用新账号登录')
  } catch (err) {
    console.error(err)
  } finally {
    registerLoading.value = false
  }
}
</script>

<style scoped>
.auth-layout {
  min-height: 100svh;
  display: flex;
  flex-direction: column;
  background: var(--mixm-paper);
}
.auth-topbar {
  min-height: 90px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin: 0 5vw;
  border-bottom: 1px solid var(--mixm-border);
}
.auth-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: var(--mixm-ink);
}
.auth-brand > span {
  font-size: 34px;
  font-weight: 600;
  letter-spacing: -0.065em;
  line-height: 1;
}
.brand-period {
  color: var(--mixm-accent);
}
.topbar-caption {
  margin-right: auto;
  margin-left: 28px;
  font-size: 12px;
  color: var(--mixm-muted);
}
.topbar-label {
  font-size: 10px;
}
.auth-body {
  display: grid;
  grid-template-columns: 1.12fr 1fr;
  flex: 1;
  width: 100%;
  max-width: 1600px;
  margin: 0 auto;
}
.auth-story {
  min-width: 0;
  padding: 66px 5vw 32px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.story-eyebrow {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 28px;
  font-size: 10px;
}
.story-eyebrow span {
  width: 7px;
  height: 7px;
  background: var(--mixm-accent);
  border-radius: 50%;
}
.story-copy h1 {
  font-size: clamp(42px, 4.6vw, 72px);
  line-height: 1.3;
  margin-bottom: 24px;
}
.story-description {
  color: var(--mixm-muted);
  font-size: 15px;
  line-height: 1.9;
}
.story-art {
  width: min(100%, 445px);
  align-self: center;
  margin: 20px 0 6px;
}
.story-bottom {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding-top: 24px;
  border-top: 1px solid var(--mixm-border);
  font-size: 11px;
  color: var(--mixm-muted);
}
.story-bottom .eyebrow {
  font-size: 9px;
}
.auth-main {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 56px 5vw;
  border-left: 1px solid var(--mixm-border);
  background: var(--mixm-surface);
}
.auth-card {
  width: 100%;
  max-width: 370px;
}
.form-eyebrow {
  margin-bottom: 20px;
  color: var(--mixm-accent);
  font-size: 10px;
}
.auth-heading {
  margin: 30px 0 32px;
}
.auth-heading h2 {
  font-size: 38px;
  margin-bottom: 12px;
}
.auth-heading p {
  font-size: 13px;
  line-height: 1.8;
  color: var(--mixm-muted);
}
.auth-tabs :deep(.n-tabs-tab) {
  font-size: 13px;
  padding-bottom: 12px;
}
.auth-tabs :deep(.n-tab-pane) {
  padding-top: 0;
}
.auth-card :deep(.n-form-item-label) {
  font-size: 12px;
}
.auth-card :deep(.n-input) {
  background: transparent;
}
.submit-btn {
  height: 48px;
  margin-top: 8px;
}
.button-arrow {
  margin-left: 14px;
  font-size: 20px;
  font-weight: 400;
}
.account-note {
  margin: 24px 0 0;
  text-align: center;
  font-size: 12px;
  color: var(--mixm-muted);
}
.account-note button {
  padding: 0;
  background: none;
  border: 0;
  color: var(--mixm-accent);
  cursor: pointer;
  margin-left: 6px;
}
.account-note button:hover {
  text-decoration: underline;
  text-underline-offset: 4px;
}
.form-footnote {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 50px;
  color: var(--mixm-muted);
  font-size: 10px;
}
.footnote-line {
  width: 26px;
  height: 1px;
  background: var(--mixm-accent);
  flex-shrink: 0;
}
.auth-footer {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 5vw;
  border-top: 1px solid var(--mixm-border);
  font-size: 10px;
  color: var(--mixm-muted);
  letter-spacing: 0.03em;
}
@media (min-width: 1800px) {
  .auth-story {
    padding-left: 32px;
  }
}
@media (max-width: 900px) {
  .auth-topbar {
    min-height: 76px;
  }
  .topbar-caption {
    display: none;
  }
  .auth-story {
    padding: 48px 32px 24px;
  }
  .auth-main {
    padding: 48px 32px;
  }
  .story-copy h1 {
    font-size: 42px;
  }
  .story-bottom {
    flex-direction: column;
    gap: 6px;
  }
}
@media (max-width: 680px) {
  .auth-topbar {
    margin-inline: 24px;
    min-height: 72px;
  }
  .auth-brand > span {
    font-size: 29px;
  }
  .topbar-label {
    font-size: 8px;
    letter-spacing: 0.08em;
  }
  .auth-body {
    grid-template-columns: 1fr;
  }
  .auth-story {
    padding: 36px 24px 30px;
  }
  .story-eyebrow {
    margin-bottom: 18px;
    font-size: 9px;
  }
  .story-copy h1 {
    font-size: 38px;
    margin-bottom: 14px;
  }
  .story-copy h1 br {
    display: none;
  }
  .story-description {
    font-size: 13px;
  }
  .desktop-break,
  .story-art,
  .story-bottom {
    display: none;
  }
  .auth-main {
    border-left: 0;
    border-top: 1px solid var(--mixm-border);
    padding: 32px 24px 38px;
  }
  .form-eyebrow {
    display: none;
  }
  .auth-heading {
    margin: 24px 0 28px;
  }
  .auth-heading h2 {
    font-size: 30px;
  }
  .form-footnote {
    margin-top: 30px;
  }
  .auth-footer {
    padding: 18px 24px;
    flex-wrap: wrap;
    font-size: 9px;
  }
}
</style>
