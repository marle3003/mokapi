<script setup lang="ts">
import { RouterView } from 'vue-router'
import Header from './components/Header.vue'

var theme = 'light';
if (import.meta.env.VITE_DASHBOARD === true) {
  theme = 'dark'
}
if (localStorage.getItem('theme')) {
    if(localStorage.getItem('theme') == 'dark'){
        theme = 'dark';
    }
}

document.documentElement.setAttribute('data-theme', theme);
document.documentElement.setAttribute('style', 'color-scheme:'+theme)
document.documentElement.setAttribute('data-bs-theme', theme)
</script>

<template>
  <Header></Header>
  <RouterView :key="$route.fullPath" />
</template>

<style scoped>
header {
  line-height: 1.5;
}
</style>
