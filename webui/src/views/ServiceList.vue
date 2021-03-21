<template>
  <div class="service-list">
      <h1>Services</h1>
      <div class="page-body">
        <b-link :to="{ name: 'service', params: {name: service.name} }" router-tag="div" v-for="service in services" :key="service.name">
          <b-card>
            <b-row>
              <b-col class="col-auto">
                <img src="@/assets/service.png" />
              </b-col>
              <b-col>
                <p class="name">{{ service.name }}</p>
                <vue-simple-markdown :source="service.description" />
              </b-col>
              <b-col>
              <p>{{ service.version }}</p>
              </b-col>
            </b-row>
          </b-card>
        </b-link>
      </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'

export default {
  components: {},
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

      function compare(a, b) {
        if (a.name < b.name) {
          return -1
        }
        if (a.name > b.name) {
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
      margin: auto;
      margin-top: 42px;
  }
  .card{
    margin: 15px;
    margin-left: 0;
    cursor: pointer;
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
