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
  .operation{
      text-transform: uppercase;
  }
  .put, .patch{
    background-color: var(--orange);
  }
  .get, .options{
    background-color: var(--blue);
  }
  .delete{
    background-color: var(--red);
  }
  .post{
    background-color: var(--green);
  }
  .head, .trace{
    background-color: var(--yellow);
  }
</style>
