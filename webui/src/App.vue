<script setup lang="ts">
import { RouterView } from 'vue-router'
import Header from './components/Header.vue'
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
