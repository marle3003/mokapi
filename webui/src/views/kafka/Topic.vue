<template>
  <div class="dashboard" v-if="topic !== null">
    <div class="page-header">
      <b-row class="mb-2 ml-0 mr-0">
        <b-col cols="auto" class="mr-auto pl-0">
          <h2>Kafka Topic</h2>
        </b-col>
        <b-col cols="auto" class="pr-0">
          <div class="close" @click="$router.go(-1)">
            <b-icon icon="x" class="border rounded p-1"></b-icon>
          </div>
        </b-col>
      </b-row>
    </div>
    <div class="page-body">
      <b-card-group deck>
      <b-card class="w-100">
        <b-row>
          <b-col>
            <p class="label">Name</p>
            <p>{{ topic.name }}</p>
          </b-col>
          <b-col>
            <p class="label">Partitions</p>
            <p>{{ topic.partitions }}</p>
          </b-col>
        </b-row>
        <b-row v-if="topic.description !== ''">
          <b-col>
            <p class="label">Description</p>
            <vue-simple-markdown :source="topic.description" />
          </b-col>
        </b-row>
      </b-card>
      </b-card-group>
      <b-card-group deck>
        <b-card body-class="info-body" class="text-center">
          <b-card-title class="info">Messages</b-card-title>
          <b-card-text class="text-center value">{{ topic.count }}</b-card-text>
        </b-card>
        <b-card body-class="info-body" class="text-center">
          <b-card-title class="info">Size</b-card-title>
          <b-card-text class="text-center value">{{ topic.size | prettyBytes }}</b-card-text>
        </b-card>
      </b-card-group>

      <b-card-group deck>
        <b-card class="w-100">
          <b-card-title class="info text-center">Messages</b-card-title>
          <b-table small hover class="dataTable" :items="topic.messages" :fields="messageFields" style="table-layout: fixed" @row-clicked="toggleDetails">
            <template v-slot:cell(show_details)="row">
              <div @click="toggleDetails(row)">
                <b-icon v-if="row.detailsShowing" icon="dash-square"></b-icon>
                <b-icon v-else icon="plus-square"></b-icon>
              </div>
            </template>
            <template v-slot:cell(time)="data">
              {{ data.item.time | moment }}
            </template>
            <template v-slot:row-details="row">
              <b-card class="w-100">
                <b-row class="mb-2">
                  {{ row.item.message }}
                </b-row>
              </b-card>
            </template>
          </b-table>
        </b-card>
      </b-card-group>
    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import moment from 'moment'

export default {
  name: 'Topic',
  mixins: [Api],
  data () {
    return {
      timer: null,
      topic: null,
      messageFields: [{key: 'show_details', label: '', thStyle: 'width: 3%'}, 'key', {key: 'message', class: 'kafka-message'}, 'partition', 'time'],
      detailsShown: []
    }
  },
  created () {
    this.getData()
    let refresh = this.$route.query.refresh
    if (refresh && refresh.length > 0) {
      let i = parseInt(refresh)
      if (!isNaN(i)) {
        this.timer = setInterval(this.getData, i * 1000)
      }
    }
    window.addEventListener('keyup', this.doCommand)
  },
  destroyed () {
    window.removeEventListener('keyup', this.doCommand)
  },
  methods: {
    async getData () {
      let kafka = this.$route.params.kafka
      let topic = this.$route.params.topic
      this.topic = await this.getKafkaTopic(kafka, topic)
    },
    toggleDetails (row) {
      row.toggleDetails()
      const index = this.detailsShown.indexOf(row.item.key)

      if (row.item._showDetails) {
        this.detailsShown.push(row.item.key)
      } else {
        this.detailsShown.splice(index, 1)
      }
    },
    doCommand (e) {
      let cmd = e.key.toLowerCase()
      if (cmd === 'escape') {
        this.$router.go(-1)
      }
    }
  },
  filters: {
    moment: function (date) {
      return moment(date).local().format('YYYY-MM-DD HH:mm:ss')
    },
    prettyBytes: function (num) {
      // jacked from: https://github.com/sindresorhus/pretty-bytes
      if (typeof num !== 'number' || isNaN(num)) {
        return 0
      }

      let exponent
      let unit
      let neg = num < 0
      let units = ['B', 'kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

      if (neg) {
        num = -num
      }

      if (num < 1) {
        return (neg ? '-' : '') + num + ' B'
      }

      exponent = Math.min(Math.floor(Math.log(num) / Math.log(1000)), units.length - 1)
      num = (num / Math.pow(1000, exponent)).toFixed(2) * 1
      unit = units[exponent]

      return (neg ? '-' : '') + num + ' ' + unit
    }
  },
  beforeDestroy () {
    clearInterval(this.timer)
  }
}
</script>

<style scoped>
.dashboard{
  width: 90%;
  margin: 42px auto auto;
}
.page-header h2{
  font-weight: 400;
}
.card{
  border-color: var(--var-border-color);
  margin: 7px;
}
.info{
  font-size: 0.7rem;
  font-weight: 300;
}
.info-body{
  padding: 0.8rem;
}
.value {
  font-size: 2.25rem;
  font-weight: 300;
}
</style>
<style>
.kafka-message{
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 50%;
  width: 50%;
}
</style>
