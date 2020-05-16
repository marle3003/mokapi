<template>
  <div id="app">
    <div id="nav">
      <top-header></top-header>
    </div>
    <router-view/>
  </div>
</template>

<script>
import TopHeader from "@/components/TopHeader.vue"

export default {
  name: 'App',
  components: {
    "top-header": TopHeader,

  }
}
</script>

<style>
body{
  background-color: #f5f6fa;
  height: 100%
}
#app {

  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
