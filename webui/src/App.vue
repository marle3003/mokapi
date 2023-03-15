<script setup lang="ts">
import { RouterView } from 'vue-router'
import Header from './components/Header.vue'

var theme = 'dark';
if (localStorage.getItem('theme')) {
    if(localStorage.getItem('theme') == 'light'){
        theme = 'light';
    }
} else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
    theme = 'light';
}

document.documentElement.setAttribute('data-theme', theme);
if (theme == 'light' ) {
  import('@/assets/github.css')
} else{
  import('@/assets/github-dark.css')
}
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
