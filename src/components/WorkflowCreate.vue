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
        
        <!-- 编辑工具栏 -->
        <div class="edit-toolbar" v-if="previewNodes.length">
          <button class="tool-btn" @click="addNode" title="添加节点">
            <span>➕</span> 添加节点
          </button>
          <button class="tool-btn" :disabled="!selectedNodeId" @click="editSelectedNode" title="编辑选中节点">
            <span>✏️</span> 编辑
          </button>
          <button class="tool-btn" :class="{ active: connectingMode }" @click="toggleConnectMode" title="连线模式">
            <span>🔗</span> {{ connectingMode ? '退出连线' : '连线' }}
          </button>
          <button class="tool-btn danger" :disabled="!selectedNodeId" @click="deleteSelectedNode" title="删除选中节点">
            <span>🗑️</span> 删除
          </button>
          <button class="tool-btn" @click="clearAll" title="清空所有">
            <span>🗑️</span> 清空
          </button>
        </div>

        <!-- 连线操作引导 -->
        <div v-if="connectingMode" class="connect-guide">
          <div class="guide-content">
            <div class="guide-icon">🔗</div>
            <div class="guide-text">
              <div v-if="!connectingFrom" class="guide-step">
                <strong>步骤 1:</strong> 点击起始节点
              </div>
              <div v-else class="guide-step">
                <strong>步骤 2:</strong> 点击目标节点完成连线
                <div class="guide-from">起始节点: {{ getNodeName(connectingFrom) }}</div>
              </div>
            </div>
            <button class="guide-close" @click="toggleConnectMode" title="退出连线模式">✕</button>
          </div>
        </div>
        <div class="preview-container">
          <DAGRenderer 
            v-if="previewNodes.length" 
            :nodes="previewNodes" 
            :edges="previewEdges"
            :width="500"
            :height="400"
            :editable="true"
            :connecting-mode="connectingMode"
            :connecting-from="connectingFrom"
            @node-select="onNodeSelect"
            @node-edit="onNodeEdit"
            @canvas-click="onCanvasClick"
          />
          <div v-else class="preview-empty">
            <div class="empty-icon">📊</div>
            <div class="empty-text">请添加节点以查看 DAG 预览</div>
            <div class="empty-hint">
              <div>💡 操作提示：</div>
              <div>• 双击画布添加节点</div>
              <div>• 双击节点编辑属性</div>
              <div>• 点击连线按钮创建节点间的连接</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 节点编辑器模态框 -->
    <div v-if="showNodeEditor" class="modal-overlay" @click="cancelEdit">
      <div class="node-editor" @click.stop>
        <h3>编辑节点</h3>
        <div class="editor-form" v-if="editingNode">
          <div class="form-row">
            <label>节点 ID</label>
            <input v-model="editingNode.id" placeholder="节点唯一标识" />
          </div>
          <div class="form-row">
            <label>节点名称</label>
            <input v-model="editingNode.name" placeholder="显示名称" />
          </div>
          <div class="form-row">
            <label>状态</label>
            <select v-model="editingNode.status">
              <option value="pending">等待</option>
              <option value="running">运行中</option>
              <option value="success">成功</option>
              <option value="failed">失败</option>
            </select>
          </div>
          <div class="form-row">
            <label>描述</label>
            <textarea v-model="editingNode.desc" rows="3" placeholder="节点描述"></textarea>
          </div>
          <div class="editor-actions">
            <button class="primary" @click="updateNode">保存</button>
            <button class="danger" @click="deleteNode">删除</button>
            <button class="secondary" @click="cancelEdit">取消</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 连线提示浮窗 -->
    <div v-if="showConnectTip" class="connect-tip">
      <div class="tip-content">
        <span class="tip-icon">💡</span>
        {{ connectTipMessage }}
      </div>
    </div>

    <!-- 边编辑器模态框 -->
    <div v-if="showEdgeEditor" class="modal-overlay" @click="cancelEdgeEdit">
      <div class="edge-editor" @click.stop>
        <h3>编辑连边</h3>
        <div class="editor-form" v-if="editingEdge">
          <div class="form-row">
            <label>起始节点</label>
            <input :value="getNodeName(editingEdge.from)" readonly />
          </div>
          <div class="form-row">
            <label>目标节点</label>
            <input :value="getNodeName(editingEdge.to)" readonly />
          </div>
          <div class="form-row">
            <label>边类型</label>
            <select v-model="editingEdge.type">
              <option value="">普通边</option>
              <option value="conditional">条件边</option>
            </select>
          </div>
          <div class="form-row">
            <label>标签</label>
            <input v-model="editingEdge.label" placeholder="边的标签（可选）" />
          </div>
          <div class="editor-actions">
            <button class="primary" @click="updateEdge">保存</button>
            <button class="danger" @click="deleteEdge">删除</button>
            <button class="secondary" @click="cancelEdgeEdit">取消</button>
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
      parseError: { nodes: '', edges: '' },
      // 编辑相关
      selectedNodeId: null,
      editingNode: null,
      showNodeEditor: false,
      nodeSeq: 0,
      // 连线相关
      connectingMode: false,
      connectingFrom: null,
      showEdgeEditor: false,
      editingEdge: null,
      // 提示相关
      showConnectTip: false,
      connectTipMessage: ''
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
    },
    // 可视化编辑方法
    onNodeSelect(node) {
      if (this.connectingMode) {
        this.handleConnectNode(node)
      } else {
        this.selectedNodeId = node.id
      }
    },
    onNodeEdit(node) {
      this.editingNode = { ...node }
      this.showNodeEditor = true
    },
    onCanvasClick(position) {
      // 在画布上添加新节点
      this.nodeSeq++
      const newNode = {
        id: `N${this.nodeSeq}`,
        name: `节点${this.nodeSeq}`,
        status: 'pending',
        desc: '新建节点'
      }
      this.previewNodes.push(newNode)
      this.syncToJSON()
    },
    updateNode() {
      if (!this.editingNode) return
      const index = this.previewNodes.findIndex(n => n.id === this.editingNode.id)
      if (index !== -1) {
        this.previewNodes[index] = { ...this.editingNode }
        this.syncToJSON()
      }
      this.showNodeEditor = false
      this.editingNode = null
    },
    deleteNode() {
      if (!this.editingNode) return
      // 删除节点
      this.previewNodes = this.previewNodes.filter(n => n.id !== this.editingNode.id)
      // 删除相关的边
      this.previewEdges = this.previewEdges.filter(e => 
        e.from !== this.editingNode.id && e.to !== this.editingNode.id
      )
      this.syncToJSON()
      this.showNodeEditor = false
      this.editingNode = null
    },
    cancelEdit() {
      this.showNodeEditor = false
      this.editingNode = null
    },
    syncToJSON() {
      // 将可视化编辑的结果同步到 JSON 文本，过滤掉布局坐标
      const cleanNodes = this.previewNodes.map(node => {
        const { x, y, ...cleanNode } = node
        return cleanNode
      })
      this.nodesText = JSON.stringify(cleanNodes, null, 2)
      this.edgesText = JSON.stringify(this.previewEdges, null, 2)
    },
    // 工具栏方法
    addNode() {
      this.nodeSeq++
      const newNode = {
        id: `N${this.nodeSeq}`,
        name: `节点${this.nodeSeq}`,
        status: 'pending',
        desc: '新建节点'
      }
      this.previewNodes.push(newNode)
      this.syncToJSON()
    },
    editSelectedNode() {
      if (!this.selectedNodeId) return
      const node = this.previewNodes.find(n => n.id === this.selectedNodeId)
      if (node) {
        this.onNodeEdit(node)
      }
    },
    deleteSelectedNode() {
      if (!this.selectedNodeId) return
      // 删除节点
      this.previewNodes = this.previewNodes.filter(n => n.id !== this.selectedNodeId)
      // 删除相关的边
      this.previewEdges = this.previewEdges.filter(e => 
        e.from !== this.selectedNodeId && e.to !== this.selectedNodeId
      )
      this.selectedNodeId = null
      this.syncToJSON()
    },
    clearAll() {
      if (confirm('确定要清空所有节点和连边吗？')) {
        this.previewNodes = []
        this.previewEdges = []
        this.selectedNodeId = null
        this.nodeSeq = 0
        this.syncToJSON()
      }
    },
    // 连线相关方法
    toggleConnectMode() {
      this.connectingMode = !this.connectingMode
      this.connectingFrom = null
      if (this.connectingMode) {
        this.selectedNodeId = null
        this.showConnectTip('连线模式已激活，请先点击起始节点')
      } else {
        this.hideConnectTip()
      }
    },
    handleConnectNode(node) {
      if (!this.connectingFrom) {
        // 选择起始节点
        this.connectingFrom = node.id
        this.showConnectTip(`已选择起始节点 "${node.name}"，现在点击目标节点`)
      } else if (this.connectingFrom === node.id) {
        // 点击同一个节点，取消选择
        this.connectingFrom = null
        this.showConnectTip('已取消选择，请重新点击起始节点')
      } else {
        // 创建连边
        this.createEdge(this.connectingFrom, node.id)
        this.connectingFrom = null
        this.showConnectTip('连边创建成功！继续点击节点创建更多连边')
      }
    },
    createEdge(fromId, toId) {
      // 检查是否已存在相同的边
      const existingEdge = this.previewEdges.find(e => e.from === fromId && e.to === toId)
      if (existingEdge) {
        this.showConnectTip('该连边已存在，请选择其他节点', 'warning')
        return
      }
      
      // 创建新边
      const newEdge = {
        from: fromId,
        to: toId,
        type: '',
        label: ''
      }
      
      // 打开边编辑器
      this.editingEdge = newEdge
      this.showEdgeEditor = true
    },
    updateEdge() {
      if (!this.editingEdge) return
      
      // 添加或更新边
      const existingIndex = this.previewEdges.findIndex(e => 
        e.from === this.editingEdge.from && e.to === this.editingEdge.to
      )
      
      if (existingIndex !== -1) {
        this.previewEdges[existingIndex] = { ...this.editingEdge }
      } else {
        this.previewEdges.push({ ...this.editingEdge })
      }
      
      this.syncToJSON()
      this.showEdgeEditor = false
      this.editingEdge = null
    },
    deleteEdge() {
      if (!this.editingEdge) return
      
      this.previewEdges = this.previewEdges.filter(e => 
        !(e.from === this.editingEdge.from && e.to === this.editingEdge.to)
      )
      
      this.syncToJSON()
      this.showEdgeEditor = false
      this.editingEdge = null
    },
    cancelEdgeEdit() {
      this.showEdgeEditor = false
      this.editingEdge = null
    },
    getNodeName(nodeId) {
      const node = this.previewNodes.find(n => n.id === nodeId)
      return node ? `${node.name} (${node.id})` : nodeId
    },
    // 提示方法
    showConnectTip(message, type = 'info') {
      this.connectTipMessage = message
      this.showConnectTip = true
      // 3秒后自动隐藏
      setTimeout(() => {
        this.hideConnectTip()
      }, 3000)
    },
    hideConnectTip() {
      this.showConnectTip = false
      this.connectTipMessage = ''
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

.edit-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 6px;
  flex-wrap: wrap;
}

.tool-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.tool-btn:hover:not(:disabled) {
  background: #e3f2fd;
  border-color: #1976d2;
}

.tool-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.tool-btn.danger:hover:not(:disabled) {
  background: #ffebee;
  border-color: #d32f2f;
  color: #d32f2f;
}

.tool-btn span {
  font-size: 14px;
}

.tool-btn.active {
  background: #e3f2fd;
  border-color: #1976d2;
  color: #1976d2;
}

/* 连线引导样式 */
.connect-guide {
  margin-bottom: 12px;
  padding: 12px;
  background: linear-gradient(135deg, #e3f2fd 0%, #f3e5f5 100%);
  border: 1px solid #2196f3;
  border-radius: 8px;
  animation: guide-pulse 2s ease-in-out infinite;
}

.guide-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.guide-icon {
  font-size: 20px;
  animation: guide-bounce 1s ease-in-out infinite;
}

.guide-text {
  flex: 1;
}

.guide-step {
  font-size: 14px;
  color: #1976d2;
  font-weight: 500;
}

.guide-from {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
  padding: 4px 8px;
  background: rgba(33, 150, 243, 0.1);
  border-radius: 4px;
  display: inline-block;
}

.guide-close {
  background: none;
  border: none;
  font-size: 16px;
  color: #666;
  cursor: pointer;
  padding: 4px;
  border-radius: 50%;
  transition: all 0.2s;
}

.guide-close:hover {
  background: rgba(0, 0, 0, 0.1);
  color: #333;
}

@keyframes guide-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(33, 150, 243, 0.3); }
  50% { box-shadow: 0 0 0 8px rgba(33, 150, 243, 0.1); }
}

@keyframes guide-bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-3px); }
}

/* 连线提示浮窗 */
.connect-tip {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 1001;
  animation: tip-slide-in 0.3s ease-out;
}

.tip-content {
  background: #2196f3;
  color: white;
  padding: 12px 16px;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(33, 150, 243, 0.3);
  display: flex;
  align-items: center;
  gap: 8px;
  max-width: 300px;
  font-size: 14px;
}

.tip-icon {
  font-size: 16px;
}

@keyframes tip-slide-in {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
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

.empty-hint {
  font-size: 12px;
  color: #999;
  margin-top: 12px;
  text-align: left;
  line-height: 1.5;
}

.empty-hint > div:first-child {
  font-weight: 600;
  margin-bottom: 8px;
  color: #666;
}

.empty-hint > div:not(:first-child) {
  margin-bottom: 4px;
  padding-left: 8px;
}

/* 节点编辑器样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.node-editor,
.edge-editor {
  background: white;
  border-radius: 8px;
  padding: 24px;
  min-width: 400px;
  max-width: 500px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.15);
}

.node-editor h3,
.edge-editor h3 {
  margin: 0 0 20px 0;
  color: #333;
  font-size: 18px;
}

.editor-form .form-row {
  margin-bottom: 16px;
}

.editor-form label {
  display: block;
  margin-bottom: 6px;
  font-weight: 600;
  color: #333;
}

.editor-form input,
.editor-form select,
.editor-form textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  box-sizing: border-box;
}

.editor-form textarea {
  resize: vertical;
  min-height: 80px;
}

.editor-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.editor-actions button {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.editor-actions button.primary {
  background: #1976d2;
  color: white;
  border-color: #1976d2;
}

.editor-actions button.primary:hover {
  background: #1565c0;
}

.editor-actions button.danger {
  background: #d32f2f;
  color: white;
  border-color: #d32f2f;
}

.editor-actions button.danger:hover {
  background: #c62828;
}

.editor-actions button.secondary {
  background: white;
  color: #666;
  border-color: #ddd;
}

.editor-actions button.secondary:hover {
  background: #f5f5f5;
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
