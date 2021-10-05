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
                <b-nav-item :to="{ name: 'http', query: {refresh: this.$route.query.refresh} }" v-if="dashboard.httpEnabled">HTTP</b-nav-item>
                <b-nav-item :to="{ name: 'kafka', query: {refresh: this.$route.query.refresh} }" v-if="dashboard.kafkaEnabled">Kafka</b-nav-item>
                <b-nav-item :to="{ name: 'smtp', query: {refresh: this.$route.query.refresh} }" v-if="dashboard.smtpEnabled">SMTP</b-nav-item>
              </b-navbar-nav>
            </b-navbar>
        </div>
        <div class="page-body">

          <b-card-group deck>
            <b-card body-class="info-body" class="text-center">
              <b-card-title class="info">Uptime Since</b-card-title>
              <b-card-text class="text-center value">{{ dashboard.serverUptime | fromNow }}</b-card-text>
              <b-card-text class="text-right additional">{{ dashboard.serverUptime | moment }}</b-card-text>
            </b-card>
            <b-card body-class="info-body" class="text-center">
              <b-card-title class="info">Memory Usage</b-card-title>
              <b-card-text class="text-center value">{{ dashboard.memoryUsage | prettyBytes }}</b-card-text>
              </b-card>
          </b-card-group>

          <b-card-group deck>
            <b-card body-class="info-body" class="text-center" v-if="dashboard.httpEnabled">
              <b-card-title class="info">Total HTTP Requests</b-card-title>
              <b-card-text class="text-center value">{{ dashboard.totalRequests }}</b-card-text>
            </b-card>
            <b-card body-class="info-body" class="text-center" v-if="dashboard.httpEnabled">
              <b-card-title class="info">HTTP Request Errors</b-card-title>
              <b-card-text class="text-center value" v-bind:class="{'text-danger': hasErrors}">{{ dashboard.requestsWithError }}</b-card-text>
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

          <b-card-group deck v-show="dashboard.httpEnabled && $route.name === 'http' || $route.name === 'dashboard'">
            <b-card body-class="info-body" class="text-center">
               <b-card-title class="info">REST Services</b-card-title>
               <b-table :items="services" :fields="serviceFields" table-class="dataTable">
                <template v-slot:cell(method)="data">
                  <b-badge pill class="operation" :class="data.item.method.toLowerCase()" >{{ data.item.method }}</b-badge>
                </template>
                <template v-slot:cell(lastRequest)="data">
                  <span v-if="data.item.lastRequest === '0001-01-01T00:00:00Z'">-</span>
                  <span v-else>{{ data.item.lastRequest | moment}}</span>
                </template>
              </b-table>
            </b-card>
          </b-card-group>

          <div v-show="dashboard.kafkaEnabled && $route.name === 'dashboard' || $route.name === 'kafka'">
            <b-card-group deck>
              <b-card body-class="info-body" class="text-center">
                 <b-card-title class="info">Kafka Topics</b-card-title>
                 <b-table :items="topics" :fields="topicFields" table-class="dataTable selectable" @row-clicked="topicClickHandler">
                  <template v-slot:cell(lastRecord)="data">
                    {{ data.item.lastRecord | moment}}
                  </template>
                  <template v-slot:cell(size)="data">
                    {{ data.item.size | prettyBytes}}
                  </template>
                </b-table>
              </b-card>
            </b-card-group>
            <b-card-group deck>
              <b-card body-class="info-body" class="text-center">
                <b-card-title class="info">Kafka Groups</b-card-title>
                <b-table :items="groups" :fields="groupFields" table-class="dataTable">
                  <template v-slot:cell(members)="data">
                    {{ data.item.members.join(', ') }}
                  </template>
                </b-table>
              </b-card>
            </b-card-group>
          </div>

          <b-card-group deck v-show="$route.name === 'http'">
            <b-card class="w-100">
              <b-card-title class="info text-center">Last Request Errors</b-card-title>
              <b-table hover :items="lastErrors" :fields="lastRequestField" class="dataTable selectable" @row-clicked="requestClickHandler">
                <template v-slot:cell(method)="data">
                  <b-badge pill class="operation" :class="data.item.method.toLowerCase()" >{{ data.item.method }}</b-badge>
                </template>
                <template v-slot:cell(httpStatus)="data">
                  <b-icon icon="circle-fill" class="response icon mr-1" variant="success" v-if="data.item.httpStatus >= 200 && data.item.httpStatus < 300"></b-icon>
                  <b-icon icon="circle-fill" class="response icon mr-1" variant="warning" v-if="data.item.httpStatus >= 300 && data.item.httpStatus < 400"></b-icon>
                  <b-icon icon="circle-fill" class="response icon mr-1 client-error" v-if="data.item.httpStatus >= 400 && data.item.httpStatus < 500"></b-icon>
                  <b-icon icon="circle-fill" class="response icon mr-1" variant="danger" v-if="data.item.httpStatus >= 500 && data.item.httpStatus < 600"></b-icon>
                  {{ data.item.httpStatus }}
                </template>
                <template v-slot:cell(time)="data">
                  {{ data.item.time | moment}}
                </template>
                <template v-slot:cell(responseTime)="data">
                  {{ data.item.responseTime | duration}}
                </template>
              </b-table>
            </b-card>
          </b-card-group>

          <b-card-group deck v-show="$route.name === 'http'">
            <b-card class="w-100">
              <b-card-title class="info text-center">Recent Request</b-card-title>
              <b-table hover :items="lastRequests" :fields="lastRequestField" class="dataTable selectable" @row-clicked="requestClickHandler">
                <template v-slot:cell(method)="data">
                  <b-badge pill class="operation" :class="data.item.method.toLowerCase()" >{{ data.item.method }}</b-badge>
                </template>
                <template v-slot:cell(httpStatus)="data">
                  <b-icon icon="circle-fill" class="response icon mr-1" variant="success" v-if="data.item.httpStatus >= 200 && data.item.httpStatus < 300"></b-icon>
                  <b-icon icon="circle-fill" class="response icon mr-1" variant="warning" v-if="data.item.httpStatus >= 300 && data.item.httpStatus < 400"></b-icon>
                  <b-icon icon="circle-fill" class="response icon mr-1 client-error" v-if="data.item.httpStatus >= 400 && data.item.httpStatus < 500"></b-icon>
                  <b-icon icon="circle-fill" class="response icon mr-1" variant="danger" v-if="data.item.httpStatus >= 500 && data.item.httpStatus < 600"></b-icon>
                  {{ data.item.httpStatus }}
                </template>
                <template v-slot:cell(time)="data">
                  {{ data.item.time | moment}}
                </template>
                <template v-slot:cell(responseTime)="data">
                  {{ data.item.responseTime | duration}}
                </template>
              </b-table>
            </b-card>
          </b-card-group>

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

export default {
  components: {
    'doughnut-chart': DoughnutChart,
    'time-chart': TimeChart
  },
  mixins: [Api],
  data () {
    return {
      dashboard: null,
      loaded: false,
      timer: null,
      lastRequestField: [
        'method',
        {key: 'url', tdClass: 'break'},
        'httpStatus',
        'time',
        'responseTime'
      ],
      topicSizes: {},
      chartTopicSize: {},
      serviceFields: [{key: 'name', class: 'text-left'}, {key: 'lastRequest', class: 'text-left'}, 'requests', 'errors'],
      topicFields: [{key: 'name', class: 'text-left'}, 'count', 'size', 'lastRecord', 'partitions', 'segments'],
      groupFields: [{key: 'name', class: 'text-left'}, 'state', {key: 'assignmentStrategy', class: 'text-left'}, {key: 'coordinator', class: 'text-left'}, {key: 'leader', class: 'text-left'}, {key: 'members', class: 'text-left'}],
      lastMailField: ['from', 'to', {key: 'subject', class: 'subject'}, 'time'],
      error: null
    }
  },
  created () {
    this.init()
  },
  computed: {
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
    },
    lastErrors: function () {
      if (this.dashboard.lastErrors === undefined) {
        return null
      }
      // eslint-disable-next-line vue/no-side-effects-in-computed-properties
      return this.dashboard.lastErrors.reverse()
    },
    lastRequests: function () {
      if (this.dashboard.lastRequests === undefined) {
        return null
      }
      // eslint-disable-next-line vue/no-side-effects-in-computed-properties
      return this.dashboard.lastRequests.reverse()
    },
    lastMails: function () {
      if (this.dashboard.lastMails === undefined) {
        return null
      }
      // eslint-disable-next-line vue/no-side-effects-in-computed-properties
      return this.dashboard.lastMails.reverse()
    },
    services: function () {
      const services = this.dashboard.services
      if (services === undefined) {
        return null
      }

      function compare (s1, s2) {
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

      return services.sort(compare)
    },
    totalMessages: function () {
      const topics = this.dashboard.kafka.topics
      if (topics === undefined || topics === null) {
        return 0
      }
      let counter = 0
      for (const topic of topics) {
        counter += topic.count
      }
      return counter
    },
    topics: function () {
      if (this.dashboard.kafka.topics === undefined || this.dashboard.kafka.topics === null) {
        return null
      }

      let topics = []
      for (let i = 0; i < this.dashboard.kafka.topics.length; i++) {
        let topic = this.dashboard.kafka.topics[i]
        let item = {
          service: topic.service,
          name: topic.name,
          count: topic.count,
          lastRecord: topic.lastRecord,
          groups: topic.groups
        }
        item.partitions = topic.partitions.length
        for (let j = 0; j < topic.partitions.length; j++) {
          item.segments = topic.partitions[j].segments
        }
        topics.push(item)
      }

      function compare (s1, s2) {
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

      return topics.sort(compare)
    },
    groups: function () {
      const groups = this.dashboard.kafka.groups
      if (groups === undefined || groups === null) {
        return null
      }

      function compare (s1, s2) {
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

      return groups.sort(compare)
    }
  },
  filters: {
    moment: function (date) {
      return moment(date).local().format('YYYY-MM-DD HH:mm:ss')
    },
    fromNow: function (date) {
      return moment(date).fromNow(true)
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
      this.loaded = true
    },
    requestClickHandler (record) {
      this.$router.push({name: 'httpRequest', params: {id: record.id}})
    },
    mailClickHandler (record) {
      this.$router.push({name: 'smtpMail', params: {id: record.id}})
    },
    topicClickHandler (record) {
      this.$router.push({name: 'kafkaTopic', params: {kafka: record.service, topic: record.name}, query: {refresh: '5'}})
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
