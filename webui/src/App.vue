<script setup lang="ts">
import { RouterView, type RouteLocationNormalizedLoaded } from 'vue-router'
import Header from './components/Header.vue'
import Shortcuts from '@/components/Shortcuts.vue'

const isDashboard = import.meta.env.VITE_DASHBOARD === 'true'

// A website page is built again on every URL. The dashboard keeps its shell (sidebar, tab bar) and
// only mounts the view again, see DashboardViews.
const viewKey = (route: RouteLocationNormalizedLoaded) => (route.meta.mode as string | undefined) ?? route.fullPath
</script>

<template>
  <Header></Header>
  <RouterView :key="viewKey($route)" />
  <shortcuts v-if="isDashboard"  />
</template>

<style scoped>
header {
  line-height: 1.5;
}
</style>
