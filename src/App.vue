<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { onStatusChange, getStatus } from './utils/serverStatus.js'

const online = ref(true)
const connected = ref(true)

let unsubscribe = null
onMounted(() => {
  const s = getStatus()
  online.value = s.online
  connected.value = s.connected
  unsubscribe = onStatusChange((st) => {
    online.value = st.online
    connected.value = st.connected
  })
})
onBeforeUnmount(() => {
  if (unsubscribe) unsubscribe()
})
</script>


<template>
  <div v-if="!online || !connected" class="conn-banner" role="alert">
    <div class="container banner-inner">
      <span v-if="!online">当前处于离线状态，功能受限。</span>
      <span v-else-if="!connected">正在尝试重新连接到服务器…</span>
    </div>
  </div>
  <nav class="top-nav">
    <div class="container nav-inner">
      <router-link to="/">首页</router-link>
      <span class="sep">|</span>
      <router-link to="/devices">设备列表</router-link>
      <span class="sep">|</span>
      <router-link to="/workflows">工作流</router-link>
      <span class="sep">|</span>
      <router-link to="/auth">登录/注册</router-link>
      <span class="sep">|</span>
      <router-link to="/about">关于</router-link>
      <span class="sep">|</span>
      <router-link to="/tech-validate">技术验证</router-link>
    </div>
  </nav>
  <main class="container">
    <router-view />
  </main>
</template>

<style scoped>
.conn-banner {
  position: sticky;
  top: 0;
  z-index: 110;
  width: 100%;
  background: #ffe9e9;
  color: #b00020;
  border-bottom: 1px solid #f5c2c2;
}
.banner-inner { padding: 8px 0; font-size: 14px; }
.logo {
  height: 6em;
  padding: 1.5em;
  will-change: filter;
  transition: filter 300ms;
}
.logo:hover {
  filter: drop-shadow(0 0 2em #646cffaa);
}
.logo.vue:hover {
  filter: drop-shadow(0 0 2em #42b883aa);
}

/* stable top navigation */
.top-nav {
  position: sticky;
  top: 0;
  z-index: 100;
  width: 100%;
  backdrop-filter: saturate(180%) blur(8px);
  -webkit-backdrop-filter: saturate(180%) blur(8px);
  margin-bottom: 1rem;
  border-bottom: 1px solid #e5e5e5;
  background-color: rgba(255, 255, 255, 0.9);
}

.nav-inner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  flex-wrap: wrap;
}

.top-nav a {
  color: inherit;
}

.top-nav a.router-link-active {
  font-weight: 600;
  color: #1976d2;
}

/* Dark mode header background */
@media (prefers-color-scheme: dark) {
  .top-nav {
    background-color: rgba(36, 36, 36, 0.9);
    border-bottom-color: #333;
  }
  .conn-banner { background: #3a2222; color: #ffb4ab; border-bottom-color: #6b2a2a; }
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .nav-inner {
    gap: 8px 10px;
  }
  .nav-inner .sep {
    display: none;
  }
}
</style>
