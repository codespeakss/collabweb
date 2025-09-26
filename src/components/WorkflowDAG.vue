<template>
  <div class="workflow-dag" ref="dagContainer">
    <h2>Workflow DAG 状态展示</h2>
    <svg :width="svgWidth" :height="svgHeight">
      <g v-for="(node, idx) in nodes" :key="node.id">
        <!-- 节点圆形 -->
          <foreignObject :x="node.x - nodeWidth / 2" :y="node.y - nodeHeight / 2" :width="nodeWidth" :height="nodeHeight">
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
        <line
          :x1="findNode(edge.from).x"
          :y1="findNode(edge.from).y + nodeHeight / 2"
          :x2="findNode(edge.to).x"
          :y2="findNode(edge.to).y - nodeHeight / 2"
          :stroke="edgeColor(edge)"
          stroke-width="2"
          :marker-end="`url(#${isEdgeTraversed(edge) ? 'arrow-green' : 'arrow-gray'})`"
        />
      </g>
      <defs>
        <marker id="arrow-green" markerWidth="6" markerHeight="6" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
          <path d="M0,0 L6,3 L0,6 Z" fill="#4caf50" />
        </marker>
        <marker id="arrow-gray" markerWidth="6" markerHeight="6" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
          <path d="M0,0 L6,3 L0,6 Z" fill="#bdbdbd" />
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
      svgWidth: 600,
      svgHeight: 460,
      nodeRadius: 28,
      nodeWidth: 80,
      nodeHeight: 72,
      selectedNode: null,
      hoverNode: null,
      tooltipX: 0,
      tooltipY: 0,
      // 示例 DAG 数据
        nodes: [
          { id: 'A', name: '任务A', x: 300, y: 40, status: 'success', desc: '数据准备' },
          { id: 'B', name: '任务B', x: 180, y: 120, status: 'running', desc: '数据清洗' },
          { id: 'C', name: '任务C', x: 420, y: 120, status: 'success', desc: '特征工程' },
          { id: 'D', name: '任务D', x: 120, y: 220, status: 'pending', desc: '模型训练1' },
          { id: 'E', name: '任务E', x: 240, y: 220, status: 'pending', desc: '模型训练2' },
          { id: 'F', name: '任务F', x: 360, y: 220, status: 'failed', desc: '模型训练3' },
          { id: 'G', name: '任务G', x: 480, y: 220, status: 'pending', desc: '模型训练4' },
          { id: 'H', name: '任务H', x: 180, y: 320, status: 'success', desc: '评估1' },
          { id: 'I', name: '任务I', x: 420, y: 320, status: 'pending', desc: '评估2' },
          { id: 'J', name: '任务J', x: 300, y: 400, status: 'pending', desc: '结果汇总' },
        ],
      edges: [
        { from: 'A', to: 'B' },
        { from: 'A', to: 'C' },
        { from: 'B', to: 'D' },
        { from: 'B', to: 'E' },
        { from: 'C', to: 'F' },
        { from: 'C', to: 'G' },
        { from: 'D', to: 'H' },
        { from: 'E', to: 'H' },
        { from: 'F', to: 'I' },
        { from: 'G', to: 'I' },
        { from: 'H', to: 'J' },
        { from: 'I', to: 'J' },
      ],
    };
  },
  methods: {
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
    // 根据是否遍历返回边颜色
    edgeColor(edge) {
      return this.isEdgeTraversed(edge) ? '#4caf50' : '#bdbdbd'
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
  },
};
</script>

<style scoped>
.workflow-dag {
  padding: 24px;
  position: relative; /* 作为 tooltip 的定位上下文 */
}
svg {
  background: #fafafa;
  border: 1px solid #eee;
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
</style>
