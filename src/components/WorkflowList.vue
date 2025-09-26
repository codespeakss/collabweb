<template>
  <div class="workflows">
    <h2>工作流列表</h2>
    <div class="toolbar">
      <button @click="fetchList">刷新</button>
    </div>
    <div v-if="loading" class="hint">加载中...</div>
    <div v-else>
      <div v-if="error" class="error">{{ error }}</div>
      <ul class="wf-list">
        <li v-for="wf in workflows" :key="wf.id" class="wf-item">
          <div class="wf-main">
            <div class="wf-name">{{ wf.name }}</div>
            <div class="wf-desc">{{ wf.desc }}</div>
          </div>
          <div class="wf-right">
            <span class="status" :class="wf.status">{{ wf.status }}</span>
            <router-link class="btn" :to="`/workflow/${wf.id}`">查看 DAG</router-link>
          </div>
        </li>
      </ul>
      <div v-if="!workflows.length" class="hint">暂无工作流</div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'WorkflowList',
  data() {
    return {
      workflows: [],
      loading: false,
      error: ''
    }
  },
  methods: {
    async fetchList() {
      this.loading = true
      this.error = ''
      try {
        const res = await fetch('/api/workflows')
        if (!res.ok) throw new Error('请求失败: ' + res.status)
        const data = await res.json()
        this.workflows = Array.isArray(data) ? data : []
      } catch (e) {
        this.error = e.message || '加载失败'
      } finally {
        this.loading = false
      }
    }
  },
  mounted() {
    this.fetchList()
  }
}
</script>

<style scoped>
.workflows { padding: 16px; }
.toolbar { margin: 8px 0 12px; }
.hint { color: #666; font-size: 14px; }
.error { color: #d32f2f; margin-bottom: 8px; }
.wf-list { list-style: none; padding: 0; margin: 0; }
.wf-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 10px;
  border: 1px solid #eee;
  border-radius: 8px;
  background: #fff;
  margin-bottom: 10px;
}
.wf-main { min-width: 0; }
.wf-name { font-weight: 600; margin-bottom: 4px; }
.wf-desc { color: #666; font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 60vw; }
.wf-right { display: flex; gap: 10px; align-items: center; }
.status { text-transform: lowercase; font-size: 12px; padding: 2px 8px; border-radius: 12px; border: 1px solid #ddd; }
.status.success { color: #2e7d32; background: #e8f5e9; border-color: #c8e6c9; }
.status.running { color: #1565c0; background: #e3f2fd; border-color: #bbdefb; }
.status.failed { color: #c62828; background: #ffebee; border-color: #ffcdd2; }
.status.pending { color: #8d6e63; background: #efebe9; border-color: #d7ccc8; }
.btn { display: inline-block; padding: 6px 10px; border: 1px solid #1976d2; color: #1976d2; border-radius: 6px; text-decoration: none; font-size: 13px; }
.btn:hover { background: #e3f2fd; }
</style>
