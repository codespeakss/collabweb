<template>
  <div class="device-list">
    <h2>设备列表</h2>
    <div class="toolbar">
      <label>
        排序字段：
        <select v-model="sortBy" @change="onSortChange">
          <option value="lastonline">最近在线</option>
          <option value="createdat">创建时间</option>
          <option value="updatedat">更新时间</option>
          <option value="name">名称</option>
          <option value="id">ID</option>
          <option value="type">类型</option>
        </select>
      </label>
      <label>
        方向：
        <select v-model="order" @change="onSortChange">
          <option value="desc">降序</option>
          <option value="asc">升序</option>
        </select>
      </label>
      <button @click="fetchDevices()">刷新</button>
    </div>
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
    <div v-if="devices.length" class="pagination">
      <button @click="changePage(currentPage - 1)" :disabled="currentPage === 1">上一页</button>
      <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
      <button @click="changePage(currentPage + 1)" :disabled="currentPage === totalPages">下一页</button>
    </div>
    <div v-else>暂无设备数据</div>
    <div v-if="error" class="error">{{ error }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

function formatUTC(ts) {
  if (!ts) return '未知';
  const date = new Date(ts * 1000);
  return date.toISOString().replace('T', ' ').replace(/\..+/, '');
}

const devices = ref([])
const error = ref('')
const currentPage = ref(1)
const pageSize = 20
const totalPages = ref(1)
const total = ref(0)
const sortBy = ref('lastonline')
const order = ref('desc')

async function fetchDevices(page = 1) {
  error.value = ''
  try {
    const params = new URLSearchParams({
      page: String(page),
      page_size: String(pageSize),
      sort_by: sortBy.value,
      order: order.value,
    })
    const res = await fetch(`/api/v1/devices?${params.toString()}`)
    if (!res.ok) throw new Error('服务端错误')
    const data = await res.json()
    devices.value = data.devices || []
    total.value = data.total || 0
    totalPages.value = Math.ceil(total.value / pageSize)
    currentPage.value = data.page || page
  } catch (e) {
    error.value = e.message
  }
}

function changePage(page) {
  if (page < 1 || page > totalPages.value) return
  fetchDevices(page)
}

function onSortChange() {
  // 变更排序后回到第 1 页
  fetchDevices(1)
}

onMounted(() => {
  fetchDevices(1)
})

</script>

<style scoped>
.device-list { padding: 2em; }
.error { color: red; margin-top: 1em; }
.toolbar { display: flex; align-items: center; gap: 1em; }
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
.pagination {
  margin: 1em 0;
  display: flex;
  align-items: center;
  gap: 1em;
}
.pagination button {
  padding: 0.4em 1em;
  border: 1px solid #ccc;
  background: #f5f5f5;
  cursor: pointer;
  border-radius: 4px;
}
.pagination button:disabled {
  background: #eee;
  color: #aaa;
  cursor: not-allowed;
}
</style>
