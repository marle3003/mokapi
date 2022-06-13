<template>
  <b-container fluid class="" v-if="service !== null">
    <b-row class="">
      <b-col>
      <b-container fluid class="mt-5">
        <b-row class="mb-4">
          <b-col>
            <service-info :service="service"></service-info>
          </b-col>
        </b-row>
        <b-row class="mb-4">
          <b-col>
            <b-card no-body class="w-100">
              <b-tabs card>
                <b-tab title="Topics" v-on:click="selectedTab='topics'" :class="[selectedTab === 'topics' ? 'active' : '']">
                  <b-table :items="service.topics" :fields="topicFields" table-class="dataTable">
                    <template v-slot:cell(key)="data">
                      <schema v-bind:schema="data.item.key" />
                    </template>
                    <template v-slot:cell(value)="data">
                      <schema v-bind:schema="data.item.value" />
                    </template>
                  </b-table>
                </b-tab>
                <b-tab title="Brokers" v-on:click="selectedTab='servers'" :class="[selectedTab === 'servers' ? 'active' : '']">
                  <b-table :items="service.servers" :fields="serverFields" table-class="dataTable">
                    <template v-slot:cell(configs)="data">
                      <div v-for="c in data.item.configs" :key="c.key" class="config">
                        <span>{{ c.key }}:&nbsp;</span><span>{{ c.value }}</span>
                      </div>
                    </template>
                  </b-table>
                </b-tab>
              </b-tabs>
            </b-card>
          </b-col>
        </b-row>
        <router-view :service="service"></router-view>
      </b-container>
      </b-col>
    </b-row>
  </b-container>
</template>

<script>
import ServiceInfo from '@/components/kafka/ServiceInfo'
import Api from '@/mixins/Api'
import Schema from '@/components/Schema'

export default {
  name: 'service',
  components: {
    'service-info': ServiceInfo,
    'schema': Schema
  },
  mixins: [Api],
  data () {
    return {
      service: null,
      timer: null,
      loaded: false,
      topicFields: ['name', 'description', 'key', 'value'],
      serverFields: ['url', 'configs'],
      selectedTab: 'topics'
    }
  },
  created () {
    this.getData()
    this.timer = setInterval(this.getData, 20000)
  },
  methods: {
    async getData () {
      let response = await this.$http.get(this.baseUrl + '/api/services/kafka/' + this.$route.params.name)
      let service = response.data
      service.servers.sort(this.compare)
      this.service = service
      this.loaded = true
    },
    routerLinkToEndpoints (item, index, event) {
      this.$router.push({ name: 'endpoints' })
    },
    routerLinkToModels (item, index, event) {
      this.$router.push({ name: 'models' })
    },
    compare (s1, s2) {
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
.config{
  padding-bottom: 3px;
}
</style>
