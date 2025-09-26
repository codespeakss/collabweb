<template>
  <div class="container">
    <h1 class="title">技术验证：打开指定 URL</h1>

    <form class="form" @submit.prevent="onOpen">
      <label for="url" class="label">目标 URL</label>
      <input
        id="url"
        v-model.trim="inputUrl"
        type="text"
        class="input"
        placeholder="例如：https://example.com 或 192.168.1.10:8080"
        @keyup.enter="onOpen"
      />

      <div class="actions">
        <button type="button" class="btn" @click="onOpen">在新标签页打开</button>
        <button type="button" class="btn secondary" @click="onClear">清空</button>
      </div>

      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="normalizedUrl" class="hint">将要打开：{{ normalizedUrl }}</p>
    </form>

    <details class="tips">
      <summary>使用说明</summary>
      <ul>
        <li>支持直接输入带协议的完整链接，如 https://example.com/path。</li>
        <li>若未填写协议，将自动补全为 http://。</li>
        <li>按回车或点击按钮即可在浏览器新标签页中打开。</li>
      </ul>
    </details>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'

const inputUrl = ref('')
const error = ref('')

const normalizedUrl = computed(() => {
  if (!inputUrl.value) return ''
  let url = inputUrl.value
  // 若没有协议，默认补 http://
  if (!/^https?:\/\//i.test(url)) {
    url = `http://${url}`
  }
  try {
    // 利用 URL 构造做一次基本校验
    const u = new URL(url)
    return u.toString()
  } catch (e) {
    return ''
  }
})

function onOpen() {
  error.value = ''
  if (!normalizedUrl.value) {
    error.value = '请输入合法的 URL（可省略协议）'
    return
  }
  // 在新标签打开
  window.open(normalizedUrl.value, '_blank', 'noopener,noreferrer')
}

function onClear() {
  inputUrl.value = ''
  error.value = ''
}
</script>

<style scoped>
.container {
  max-width: 720px;
  margin: 32px auto;
  padding: 16px;
}
.title {
  font-size: 22px;
  margin-bottom: 16px;
}
.form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.label {
  font-size: 14px;
  color: #555;
}
.input {
  padding: 10px 12px;
  border: 1px solid #ccc;
  border-radius: 6px;
  outline: none;
  font-size: 14px;
}
.input:focus {
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}
.actions {
  display: flex;
  gap: 10px;
}
.btn {
  background: #4f46e5;
  color: #fff;
  border: none;
  padding: 8px 14px;
  border-radius: 6px;
  cursor: pointer;
}
.btn:hover { opacity: 0.95; }
.btn.secondary { background: #64748b; }
.error { color: #dc2626; }
.hint { color: #16a34a; }
.tips { margin-top: 12px; }
</style>
