<template>
  <div class="device-list">
    <h2>设备列表</h2>
    <button @click="fetchDevices">刷新</button>
      <table v-if="devices.length" class="device-table">
        <thead>
          <tr>
            <th>设备名称</th>
            <th>ID</th>
            <th>类型</th>
            <th>最近在线</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="device in devices" :key="device.id">
            <td>{{ device.name }}</td>
            <td>{{ device.id }}</td>
            <td>{{ device.type }}</td>
            <td>{{ formatUTC(device.lastOnline) }}</td>
          </tr>
        </tbody>
      </table>
    <div v-else>暂无设备数据</div>
    <div v-if="error" class="error">{{ error }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

function formatUTC(ts) {
  if (!ts) return '未知';
  const date = new Date(ts * 1000);
  return date.toISOString().replace('T', ' ').replace(/\..+/, '');
}

const devices = ref([])
const error = ref('')

async function fetchDevices() {
  error.value = ''
  try {
    // 请将 URL 替换为实际服务端接口
    const res = await fetch('/api/devices')
    if (!res.ok) throw new Error('服务端错误')
    devices.value = await res.json()
  } catch (e) {
    error.value = e.message
  }
}

fetchDevices()
</script>

<style scoped>
.device-list { padding: 2em; }
.error { color: red; margin-top: 1em; }

.device-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 1em;
}
.device-table th, .device-table td {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}
.device-table th {
  background-color: #f5f5f5;
}
.device-table tr:nth-child(even) {
  background-color: #fafafa;
}
</style>
