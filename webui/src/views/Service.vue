<template>
  <b-container fluid class="">
    <b-row class="">
      <b-col class="sidebar p-0 pt-5">
        <div>
          <img src="@/assets/endpoint.png" width="36" height="36" title="Endpoints" />
        </div>
        <div>
          <img src="@/assets/models.png"  width="36" height="36" title="Models" />
        </div>
      </b-col>
      <b-col>
      <b-container fluid class="mt-5">
        <b-row class="mb-4 ml-4 mr-4">
          <b-col>
            <service-info :service="service"></service-info>
          </b-col>
        </b-row>
        <router-view :service="service"></router-view>
      </b-container>
      </b-col>
    </b-row>
  </b-container>
</template>

<script>
import ServiceInfo from '@/components/ServiceInfo'
import Api from '@/mixins/Api'

export default {
    name: 'service',
    components: {   
      'service-info': ServiceInfo,
    },
    mixins: [Api],
    data(){
      return {
        service: null,
        timer: null,
        loaded: false
      }
    },
    created() {
      this.getData();
      this.timer = setInterval(this.getData, 20000)
    },
    methods: {
      async getData() {
        let serviceName = this.$route.params.name
        this.service = await this.getService(serviceName)
        this.loaded = true
      }
    },
    beforeDestroy () {
      clearInterval(this.timer)
    }
}
</script>

<style scoped>
.container-fluid{
  height: 100vh;
}
.sidebar{
  width: 54px;
  background-color: white;
  min-height: 700px;
  position: sticky;
  flex: 0 0 54px;
}
.sidebar div{
  margin-top: 1rem;
  padding-left: 6px;
  padding-right: 6px;
}
.sidebar div:hover{
  background-color: #f5f6fa;
  cursor: pointer;
}
.content{
  min-width: 800px
}
</style>