<template>
  <div class="wf-create">
    <h2>创建 DAG</h2>
    <div class="toolbar">
      <router-link class="btn" to="/workflows">返回列表</router-link>
    </div>

    <div class="main-layout">
      <!-- 左侧表单 -->
      <div class="form-panel">
        <div class="form">
          <div class="form-row">
            <label>名称</label>
            <input v-model.trim="form.name" placeholder="例如：数据清洗流程" />
          </div>
          <div class="form-row">
            <label>描述</label>
            <input v-model.trim="form.desc" placeholder="简要描述" />
          </div>
          <div class="form-row">
            <label>自定义 ID（可选）</label>
            <input v-model.trim="form.id" placeholder="不填则自动生成，如 user-wf-1" />
          </div>

          <div class="form-row">
            <label>
              节点（JSON 数组）
              <span class="hint">字段：id, name, status[pending|running|success|failed], desc</span>
            </label>
            <textarea v-model="nodesText" rows="8" spellcheck="false" @input="updatePreview"></textarea>
            <div v-if="parseError.nodes" class="parse-error">{{ parseError.nodes }}</div>
          </div>

          <div class="form-row">
            <label>
              连边（JSON 数组）
              <span class="hint">字段：from, to, 可选 type="conditional", label</span>
            </label>
            <textarea v-model="edgesText" rows="6" spellcheck="false" @input="updatePreview"></textarea>
            <div v-if="parseError.edges" class="parse-error">{{ parseError.edges }}</div>
          </div>

          <div class="actions">
            <button :disabled="submitting" @click="submit">提交创建</button>
            <button class="secondary" @click="fillExample">填充示例</button>
          </div>

          <div v-if="error" class="error">{{ error }}</div>
          <div v-if="success" class="success">创建成功：{{ success }}，3 秒后跳转到 DAG 页面...</div>
        </div>
      </div>

      <!-- 右侧预览 -->
      <div class="preview-panel">
        <div class="preview-header">
          <h3>实时预览</h3>
          <div class="preview-stats" v-if="previewNodes.length">
            {{ previewNodes.length }} 个节点，{{ previewEdges.length }} 条边
          </div>
        </div>
        <div class="preview-container">
          <DAGRenderer 
            v-if="previewNodes.length" 
            :nodes="previewNodes" 
            :edges="previewEdges"
            :width="500"
            :height="400"
          />
          <div v-else class="preview-empty">
            <div class="empty-icon">📊</div>
            <div class="empty-text">请添加节点以查看 DAG 预览</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import DAGRenderer from './DAGRenderer.vue'

export default {
  name: 'WorkflowCreate',
  components: {
    DAGRenderer
  },
  data() {
    const exampleNodes = [
      { id: 'START', name: '数据接入', status: 'success', desc: '从多源系统拉取原始数据' },
      { id: 'CLEAN', name: '数据清洗', status: 'running', desc: '去重、缺失值处理' },
      { id: 'SPLIT', name: '数据分流', status: 'pending', desc: '按业务规则分流' },
      { id: 'FEAT_A', name: '特征工程A', status: 'pending', desc: '用户行为特征' },
      { id: 'FEAT_B', name: '特征工程B', status: 'pending', desc: '商品属性特征' },
      { id: 'MODEL', name: '模型训练', status: 'pending', desc: '集成多特征训练' },
      { id: 'VALID', name: '模型验证', status: 'pending', desc: '交叉验证评估' },
      { id: 'DEPLOY', name: '模型部署', status: 'pending', desc: '生产环境发布' }
    ]
    const exampleEdges = [
      { from: 'START', to: 'CLEAN' },
      { from: 'CLEAN', to: 'SPLIT' },
      { from: 'SPLIT', to: 'FEAT_A', type: 'conditional', label: '用户数据' },
      { from: 'SPLIT', to: 'FEAT_B', type: 'conditional', label: '商品数据' },
      { from: 'FEAT_A', to: 'MODEL' },
      { from: 'FEAT_B', to: 'MODEL' },
      { from: 'MODEL', to: 'VALID' },
      { from: 'VALID', to: 'DEPLOY', type: 'conditional', label: 'AUC>0.85' }
    ]
    return {
      form: { id: '', name: '', desc: '' },
      nodesText: JSON.stringify(exampleNodes, null, 2),
      edgesText: JSON.stringify(exampleEdges, null, 2),
      submitting: false,
      error: '',
      success: '',
      // 预览相关
      previewNodes: [...exampleNodes],
      previewEdges: [...exampleEdges],
      parseError: { nodes: '', edges: '' }
    }
  },
  mounted() {
    this.updatePreview()
  },
  methods: {
    updatePreview() {
      // 清除之前的错误
      this.parseError.nodes = ''
      this.parseError.edges = ''
      
      // 解析节点
      let nodes = []
      try {
        if (this.nodesText.trim()) {
          nodes = JSON.parse(this.nodesText)
          if (!Array.isArray(nodes)) {
            throw new Error('节点必须是数组格式')
          }
        }
      } catch (e) {
        this.parseError.nodes = 'JSON 解析错误: ' + e.message
        this.previewNodes = []
        this.previewEdges = []
        return
      }
      
      // 解析连边
      let edges = []
      try {
        if (this.edgesText.trim()) {
          edges = JSON.parse(this.edgesText)
          if (!Array.isArray(edges)) {
            throw new Error('连边必须是数组格式')
          }
        }
      } catch (e) {
        this.parseError.edges = 'JSON 解析错误: ' + e.message
        this.previewNodes = nodes
        this.previewEdges = []
        return
      }
      
      // 更新预览
      this.previewNodes = nodes
      this.previewEdges = edges
    },
    fillExample() {
      const nodes = [
        { id: 'INIT', name: '初始化', status: 'success', desc: '系统启动检查' },
        { id: 'AUTH', name: '身份验证', status: 'success', desc: '用户权限校验' },
        { id: 'LOAD', name: '数据加载', status: 'running', desc: '批量数据读取' },
        { id: 'PROC_A', name: '处理分支A', status: 'pending', desc: '实时数据处理' },
        { id: 'PROC_B', name: '处理分支B', status: 'pending', desc: '历史数据处理' },
        { id: 'MERGE', name: '数据合并', status: 'pending', desc: '多源数据融合' },
        { id: 'CHECK', name: '质量检查', status: 'pending', desc: '数据完整性验证' },
        { id: 'OUTPUT', name: '结果输出', status: 'pending', desc: '生成最终报告' }
      ]
      const edges = [
        { from: 'INIT', to: 'AUTH' },
        { from: 'AUTH', to: 'LOAD', type: 'conditional', label: '验证通过' },
        { from: 'LOAD', to: 'PROC_A', type: 'conditional', label: '实时流' },
        { from: 'LOAD', to: 'PROC_B', type: 'conditional', label: '批处理' },
        { from: 'PROC_A', to: 'MERGE' },
        { from: 'PROC_B', to: 'MERGE' },
        { from: 'MERGE', to: 'CHECK' },
        { from: 'CHECK', to: 'OUTPUT', type: 'conditional', label: '检查通过' }
      ]
      this.nodesText = JSON.stringify(nodes, null, 2)
      this.edgesText = JSON.stringify(edges, null, 2)
      this.updatePreview()
    },
    parseJSON(text, fallback) {
      if (!text || !text.trim()) return fallback
      try { return JSON.parse(text) } catch (e) { throw new Error('JSON 解析失败：' + e.message) }
    },
    async submit() {
      this.error = ''
      this.success = ''
      let nodes, edges
      try {
        nodes = this.parseJSON(this.nodesText, [])
        edges = this.parseJSON(this.edgesText, [])
      } catch (e) {
        this.error = e.message
        return
      }
      if (!Array.isArray(nodes) || nodes.length === 0) {
        this.error = '请至少提供 1 个节点'
        return
      }
      // 简单本地校验
      const ids = new Set()
      for (const n of nodes) {
        if (!n.id) { this.error = '存在缺少 id 的节点'; return }
        if (ids.has(n.id)) { this.error = '重复的节点 id：' + n.id; return }
        ids.add(n.id)
      }
      this.submitting = true
      try {
        const res = await fetch('/api/v1/workflows', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            id: this.form.id || undefined,
            name: this.form.name,
            desc: this.form.desc,
            nodes,
            edges
          })
        })
        if (!res.ok) {
          const t = await res.text()
          throw new Error('创建失败：' + res.status + ' ' + t)
        }
        const data = await res.json()
        const id = data.id
        this.success = id
        // 延迟跳转到新 DAG 页面
        setTimeout(() => {
          this.$router.push(`/workflow/${id}`)
        }, 1000 * 3)
      } catch (e) {
        this.error = e.message || '提交失败'
      } finally {
        this.submitting = false
      }
    }
  }
}
</script>

<style scoped>
.wf-create { 
  padding: 16px; 
  height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
}
.toolbar { margin: 8px 0 12px; }
.btn { display: inline-block; padding: 6px 10px; border: 1px solid #1976d2; color: #1976d2; border-radius: 6px; text-decoration: none; font-size: 13px; }
.btn:hover { background: #e3f2fd; }

.main-layout {
  display: flex;
  gap: 20px;
  flex: 1;
  min-height: 0;
}

.form-panel {
  flex: 0 0 450px;
  display: flex;
  flex-direction: column;
}

.form { 
  flex: 1;
  overflow-y: auto;
  padding-right: 8px;
}

.form-row { 
  margin-bottom: 12px; 
  display: flex; 
  flex-direction: column; 
  gap: 6px; 
}
.form-row label { font-weight: 600; color: #333; }
.form-row .hint { font-weight: 400; color: #777; margin-left: 8px; font-size: 12px; }
.form-row input { height: 32px; padding: 0 8px; border: 1px solid #ddd; border-radius: 6px; }
.form-row textarea { 
  padding: 8px; 
  border: 1px solid #ddd; 
  border-radius: 6px; 
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  resize: vertical;
  min-height: 120px;
}

.parse-error {
  color: #d32f2f;
  font-size: 12px;
  margin-top: 4px;
  padding: 4px 8px;
  background: #ffebee;
  border-radius: 4px;
  border-left: 3px solid #d32f2f;
}

.actions { display: flex; gap: 10px; margin-top: 8px; }
.actions button { 
  padding: 8px 14px; 
  border-radius: 6px; 
  border: 1px solid #1976d2; 
  color: #fff; 
  background: #1976d2; 
  cursor: pointer;
  font-size: 13px;
}
.actions button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.actions button.secondary { color: #1976d2; background: #fff; }
.actions button.secondary:hover { background: #e3f2fd; }

.error { 
  color: #d32f2f; 
  margin-top: 8px;
  padding: 8px;
  background: #ffebee;
  border-radius: 4px;
  border-left: 3px solid #d32f2f;
}
.success { 
  color: #2e7d32; 
  margin-top: 8px;
  padding: 8px;
  background: #e8f5e9;
  border-radius: 4px;
  border-left: 3px solid #2e7d32;
}

.preview-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 500px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
}

.preview-header h3 {
  margin: 0;
  color: #333;
  font-size: 16px;
}

.preview-stats {
  font-size: 12px;
  color: #666;
  background: #f5f5f5;
  padding: 4px 8px;
  border-radius: 12px;
}

.preview-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}

.preview-empty {
  text-align: center;
  color: #999;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
  opacity: 0.5;
}

.empty-text {
  font-size: 14px;
}

/* 响应式调整 */
@media (max-width: 1200px) {
  .main-layout {
    flex-direction: column;
  }
  
  .form-panel {
    flex: none;
  }
  
  .preview-panel {
    min-width: auto;
    min-height: 400px;
  }
}
</style>
