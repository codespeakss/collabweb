<template>
  <div class="workflow-dag" ref="dagContainer">
    <h2>Workflow DAG 状态展示</h2>
    <div class="toolbar">
      <button @click="resetDAG">重置</button>
      <button @click="stepDAG">下一步</button>
      <button @click="randomizeStatuses">随机状态</button>
      <button @click="fetchWorkflow" title="从后端刷新">刷新</button>
      <span class="sep"></span>
      <label class="ctrl"><input type="checkbox" v-model="useAutoSize" /> 自动尺寸</label>
      <template v-if="!useAutoSize">
        <label class="ctrl">宽度 <input type="number" v-model.number="userWidth" min="400" step="50" style="width:90px" /></label>
        <label class="ctrl">高度 <input type="number" v-model.number="userHeight" min="300" step="50" style="width:90px" /></label>
      </template>
    </div>
    <svg ref="svgEl" :width="svgAttrWidth" :height="svgAttrHeight" :viewBox="`0 0 ${svgViewWidth} ${svgViewHeight}`">
      <g v-for="(node, idx) in nodes" :key="node.id">
        <!-- 节点圆形 -->
          <foreignObject :x="posX(node) - nodeWidth / 2" :y="posY(node) - nodeHeight / 2" :width="nodeWidth" :height="nodeHeight">
            <div
              class="dag-node"
              :class="{ 'is-running': node.status === 'running' }"
              :style="{ borderColor: statusColor(node.status), width: nodeWidth + 'px', height: nodeHeight + 'px' }"
              @click="selectNode(node)"
              @mouseenter="onNodeEnter(node, $event)"
              @mousemove="onNodeMove($event)"
              @mouseleave="onNodeLeave"
            >
              <span class="dag-node-dot" :style="{ background: statusColor(node.status) }"></span>
              <div class="dag-node-title">{{ node.name }}</div>
              <div class="dag-node-status" :style="{ color: statusColor(node.status) }">{{ node.status }}</div>
            </div>
          </foreignObject>
      </g>
      <!-- 连线 -->
      <g v-for="edge in edges" :key="edge.from + '-' + edge.to">
        <path
          :d="edgePath(edge)"
          fill="none"
          :stroke="edgeColor(edge)"
          stroke-width="2"
          :marker-end="edgeMarker(edge)"
          :stroke-dasharray="edge.type === 'conditional' ? '6,6' : '0'"
        />
        <text v-if="edge.label" :x="edgeLabelPos(edge).x" :y="edgeLabelPos(edge).y" text-anchor="middle" class="edge-label">
          {{ edge.label }}
        </text>
      </g>
      <defs>
        <marker id="arrow-green" markerWidth="6" markerHeight="6" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
          <path d="M0,0 L6,3 L0,6 Z" fill="#4caf50" />
        </marker>
        <marker id="arrow-gray" markerWidth="6" markerHeight="6" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
          <path d="M0,0 L6,3 L0,6 Z" fill="#bdbdbd" />
        </marker>
        <marker id="arrow-red" markerWidth="6" markerHeight="6" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
          <path d="M0,0 L6,3 L0,6 Z" fill="#f44336" />
        </marker>
      </defs>
      <!-- 内嵌于 SVG 的全屏控制按钮 -->
      <g class="svg-fs-btn" @click="toggleFullscreen" style="cursor: pointer; user-select: none;" :transform="`translate(${svgViewWidth - 80}, 8)`">
        <rect rx="6" ry="6" width="72" height="26" fill="none" />
        <text x="36" y="17" text-anchor="middle" font-size="12">{{ isFullscreen ? '退出全屏' : '全屏' }}</text>
      </g>
    </svg>
    <div class="legend">
      <div class="legend-item"><span class="swatch" :style="{background: statusColor('success')}"></span>成功</div>
      <div class="legend-item"><span class="swatch" :style="{background: statusColor('running')}"></span>运行中</div>
      <div class="legend-item"><span class="swatch" :style="{background: statusColor('failed')}"></span>失败</div>
      <div class="legend-item"><span class="swatch" :style="{background: statusColor('pending')}"></span>等待</div>
      <div class="legend-item"><span class="swatch dashed"></span>条件边</div>
    </div>
    <!-- 悬浮提示：鼠标移入节点时展示 -->
    <div
      v-if="hoverNode"
      class="node-tooltip"
      :style="{ left: tooltipX + 'px', top: tooltipY + 'px' }"
    >
      <div class="node-tooltip-title">{{ hoverNode.name }}</div>
      <div class="node-tooltip-line">
        <span class="label">状态：</span>
        <span :style="{ color: statusColor(hoverNode.status) }">{{ hoverNode.status }}</span>
      </div>
      <div class="node-tooltip-line"><span class="label">描述：</span>{{ hoverNode.desc }}</div>
    </div>
    <div v-if="selectedNode" class="node-detail">
      <h3>任务详情：{{ selectedNode.name }}</h3>
      <p>状态：{{ selectedNode.status }}</p>
      <p>描述：{{ selectedNode.desc }}</p>
      <button @click="selectedNode = null">关闭</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'WorkflowDAG',
  data() {
    return {
      svgWidth: 780,
      svgHeight: 560,
      nodeRadius: 28,
      nodeWidth: 80,
      nodeHeight: 72,
      selectedNode: null,
      hoverNode: null,
      tooltipX: 0,
      tooltipY: 0,
      // 让布局更舒展的缩放与边距
      layoutScaleX: 1.25,
      layoutScaleY: 1.20,
      canvasPadding: 60,
      // UI 宽高控制
      useAutoSize: true,
      userWidth: 900,
      userHeight: 600,
      // 后端返回的 DAG 数据
      nodes: [],
      edges: [],
      // 全屏状态
      isFullscreen: false,
    };
  },
  methods: {
    async fetchWorkflow() {
      try {
        const res = await fetch('/api/workflow');
        if (!res.ok) throw new Error('请求失败: ' + res.status);
        const data = await res.json();
        // 直接赋值后端 nodes/edges
        this.nodes = Array.isArray(data.nodes) ? data.nodes : [];
        this.edges = Array.isArray(data.edges) ? data.edges : [];
      } catch (err) {
        console.error('加载工作流失败', err);
      }
    },
    // 布局缩放后的坐标
    posX(node) { return node.x * this.layoutScaleX + this.canvasPadding },
    posY(node) { return node.y * this.layoutScaleY + this.canvasPadding },
    findNode(id) {
      return this.nodes.find(n => n.id === id);
    },
    statusColor(status) {
      switch (status) {
        case 'success': return '#4caf50';
        case 'running': return '#2196f3';
        case 'failed': return '#f44336';
        case 'pending': return '#ffc107';
        default: return '#bdbdbd';
      }
    },
    // 判断一条边是否已被遍历：这里以“目标节点不为 pending”作为已遍历的判定依据
    isEdgeTraversed(edge) {
      const to = this.findNode(edge.to)
      return to ? to.status !== 'pending' : false
    },
    // 根据状态返回边颜色：优先目标失败 -> 红；否则源成功 -> 绿；否则灰
    edgeColor(edge) {
      const from = this.findNode(edge.from)
      const to = this.findNode(edge.to)
      if (to && to.status === 'failed') return '#f44336'
      if (from && from.status === 'success') return '#4caf50'
      return '#bdbdbd'
    },
    // 返回箭头 marker（与颜色一致）
    edgeMarker(edge) {
      const from = this.findNode(edge.from)
      const to = this.findNode(edge.to)
      if (to && to.status === 'failed') return 'url(#arrow-red)'
      if (from && from.status === 'success') return 'url(#arrow-green)'
      return 'url(#arrow-gray)'
    },
    // 动态选择端口：横向用左右侧，纵向用上下侧，斜向则混合
    edgePorts(from, to) {
      const fx = this.posX(from), fy = this.posY(from)
      const tx = this.posX(to), ty = this.posY(to)
      const dx = tx - fx
      const dy = ty - fy
      const horizThreshold = this.nodeHeight * 0.3
      // 端口：中心点基础上向外偏移到节点边缘
      const left   = { x: fx - this.nodeWidth / 2, y: fy }
      const right  = { x: fx + this.nodeWidth / 2, y: fy }
      const top    = { x: fx, y: fy - this.nodeHeight / 2 }
      const bottom = { x: fx, y: fy + this.nodeHeight / 2 }
      const toLeft   = { x: tx - this.nodeWidth / 2, y: ty }
      const toRight  = { x: tx + this.nodeWidth / 2, y: ty }
      const toTop    = { x: tx, y: ty - this.nodeHeight / 2 }
      const toBottom = { x: tx, y: ty + this.nodeHeight / 2 }
      if (Math.abs(dy) <= horizThreshold) {
        // 基本横向
        if (dx >= 0) return { p1: right, p2: toLeft }
        return { p1: left, p2: toRight }
      }
      // 明显下游
      if (dy > 0) return { p1: bottom, p2: toTop }
      // 明显上游（不常见）
      return { p1: top, p2: toBottom }
    },
    // 生成贝塞尔曲线路径，提升可读性
    edgePath(edge) {
      const from = this.findNode(edge.from)
      const to = this.findNode(edge.to)
      if (!from || !to) return ''
      const { p1, p2 } = this.edgePorts(from, to)
      const dx = p2.x - p1.x
      const dy = p2.y - p1.y
      // 控制点偏移：根据方向在 x 或 y 方向拉开一定弧度
      const offset = 40
      let c1 = { x: p1.x, y: p1.y }
      let c2 = { x: p2.x, y: p2.y }
      if (Math.abs(dy) <= this.nodeHeight * 0.3) {
        // 横向：在 x 方向拉开
        const s = dx >= 0 ? 1 : -1
        c1 = { x: p1.x + s * offset, y: p1.y }
        c2 = { x: p2.x - s * offset, y: p2.y }
      } else {
        // 纵向：在 y 方向拉开
        const s = dy >= 0 ? 1 : -1
        c1 = { x: p1.x, y: p1.y + s * offset }
        c2 = { x: p2.x, y: p2.y - s * offset }
      }
      return `M ${p1.x} ${p1.y} C ${c1.x} ${c1.y}, ${c2.x} ${c2.y}, ${p2.x} ${p2.y}`
    },
    // 边标签位置（路径端点中点，稍微上移一点）
    edgeLabelPos(edge) {
      const from = this.findNode(edge.from)
      const to = this.findNode(edge.to)
      if (!from || !to) return { x: 0, y: 0 }
      const { p1, p2 } = this.edgePorts(from, to)
      return { x: (p1.x + p2.x) / 2, y: (p1.y + p2.y) / 2 - 6 }
    },
    onNodeEnter(node, evt) {
      this.hoverNode = node
      this.updateTooltipPosition(evt)
    },
    onNodeMove(evt) {
      if (this.hoverNode) this.updateTooltipPosition(evt)
    },
    onNodeLeave() {
      this.hoverNode = null
    },
    updateTooltipPosition(evt) {
      const container = this.$refs.dagContainer
      if (!container) return
      const rect = container.getBoundingClientRect()
      const padding = 12
      // 将鼠标位置转换为容器内的相对坐标，并稍作偏移
      this.tooltipX = evt.clientX - rect.left + 12
      this.tooltipY = evt.clientY - rect.top + 12
      // 简单的边界处理，避免超出容器可视范围
      const maxX = rect.width - 220 - padding
      const maxY = rect.height - 100 - padding
      if (this.tooltipX > maxX) this.tooltipX = maxX
      if (this.tooltipY > maxY) this.tooltipY = maxY
    },
    selectNode(node) {
      this.selectedNode = node;
    },
    // 控制条：重置、步进、随机
    resetDAG() {
      const defaults = {
        A: 'success', B: 'running', C: 'success', D: 'pending', E: 'pending', F: 'failed', G: 'pending', H: 'success', I: 'pending', J: 'pending', K: 'pending', L: 'pending', M: 'pending', N: 'pending'
      }
      this.nodes.forEach(n => {
        if (defaults[n.id]) n.status = defaults[n.id]; else n.status = 'pending'
      })
    },
    stepDAG() {
      // 简单的“下一步”策略：找到第一个 running -> success；然后解锁其直接子节点（pending 变 running）
      const running = this.nodes.find(n => n.status === 'running')
      if (running) {
        running.status = 'success'
        // 解锁子节点：若其所有前置均非 pending/failed，则置为 running
        const children = this.edges.filter(e => e.from === running.id).map(e => this.findNode(e.to)).filter(Boolean)
        children.forEach(ch => {
          const parents = this.edges.filter(e => e.to === ch.id).map(e => this.findNode(e.from)).filter(Boolean)
          const allReady = parents.every(p => p.status === 'success')
          if (ch.status === 'pending' && allReady) ch.status = 'running'
        })
        return
      }
      // 若没有 running，则尝试将第一个 pending 且所有前置成功的节点置为 running
      const candidate = this.nodes.find(n => {
        if (n.status !== 'pending') return false
        const parents = this.edges.filter(e => e.to === n.id).map(e => this.findNode(e.from)).filter(Boolean)
        return parents.every(p => p.status === 'success')
      })
      if (candidate) candidate.status = 'running'
    },
    randomizeStatuses() {
      const pool = ['pending', 'running', 'success', 'failed']
      this.nodes.forEach(n => {
        // 源头节点 A 保持成功，提高演示稳定性
        if (n.id === 'A') { n.status = 'success'; return }
        n.status = pool[Math.floor(Math.random() * pool.length)]
      })
    },
    // 进入/退出全屏
    async toggleFullscreen() {
      if (this.isFullscreen) {
        try {
          if (document.fullscreenElement) await document.exitFullscreen()
        } catch (e) {
          console.warn('退出全屏失败', e)
        }
      } else {
        const el = this.$refs.svgEl
        if (!el) return
        try {
          await el.requestFullscreen()
        } catch (e) {
          console.warn('进入全屏失败', e)
        }
      }
    },
    onFsChange() {
      const el = this.$refs.svgEl
      this.isFullscreen = !!document.fullscreenElement && (document.fullscreenElement === el)
    }
  },
  computed: {
    // 根据缩放后节点坐标自动计算画布尺寸
    canvasWidth() {
      const maxX = Math.max(...this.nodes.map(n => this.posX(n)), 0)
      return Math.ceil(maxX + this.nodeWidth / 2 + this.canvasPadding)
    },
    canvasHeight() {
      const maxY = Math.max(...this.nodes.map(n => this.posY(n)), 0)
      return Math.ceil(maxY + this.nodeHeight / 2 + this.canvasPadding)
    },
    // 实际 SVG 采用的宽高（可切换自动或自定义）
    svgViewWidth() {
      return this.useAutoSize ? this.canvasWidth : this.userWidth
    },
    svgViewHeight() {
      return this.useAutoSize ? this.canvasHeight : this.userHeight
    },
    // 在全屏状态下，SVG 使用百分比占满屏幕
    svgAttrWidth() {
      return this.isFullscreen ? '100%' : this.svgViewWidth
    },
    svgAttrHeight() {
      return this.isFullscreen ? '100%' : this.svgViewHeight
    }
  },
  mounted() {
    this.fetchWorkflow()
    document.addEventListener('fullscreenchange', this.onFsChange)
  },
  beforeUnmount() {
    document.removeEventListener('fullscreenchange', this.onFsChange)
  }
};
</script>

<style scoped>
.workflow-dag {
  padding: 24px;
  position: relative; /* 作为 tooltip 的定位上下文 */
  overflow: auto; /* 画布更大时允许滚动查看 */
}
.toolbar {
  margin: 8px 0 12px;
}
.toolbar button + button { margin-left: 8px; }
.toolbar .sep { display:inline-block; width: 16px; }
.toolbar .ctrl { margin-left: 8px; font-size: 12px; color: #444; }
svg {
  background: #fafafa;
  border: 1px solid #eee;
}
/* 全屏时，让 svg 占满可用空间（不改变背景色） */
svg:fullscreen {
  width: 100% !important;
  height: 100% !important;
  border: none;
}
/* SVG 内全屏按钮：低存在感，悬停更明显 */
.svg-fs-btn rect {
  stroke: #666;
  stroke-width: 1;
  fill: rgba(255, 255, 255, 0.4);
  opacity: 0.35;
  transition: opacity 0.2s ease, fill 0.2s ease, stroke 0.2s ease;
}
.svg-fs-btn text {
  fill: #333;
  opacity: 0.7;
  pointer-events: none;
  transition: opacity 0.2s ease, fill 0.2s ease;
}
.svg-fs-btn:hover rect {
  opacity: 0.95;
  fill: rgba(255, 255, 255, 0.9);
}
.svg-fs-btn:hover text {
  opacity: 1;
  fill: #111;
}
.legend {
  display: flex;
  gap: 12px;
  margin-top: 10px;
  align-items: center;
  flex-wrap: wrap;
}
.legend-item { font-size: 12px; color: #555; display: flex; align-items: center; }
.legend .swatch {
  width: 14px; height: 14px; border-radius: 3px; display: inline-block; margin-right: 6px; background: #bdbdbd;
}
.legend .swatch.dashed {
  background: linear-gradient(90deg, #bdbdbd 0 40%, transparent 40% 60%, #bdbdbd 60% 100%);
}
.node-tooltip {
  position: absolute;
  width: 220px;
  max-width: calc(100% - 24px);
  background: #ffffff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  padding: 10px 12px;
  z-index: 10;
  pointer-events: none; /* 避免遮挡鼠标事件而导致闪烁 */
}
.node-tooltip-title {
  font-weight: 600;
  margin-bottom: 6px;
}
.node-tooltip-line {
  font-size: 12px;
  line-height: 18px;
  color: #333;
}
.node-tooltip-line .label {
  color: #888;
}
.node-detail {
  margin-top: 16px;
  padding: 12px;
  background: #fffbe6;
  border: 1px solid #ffe58f;
  border-radius: 6px;
}
/* DAG 节点卡片样式 */
.dag-node {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border: 2px solid #bdbdbd;
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.10);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 13px;
  user-select: none;
  position: relative;
  transition: box-shadow 0.2s, border-color 0.2s;
  box-sizing: border-box;
  padding: 6px 8px;
  text-align: center;
}
.dag-node:hover {
  box-shadow: 0 8px 24px rgba(33,150,243,0.18);
  border-color: #2196f3;
}
.dag-node-dot {
  position: absolute;
  top: 8px;
  left: 8px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  box-shadow: 0 1px 4px rgba(0,0,0,0.10);
}
.dag-node.is-running {
  /* 运行中：脉冲效果 */
  animation: dag-node-pulse 1.6s ease-in-out infinite;
}
.dag-node.is-running .dag-node-dot {
  /* 点也闪烁，增强视觉提示 */
  animation: dag-dot-blink 1s ease-in-out infinite;
}
@keyframes dag-node-pulse {
  0% { box-shadow: 0 4px 16px rgba(33,150,243,0.10); }
  50% { box-shadow: 0 0 0 8px rgba(33,150,243,0.28); }
  100% { box-shadow: 0 4px 16px rgba(33,150,243,0.10); }
}
@keyframes dag-dot-blink {
  0% { opacity: 1; }
  50% { opacity: 0.25; }
  100% { opacity: 1; }
}
.dag-node-title {
  font-weight: bold;
  margin-bottom: 2px;
  white-space: normal; /* 允许换行 */
  word-break: break-word; /* 中文/英文混排都能换行 */
}
.dag-node-status {
  font-size: 11px;
}
.edge-label {
  font-size: 11px;
  fill: #444;
  user-select: none;
  paint-order: stroke;
  stroke: #ffffff;
  stroke-width: 3px; /* 白色描边，增强可读性 */
  text-shadow: 0 1px 1px rgba(0,0,0,0.1);
}
</style>
