<script setup lang="ts">
import { RouterView } from 'vue-router'
import Header from './components/Header.vue'
import Shortcuts from '@/components/Shortcuts.vue'

const isDashboard = import.meta.env.VITE_DASHBOARD === 'true'
</script>

<template>
  <Header></Header>
  <RouterView :key="$route.fullPath" />
  <shortcuts v-if="isDashboard"  />
</template>

<style scoped>
header {
  line-height: 1.5;
}
</style>
