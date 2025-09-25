<script setup>
import HelloWorld from './components/HelloWorld.vue'
</script>


<template>
  <nav style="margin-bottom:2em">
    <router-link to="/">首页</router-link> |
    <router-link to="/devices">设备列表</router-link> |
    <router-link to="/about">关于</router-link>
  </nav>
  <router-view />
</template>

<style scoped>
.logo {
  height: 6em;
  padding: 1.5em;
  will-change: filter;
  transition: filter 300ms;
}
.logo:hover {
  filter: drop-shadow(0 0 2em #646cffaa);
}
.logo.vue:hover {
  filter: drop-shadow(0 0 2em #42b883aa);
}
</style>
