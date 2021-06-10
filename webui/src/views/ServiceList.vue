<template>
  <div class="service-list">
      <h1>Services</h1>
      <div class="page-body">
        <b-link :to="{ name: service.type.toLowerCase(), params: {name: service.name } }"
          router-tag="div" v-for="service in services" :key="service.name"
          :class="[service.name === undefined ? 'disabled' : '']" :event="[service.name === undefined ? '' : 'click']">
          <async-service-info :service="service" v-if="service.type === 'AsyncAPI'"></async-service-info>
          <service-info :service="service" v-else></service-info>
        </b-link>
      </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import ServiceInfo from '@/components/ServiceInfo'
import AsyncServiceInfo from '@/components/asyncapi/ServiceInfo'

export default {
  components: {
    'service-info': ServiceInfo,
    'async-service-info': AsyncServiceInfo
  },
  mixins: [Api],
  data () {
    return {
      services: [],
      timer: null,
      loaded: false
    }
  },
  created () {
    this.getData()
    this.timer = setInterval(this.getData, 20000)
  },
  methods: {
    async getData () {
      function compare (s1, s2) {
        if (s1.name === s2.name) {
          return 0
        }
        if (s1.name === undefined) {
          return 1
        } else if (s2.name === undefined) {
          return -1
        }
        const a = s1.name.toLowerCase()
        const b = s2.name.toLowerCase()
        if (a < b) {
          return -1
        }
        if (a > b) {
          return 1
        }
        return 0
      }

      let services = await this.getServices()
      this.services = services.sort(compare)
      this.loaded = true
    }
  },
  beforeDestroy () {
    clearInterval(this.timer)
  }
}
</script>

<style scoped>
  .service-list{
      width: 90%;
    margin: 42px auto auto;
  }
  .card{
    margin: 15px 15px 15px 0;
    cursor: pointer;
  }
  .disabled .card{
    cursor: not-allowed;
  }
  .card p{
      margin-bottom: 0;
  }
  .name{
    font-size: 1.25rem;
    font-weight: 500;
    padding-bottom: 0.5rem;
  }
</style>
