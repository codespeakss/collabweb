<template>
  <div class="workflow-dag">
    <h2>Workflow DAG 状态展示</h2>
    <svg :width="svgWidth" :height="svgHeight">
      <g v-for="(node, idx) in nodes" :key="node.id">
        <!-- 节点圆形 -->
          <foreignObject :x="node.x - nodeRadius" :y="node.y - nodeRadius" :width="nodeRadius * 2" :height="nodeRadius * 2">
            <div class="dag-node" :style="{ borderColor: statusColor(node.status) }" @click="selectNode(node)">
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
          :y1="findNode(edge.from).y + nodeRadius"
          :x2="findNode(edge.to).x"
          :y2="findNode(edge.to).y - nodeRadius"
          stroke="#888"
          stroke-width="2"
          marker-end="url(#arrow)"
        />
      </g>
      <defs>
        <marker id="arrow" markerWidth="6" markerHeight="6" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
          <path d="M0,0 L6,3 L0,6 Z" fill="#888" />
        </marker>
      </defs>
    </svg>
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
      svgHeight: 400,
      nodeRadius: 24,
      selectedNode: null,
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
    selectNode(node) {
      this.selectedNode = node;
    },
  },
};
</script>

<style scoped>
.workflow-dag {
  padding: 24px;
}
svg {
  background: #fafafa;
  border: 1px solid #eee;
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
  width: 56px;
  height: 56px;
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
.dag-node-title {
  font-weight: bold;
  margin-bottom: 2px;
}
.dag-node-status {
  font-size: 11px;
}
</style>
