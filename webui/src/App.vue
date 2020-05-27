<template>
  <div id="app">
    <div id="nav">
      <top-header></top-header>
    </div>
    <router-view/>
  </div>
</template>

<script>
import TopHeader from '@/components/TopHeader.vue'

export default {
  name: 'App',
  components: {
    'top-header': TopHeader,

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
  .operation.put, .operation.patch{
    background-color: var(--orange);
  }
  .operation.get, .operation.options{
    background-color: var(--blue);
  }
  .operation.delete{
    background-color: var(--red);
  }
  .operation.post{
    background-color: var(--green);
  }
  .operation.head, .operation.trace{
    background-color: var(--yellow);
  }
</style>
