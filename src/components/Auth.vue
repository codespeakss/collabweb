<script setup>
import { ref, computed } from 'vue'

// tabs: login | register
const activeTab = ref('login')

// shared fields
const emailOrPhone = ref('')
const password = ref('')
const verifyCode = ref('')

// register-only fields
const agree = ref(true)

// state
const loading = ref(false)
const message = ref('')
const messageType = ref('') // success | error | info

const isEmail = computed(() => /.+@.+\..+/.test(emailOrPhone.value))

function setMsg(type, msg) {
  messageType.value = type
  message.value = msg
}

async function sendCode() {
  if (!emailOrPhone.value || !isEmail.value) {
    setMsg('error', '请先输入有效邮箱')
    return
  }
  loading.value = true
  try {
    const res = await fetch('/api/auth/send-code', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ account: emailOrPhone.value, channel: 'email' })
    })
    const data = await res.json()
    if (res.ok) setMsg('success', data.message || '验证码已发送')
    else setMsg('error', data.error || '发送失败')
  } catch (e) {
    setMsg('error', '网络错误，请稍后重试')
  } finally {
    loading.value = false
  }
}

async function doLogin() {
  if (!emailOrPhone.value || !password.value) {
    setMsg('error', '请输入邮箱与密码')
    return
  }
  if (!isEmail.value) {
    setMsg('error', '请输入有效邮箱')
    return
  }
  loading.value = true
  try {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ account: emailOrPhone.value, password: password.value })
    })
    const data = await res.json()
    if (res.ok) {
      setMsg('success', '登录成功，正在跳转...')
      setTimeout(() => {
        window.location.href = '/'
      }, 800)
    } else {
      setMsg('error', data.error || '登录失败')
    }
  } catch (e) {
    setMsg('error', '网络错误，请稍后重试')
  } finally {
    loading.value = false
  }
}

async function doRegister() {
  if (!emailOrPhone.value || !password.value || !verifyCode.value) {
    setMsg('error', '请输入完整信息：邮箱、密码与验证码')
    return
  }
  if (!isEmail.value) {
    setMsg('error', '请输入有效邮箱')
    return
  }
  if (!agree.value) {
    setMsg('error', '请先同意服务条款')
    return
  }
  loading.value = true
  try {
    const res = await fetch('/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ account: emailOrPhone.value, password: password.value, code: verifyCode.value })
    })
    const data = await res.json()
    if (res.ok) {
      setMsg('success', '注册成功，请登录')
      activeTab.value = 'login'
    } else {
      setMsg('error', data.error || '注册失败')
    }
  } catch (e) {
    setMsg('error', '网络错误，请稍后重试')
  } finally {
    loading.value = false
  }
}

function onTabChange(tab) {
  activeTab.value = tab
  setMsg('', '')
}
</script>

<template>
  <div class="auth-wrapper">
    <div class="card">
      <div class="tabs">
        <button :class="{ active: activeTab === 'login' }" @click="onTabChange('login')">登录</button>
        <button :class="{ active: activeTab === 'register' }" @click="onTabChange('register')">注册</button>
      </div>

      <div v-if="message" :class="['msg', messageType]">{{ message }}</div>

      <div v-if="activeTab === 'login'" class="tab-panel">
        <div class="row">
          <label>邮箱</label>
          <input v-model="emailOrPhone" placeholder="example@mail.com" />
        </div>
        <div class="row">
          <label>密码</label>
          <input v-model="password" type="password" placeholder="请输入密码" />
        </div>
        
        <div class="actions">
          <button class="primary" @click="doLogin" :disabled="loading">登录</button>
        </div>
      </div>

      <div v-else class="tab-panel">
        <div class="row">
          <label>邮箱</label>
          <input v-model="emailOrPhone" placeholder="example@mail.com" />
        </div>
        <div class="row">
          <label>密码</label>
          <input v-model="password" type="password" placeholder="至少 6 位" />
        </div>
        <div class="row two-cols">
          <div>
            <label>验证码</label>
            <input v-model="verifyCode" placeholder="邮箱验证码" />
          </div>
          <div class="send-code">
            <button @click="sendCode" :disabled="loading">发送验证码</button>
          </div>
        </div>
        <div class="row inline">
          <label><input type="checkbox" v-model="agree" /> 我已阅读并同意《服务条款》</label>
        </div>
        <div class="actions">
          <button class="primary" @click="doRegister" :disabled="loading">注册</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-wrapper {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 40px;
}
.card {
  width: 100%;
  max-width: 560px;
  border: 1px solid var(--vp-c-divider, #e5e5e5);
  border-radius: 12px;
  padding: 20px 20px 8px;
  background: var(--vp-c-bg, #fff);
  box-shadow: 0 2px 8px rgba(0,0,0,.06);
}
.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}
.tabs button {
  flex: 1;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  background: #fafafa;
}
.tabs button.active {
  background: #1976d2;
  color: white;
  border-color: #1976d2;
}
.msg { margin: 8px 0 4px; font-size: 14px; }
.msg.success { color: #2e7d32; }
.msg.error { color: #d32f2f; }

.tab-panel { margin-top: 8px; }
.row { display: flex; flex-direction: column; gap: 6px; margin: 10px 0; }
.row.inline { flex-direction: row; align-items: center; gap: 10px; }
.row.two-cols { display: grid; grid-template-columns: 1fr 140px; gap: 10px; align-items: end; }
label { font-size: 14px; color: #555; }
input { padding: 10px 12px; border: 1px solid #ddd; border-radius: 8px; outline: none; }
input:focus { border-color: #1976d2; }
.actions { margin-top: 12px; }
button.primary { background: #1976d2; color: #fff; border: none; padding: 10px 16px; border-radius: 8px; }
button[disabled] { opacity: .6; cursor: not-allowed; }

@media (prefers-color-scheme: dark) {
  .card { background: #2a2a2a; border-color: #3a3a3a; }
  input { background: #1e1e1e; color: #eaeaea; border-color: #444; }
}
</style>
