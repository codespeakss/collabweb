<template>
  <div class="about-page">
    <h2>关于本项目</h2>
    <div class="about-card">
      <p class="about-desc">本项目基于 <b>Vue 3</b> + <b>Vite</b> 构建，采用组件化开发，支持现代前端开发流程。主要功能包括设备列表展示（通过 RESTful API 获取）和项目简介页面，适合快速原型和中小型项目开发。</p>
      <ul class="about-list">
        <li><span>前端框架：</span>Vue 3</li>
        <li><span>构建工具：</span>Vite</li>
        <li><span>组件化开发：</span>易于扩展</li>
        <li><span>开发体验：</span>支持热重载和高效开发</li>
      </ul>
    </div>
  </div>
  
</template>

<script setup>
// 无需脚本逻辑
</script>

<style scoped>
.about-page {
  padding: 24px; /* 与 WorkflowDAG、DeviceList 对齐 */
}

.about-card {
  max-width: 680px;
  width: 100%;
  margin: 12px 0 0 0; /* 不再水平居中，和其它页面左对齐 */
  padding: 2em 2.5em;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
  display: flex;
  flex-direction: column;
  gap: 1.5em;
}

.about-desc {
  font-size: 1.1em;
  color: #555;
  line-height: 1.7;
}

.about-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.8em;
}
.about-list li {
  background: #f7f7fa;
  border-radius: 8px;
  padding: 0.7em 1em;
  font-size: 1em;
  color: #333;
  display: flex;
  align-items: center;
}
.about-list li span {
  font-weight: 600;
  color: #1976d2;
  margin-right: 0.5em;
}

@media (max-width: 600px) {
  .about-card {
    padding: 1em;
    max-width: 98vw;
  }
}
</style>
