<template>
  <div class="dag-renderer" ref="dagContainer">
    <svg ref="svgEl" :width="svgAttrWidth" :height="svgAttrHeight" :viewBox="`0 0 ${svgViewWidth} ${svgViewHeight}`">
      <g v-for="(node, idx) in computedNodes" :key="node.id">
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
      <g v-for="edge in computedEdges" :key="edge.from + '-' + edge.to">
        <path
          :d="edgePath(edge)"
          fill="none"
          :stroke="edgeColor(edge)"
          stroke-width="2"
          :marker-end="edgeMarker(edge)"
          :stroke-dasharray="edgeDasharray(edge)"
          :class="edgeClasses(edge)"
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
    </svg>
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
  </div>
</template>

<script>
export default {
  name: 'DAGRenderer',
  props: {
    nodes: { type: Array, default: () => [] },
    edges: { type: Array, default: () => [] },
    width: { type: Number, default: 600 },
    height: { type: Number, default: 400 }
  },
  data() {
    return {
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
      // 计算后的节点位置
      computedNodes: [],
      computedEdges: []
    };
  },
  watch: {
    nodes: {
      handler() { this.updateLayout() },
      deep: true,
      immediate: true
    },
    edges: {
      handler() { this.updateLayout() },
      deep: true,
      immediate: true
    }
  },
  methods: {
    updateLayout() {
      if (!this.nodes.length) {
        this.computedNodes = []
        this.computedEdges = []
        return
      }
      const { nodes: laidOutNodes, edges: laidOutEdges } = this.computeLayout([...this.nodes], [...this.edges])
      this.computedNodes = laidOutNodes
      this.computedEdges = laidOutEdges
    },
    // 基于有向无环图的简单前端布局（从 WorkflowDAG.vue 复制）
    computeLayout(nodes, edges) {
      // 构建索引
      const idToNode = new Map(nodes.map(n => [n.id, n]))
      // 过滤掉在 nodes 里不存在的边
      const safeEdges = edges.filter(e => idToNode.has(e.from) && idToNode.has(e.to))
      // 统计入度与父子关系
      const indeg = new Map()
      const parents = new Map()
      const children = new Map()
      nodes.forEach(n => {
        indeg.set(n.id, 0)
        parents.set(n.id, [])
        children.set(n.id, [])
      })
      safeEdges.forEach(e => {
        indeg.set(e.to, (indeg.get(e.to) || 0) + 1)
        parents.get(e.to).push(e.from)
        children.get(e.from).push(e.to)
      })
      // Kahn 拓扑 + 最长路径分层
      const queue = []
      indeg.forEach((v, k) => { if (v === 0) queue.push(k) })
      const order = []
      const level = new Map()
      queue.forEach(id => level.set(id, 0))
      while (queue.length) {
        const u = queue.shift()
        order.push(u)
        const lv = level.get(u) || 0
        for (const v of (children.get(u) || [])) {
          // 以最长父路径为层级
          level.set(v, Math.max((level.get(v) || 0), lv + 1))
          indeg.set(v, (indeg.get(v) || 0) - 1)
          if (indeg.get(v) === 0) queue.push(v)
        }
      }
      // 若存在环，fallback：将未出现在 order 的节点追加并给默认层
      if (order.length < nodes.length) {
        nodes.forEach(n => {
          if (!order.includes(n.id)) {
            order.push(n.id)
            if (!level.has(n.id)) level.set(n.id, 0)
          }
        })
      }
      // 分层分组
      const layers = []
      nodes.forEach(n => {
        const lv = level.get(n.id) || 0
        if (!layers[lv]) layers[lv] = []
        layers[lv].push(n)
      })
      // 初始层内顺序：按原有 x 提示或 id，提供稳定基线
      layers.forEach(arr => arr.sort((a, b) => ((a.x ?? 0) - (b.x ?? 0)) || a.id.localeCompare(b.id)))
      // 布局参数
      const layerGap = Math.max(this.nodeHeight * 1.8, 120)
      const nodeGap = Math.max(this.nodeWidth * 1.8, 140)
      const paddingX = this.canvasPadding
      const paddingY = this.canvasPadding
      const maxLayerSize = Math.max(1, ...layers.map(arr => arr.length))
      // 逐层定位
      layers.forEach((arr, li) => {
        // 水平居中：以最大层宽为参考，当前层在其中居中摆放
        const baseX = paddingX + ((maxLayerSize - arr.length) * nodeGap) / 2
        const y = paddingY + li * layerGap
        arr.forEach((n, idx) => {
          // 将布局写入 n.x/n.y（前端专用）
          n.x = baseX + (idx * nodeGap)
          n.y = y
        })
      })
      return { nodes, edges: safeEdges }
    },
    // 布局缩放后的坐标
    posX(node) { return node.x * this.layoutScaleX + this.canvasPadding },
    posY(node) { return node.y * this.layoutScaleY + this.canvasPadding },
    findNode(id) {
      return this.computedNodes.find(n => n.id === id);
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
    // 与被选中节点相连的边：用于判断是否需要"流水效果"
    isEdgeConnectedToSelected(edge) {
      if (!this.selectedNode) return false
      return edge.from === this.selectedNode.id || edge.to === this.selectedNode.id
    },
    // 边的 class 绑定：当与选中节点相连时添加动画类
    edgeClasses(edge) {
      return { 'edge-flow': this.isEdgeConnectedToSelected(edge) }
    },
    // 决定边的虚线样式
    edgeDasharray(edge) {
      if (edge.type === 'conditional') return '6,6'
      if (this.isEdgeConnectedToSelected(edge)) return '10,10'
      return '0'
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
    }
  },
  computed: {
    // 根据缩放后节点坐标自动计算画布尺寸
    canvasWidth() {
      if (!this.computedNodes.length) return this.width
      const maxX = Math.max(...this.computedNodes.map(n => this.posX(n)), 0)
      return Math.ceil(maxX + this.nodeWidth / 2 + this.canvasPadding)
    },
    canvasHeight() {
      if (!this.computedNodes.length) return this.height
      const maxY = Math.max(...this.computedNodes.map(n => this.posY(n)), 0)
      return Math.ceil(maxY + this.nodeHeight / 2 + this.canvasPadding)
    },
    // 实际 SVG 采用的宽高
    svgViewWidth() {
      return Math.max(this.canvasWidth, this.width)
    },
    svgViewHeight() {
      return Math.max(this.canvasHeight, this.height)
    },
    svgAttrWidth() {
      return this.svgViewWidth
    },
    svgAttrHeight() {
      return this.svgViewHeight
    }
  }
}
</script>

<style scoped>
.dag-renderer {
  position: relative;
  overflow: auto;
  border: 1px solid #eee;
  border-radius: 8px;
  background: #fafafa;
}
svg {
  background: #fafafa;
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
  pointer-events: none;
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
  animation: dag-node-pulse 1.6s ease-in-out infinite;
}
.dag-node.is-running .dag-node-dot {
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
  white-space: normal;
  word-break: break-word;
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
  stroke-width: 3px;
  text-shadow: 0 1px 1px rgba(0,0,0,0.1);
}
/* 边的"流水"动态效果 */
.edge-flow {
  stroke-linecap: round;
  animation: edge-flow 1.2s linear infinite;
}
@keyframes edge-flow {
  from { stroke-dashoffset: 0; }
  to { stroke-dashoffset: -28; }
}
</style>
