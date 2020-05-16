<template>
  <div class="container mt-5">
  <div class="row mb-4">
    <service-info></service-info>
  </div>
  <div class="row">
    <router-view></router-view>
  </div>
  </div>
</template>

<script>
import serviceInfo from "@/components/serviceInfo.vue"

export default {
    name: "service",
    components: {    
      "service-info": serviceInfo,
    }
}
</script>

<style scoped>

</style>