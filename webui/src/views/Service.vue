<template>
  <b-container fluid="md" class="mt-5">
    <b-row class="mb-4 ml-4 mr-4">
      <b-col>
        <service-info :service="service"></service-info>
      </b-col>
    </b-row>
    <b-row class="ml-4 mr-4">
      <b-col>
      <router-view :service="service"></router-view>
      </b-col>
    </b-row>
  </b-container>
</template>

<script>
import serviceInfo from "@/components/serviceInfo.vue"

export default {
    name: "service",
    components: {    
      "service-info": serviceInfo,
    },
    data(){
      return {
        service: null
      }
    },
    mounted(){
      let serviceName = this.$route.params.name
      this.$http.get(
      'http://localhost:8081/api/services/'+serviceName)
        .then(response => (this.service = response.data))
    }
}
</script>

<style scoped>

</style>