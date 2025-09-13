<script setup lang="ts">
import { RouterView } from 'vue-router'
import Header from './components/Header.vue'
import { inject } from 'vue';


const canSwitchTheme = inject<boolean>('canSwitchTheme')
var theme = 'light';
if (!canSwitchTheme) {
  theme = 'dark'
} else {
  const saved = localStorage.getItem('theme')
  if (saved) {
      theme = saved;
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
