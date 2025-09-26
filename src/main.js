import { createApp } from 'vue'
import './style.css'
import App from './App.vue'

import { createRouter, createWebHistory } from 'vue-router'
import Home from './components/Home.vue'
import DeviceList from './components/DeviceList.vue'
import About from './components/About.vue'
import WorkflowDAG from './components/WorkflowDAG.vue'
import Auth from './components/Auth.vue'
import WorkflowList from './components/WorkflowList.vue'

const routes = [
	{ path: '/', component: Home },
	{ path: '/devices', component: DeviceList },
	{ path: '/about', component: About },
	{ path: '/workflows', component: WorkflowList },
	{ path: '/workflow/:id', component: WorkflowDAG },
	// 兼容旧地址，直接进入一个默认 DAG
	{ path: '/workflow', component: WorkflowDAG },
	{ path: '/auth', component: Auth }
]

const router = createRouter({
	history: createWebHistory(),
	routes
})

const app = createApp(App)
app.use(router)
app.mount('#app')
