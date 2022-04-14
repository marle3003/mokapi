<template>
  <div>
    <div v-if="error !== null">
      <b-alert show variant="danger">{{ error }}</b-alert>
    </div>
    <div class="dashboard" v-if="dashboard !== null">
        <div class="page-header">
            <b-navbar class="p-0">
              <b-navbar-nav>
                <b-nav-item :to="{ name: 'dashboard', query: {refresh: this.$route.query.refresh} }">Overview</b-nav-item>
                <b-nav-item :to="{ name: 'http', query: {refresh: this.$route.query.refresh} }" v-if="httpEnabled">HTTP</b-nav-item>
                <b-nav-item :to="{ name: 'kafka', query: {refresh: this.$route.query.refresh} }" v-if="dashboard.kafkaEnabled">Kafka</b-nav-item>
                <b-nav-item :to="{ name: 'smtp', query: {refresh: this.$route.query.refresh} }" v-if="dashboard.smtpEnabled">SMTP</b-nav-item>
              </b-navbar-nav>
            </b-navbar>
        </div>
        <div class="page-body">

          <b-card-group deck>
            <b-card body-class="info-body" class="text-center">
              <b-card-title class="info">Uptime Since</b-card-title>
              <b-card-text class="text-center value">{{ dashboard.startTime | fromNow }}</b-card-text>
              <b-card-text class="text-right additional">{{ dashboard.serverUptime | moment }}</b-card-text>
            </b-card>
            <b-card body-class="info-body" class="text-center">
              <b-card-title class="info">Memory Usage</b-card-title>
              <b-card-text class="text-center value">{{ dashboard.memoryUsage | prettyBytes }}</b-card-text>
              </b-card>
          </b-card-group>

          <b-card-group deck>
            <b-card body-class="info-body" class="text-center" v-if="httpEnabled">
              <b-card-title class="info">Total HTTP Requests</b-card-title>
              <b-card-text class="text-center value">{{ dashboard.httpRequests }}</b-card-text>
            </b-card>
            <b-card body-class="info-body" class="text-center" v-if="httpEnabled">
              <b-card-title class="info">HTTP Request Errors</b-card-title>
              <b-card-text class="text-center value" v-bind:class="{'text-danger': dashboard.httpErrorRequests > 0}">{{ dashboard.httpErrorRequests }}</b-card-text>
            </b-card>
            <b-card body-class="info-body" class="text-center" v-if="dashboard.kafkaEnabled">
              <b-card-title class="info">Received Kafka Messages</b-card-title>
              <b-card-text class="text-center value">{{ totalMessages }}</b-card-text>
            </b-card>
            <b-card body-class="info-body" class="text-center" v-if="dashboard.smtpEnabled">
              <b-card-title class="info">Total Mails</b-card-title>
              <b-card-text class="text-center value">{{ dashboard.totalMails }}</b-card-text>
            </b-card>
          </b-card-group>

          <b-card-group deck v-show="httpEnabled && $route.name === 'http' || $route.name === 'dashboard'">
            <b-card body-class="info-body" class="text-center">
               <b-card-title class="info">HTTP Services</b-card-title>
               <b-table :items="httpServices" :fields="httpFields" table-class="dataTable">
                <template v-slot:cell(method)="data">
                  <b-badge pill class="operation" :class="data.item.method.toLowerCase()" >{{ data.item.method }}</b-badge>
                </template>
                <template v-slot:cell(lastRequest)="data">
                  <span v-if="data.item.lastRequest === 0">-</span>
                  <span v-else>{{ data.item.lastRequest | moment}}</span>
                </template>
              </b-table>
            </b-card>
          </b-card-group>

          <div v-show="kafkaEnabled && $route.name === 'dashboard' || $route.name === 'kafka'">
            <b-card-group deck>
              <b-card body-class="info-body" class="text-center">
                 <b-card-title class="info">Kafka Clusters</b-card-title>
                 <b-table :items="kafkaServices" :fields="kafkaFields" table-class="dataTable selectable" @row-clicked="topicClickHandler">
                  <template v-slot:cell(lastMessage)="data">
                    <span v-if="data.item.lastMessage === 0">-</span>
                    <span v-else>{{ data.item.lastMessage | moment}}</span>
                  </template>
                </b-table>
              </b-card>
            </b-card-group>
          </div>

          <b-card-group deck v-show="$route.name === 'kafka'">
            <b-card body-class="info-body" class="text-center">
              <b-card-title class="info">Kafka Groups</b-card-title>
              <b-table :items="groups" :fields="groupFields" table-class="dataTable">
                <template v-slot:cell(members)="data">
                  {{ data.item.members.join(', ') }}
                </template>
              </b-table>
            </b-card>
          </b-card-group>

          <http-overview v-show="$route.name === 'http'" />

          <b-card-group deck v-show="$route.name === 'smtp'">
            <b-card class="w-100">
              <b-card-title class="info text-center">Recent Mails</b-card-title>
              <b-table hover :items="lastMails" :fields="lastMailField" class="dataTable selectable" @row-clicked="mailClickHandler">
                <template v-slot:cell(from)="data">
                  <div v-for="from in data.item.from" :key="from.Address">
                    <span v-if="from.Name !== ''">{{ from.Name }} &lt;</span><span>{{ from.Address }}</span><span v-if="from.Name !== ''">&gt;</span>
                  </div>
                </template>
                <template v-slot:cell(to)="data">
                  <div v-for="to in data.item.to" :key="to.Address">
                    <span v-if="to.Name !== ''">{{ to.Name }} &lt;</span><span>{{ to.Address }}</span><span v-if="to.Name !== ''">&gt;</span>
                  </div>
                </template>
                <template v-slot:cell(time)="data">
                  {{ data.item.time | moment}}
                </template>
              </b-table>
            </b-card>
          </b-card-group>
        </div>
    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import moment from 'moment'
import DoughnutChart from '@/components/DoughnutChart'
import TimeChart from '@/components/TimeChart'
import HttpOverview from '@/components/dashboard/HttpOverview'

export default {
  components: {
    'doughnut-chart': DoughnutChart,
    'time-chart': TimeChart,
    'http-overview': HttpOverview
  },
  mixins: [Api],
  data () {
    return {
      dashboard: null,
      httpServices: null,
      kafkaServices: null,
      groups: null,
      lastRequests: null,
      lastErrors: null,
      lastMails: null,
      loaded: false,
      timer: null,
      topicSizes: {},
      chartTopicSize: {},
      httpFields: [{key: 'name', class: 'text-left'}, {key: 'lastRequest', class: 'text-left'}, 'requests', 'errors'],
      kafkaFields: [{key: 'name', class: 'text-left'}, 'topics', 'lastMessage', 'messages', 'errors'],
      groupFields: [{key: 'name', class: 'text-left'}, {key: 'state', thStyle: 'width:10%'}, {key: 'assignmentStrategy', class: 'text-left'}, {key: 'coordinator', class: 'text-left'}, {key: 'leader', class: 'text-left'}, {key: 'members', class: 'text-left'}],
      lastMailField: ['from', 'to', {key: 'subject', class: 'subject'}, 'time'],
      error: null
    }
  },
  created () {
    this.init()
  },
  computed: {
    httpEnabled: function () {
      return this.httpServices !== null && this.httpServices.length > 0
    },
    kafkaEnabled: function () {
      return this.kafkaServices !== null && this.kafkaServices.length > 0
    },
    serviceStatus: function () {
      let serviceStatus = this.dashboard.serviceStatus
      let success = serviceStatus.total - serviceStatus.errors
      return {
        datasets: [{
          data: [success, serviceStatus.errors],
          backgroundColor: ['rgb(110, 181, 110)', 'rgb(186, 86, 86)']
        }],
        labels: ['Success', 'Errors']
      }
    },
    hasErrors: function () {
      return this.dashboard.lastErrors !== undefined && this.dashboard.lastErrors.length > 0
    }
  },
  filters: {
    moment: function (value) {
      return moment.unix(value).local().format('YYYY-MM-DD HH:mm:ss')
    },
    fromNow: function (value) {
      return moment.unix(value).fromNow(true)
    },
    duration: function (time) {
      let ms = Math.round(time / 1000000)
      let d = moment.duration(ms)
      if (d.seconds() < 1) {
        return d.milliseconds() + ' [ms]'
      } else if (d.minutes() < 1) {
        return d.seconds() + ' [sec]'
      }
      return moment.duration(d).minutes()
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
  methods: {
    async getData () {
      this.getDashboard().then(r => {
        this.dashboard = r
        this.error = null
      }, r => {
        this.dashboard = null
        this.error = r
      })
      this.getHttpServices().then(r => {
        this.httpServices = r
        this.error = null
      }, r => {
        this.httpServices = null
        this.error = r
      })
      this.getKafkaServices().then(r => {
        console.log(r)
        this.kafkaServices = r
        this.error = null
      }, r => {
        this.kafkaServices = null
        this.error = r
      })
      this.loaded = true
    },
    mailClickHandler (record) {
      this.$router.push({name: 'smtpMail', params: {id: record.id}})
    },
    topicClickHandler (record) {
      // this.$router.push({name: 'kafkaTopic', params: {kafka: record.service, topic: record.name}, query: {refresh: '5'}})
    },
    init () {
      this.getData()
      clearInterval(this.timer)
      let refresh = this.$route.query.refresh
      if (refresh && refresh.length > 0) {
        let i = parseInt(refresh)
        if (!isNaN(i)) {
          this.timer = setInterval(this.getData, i * 1000)
        }
      }
    }
  },
  beforeDestroy () {
    clearInterval(this.timer)
  },
  watch: {
    $route () {
      this.init()
    }
  }
}
</script>

<style scoped>
  .dashboard{
    width: 90%;
    margin: 12px auto auto;
  }
  .page-header{
    margin-left: -8px;
  }
  .page-header .nav-link{
    color: var(--var-color-primary);
    position:relative;
    border-radius: 6px;
    margin-right: 5px;
  }
  .page-header .nav-link:hover{
    color: var(--var-color-primary);
    text-decoration: none;
    background-color: var(--var-bg-color-secondary);
    opacity: 0.8;
  }
  .page-header .nav-link.router-link-exact-active{
    color: var(--var-color-primary);
    text-decoration: none;
    background-color: var(--var-bg-color-secondary);
    opacity: 0.8;
  }
  .card{
    border-color: var(--var-border-color);
    margin: 7px;
  }
  .card p{
    margin-bottom: 0;
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
  .additional{
    color: #a0a1a7;
    font-size: 0.7rem;
  }
  .legend-item {
    border: 0 none;
    font-weight: 600;
  }
  .response.icon{
    vertical-align: middle;
    font-size: 0.5rem;
  }
  .dataTable.selectable{
    cursor: pointer;
  }
</style>
<style>
.subject {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 250px;
  width: 250px;
}
</style>
