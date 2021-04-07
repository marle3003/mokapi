<template>
  <div class="dashboard" v-if="dashboard !== null">
      <div class="page-header">
          <h2>Dashboard</h2>
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
          <b-card body-class="info-body" class="text-center">
            <b-card-title class="info">Total REST Requests</b-card-title>
            <b-card-text class="text-center value">{{ dashboard.totalRequests }}</b-card-text>
          </b-card>
          <b-card body-class="info-body" class="text-center">
            <b-card-title class="info">Web Request Errors</b-card-title>
            <b-card-text class="text-center value" v-bind:class="{'text-danger': hasErrors}">{{ dashboard.requestsWithError }}</b-card-text>
          </b-card>
        </b-card-group>

        <b-card-group deck>
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

        <b-card-group deck>
          <b-card body-class="info-body" class="text-center">
             <b-card-title class="info">Kafka Topics</b-card-title>
             <b-table :items="topics" :fields="topicFields" table-class="dataTable">
              <template v-slot:cell(method)="data">
                <b-badge pill class="operation" :class="data.item.method.toLowerCase()" >{{ data.item.method }}</b-badge>
              </template>
              <template v-slot:cell(lastRecord)="data">
                {{ data.item.lastRecord | moment}}
              </template>
              <template v-slot:cell(size)="data">
                {{ data.item.size | prettyBytes}}
              </template>
            </b-table>
          </b-card>
        </b-card-group>

        <!-- <b-card-group deck>
          <b-card body-class="info-body">
           <b-card-title class="info">Topic Memory Usage</b-card-title>
           <b-card-text class="text-center value">
             <time-chart :chart-data="chartTopicSize" :height="250" />
           </b-card-text>
          </b-card>
        </b-card-group> -->

        <!-- <b-card-group deck>
          <b-card>
            <b-card-title class="info">Services</b-card-title>
            <b-card-text v-if="loaded">
              <div class="row">
              <div class="col-5 p-0">
            <doughnut-chart :height="250" :chartdata="serviceStatus" />
              </div>
              <div class="col-7 p-0 align-self-center">
                <b-list-group>
                  <b-list-group-item class="legend-item p-0 pb-2">
                    <span class="mr-auto">Success</span>
                    <b-badge>{{ dashboard.serviceStatus.total - dashboard.serviceStatus.errors }}</b-badge>
                  </b-list-group-item>
                  <b-list-group-item class="legend-item p-0 pb-2">
                    <span class="mr-auto">Errors</span>
                    <b-badge>{{ dashboard.serviceStatus.errors }}</b-badge>
                  </b-list-group-item>
                </b-list-group>
              </div>
              </div>
            </b-card-text>
          </b-card>
        </b-card-group> -->

        <b-card-group deck>
          <b-card class="w-100">
            <b-card-title class="info text-center">Last Request Errors</b-card-title>
            <b-table :items="lastErrors" :fields="lastRequestField" table-class="dataTable">
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

        <b-card-group deck>
          <b-card class="w-100">
            <b-card-title class="info text-center">Recent Request</b-card-title>
            <b-table :items="lastRequests" :fields="lastRequestField" table-class="dataTable">
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
        'error',
        'time',
        'responseTime',
        ],
      topicSizes: {},
      chartTopicSize: {},
      serviceFields: [{key: 'name', class: 'text-left'}, 'lastRequest', 'requests', 'errors'],
      topicFields: [{key: 'name', class: 'text-left'}, 'count', 'size', 'lastRecord', 'partitions', 'segments']
    }
  },
  created () {
    this.getData()
    this.timer = setInterval(this.getData, 2000)
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
    hasErrors: function() {
      return this.dashboard.lastErrors !== undefined && this.dashboard.lastErrors.length > 0;
    },
    lastErrors: function() {
      if (this.dashboard.lastErrors === undefined) {
        return null
      }
      return this.dashboard.lastErrors.reverse()
    },
    lastRequests: function() {
      if (this.dashboard.lastRequests === undefined) {
        return null
      }
      return this.dashboard.lastRequests.reverse()
    },
    services: function() {
      const services = this.dashboard.services;
      if (services === undefined){
        return null
      }

      function compare(s1, s2) {
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

      return services.sort(compare);
    },
    topics: function() {
      const topics = this.dashboard.kafkaTopics;
      if (topics === undefined || topics === null){
        return null
      }

      function compare(s1, s2) {
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

      return topics.sort(compare);
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
      let ms = Math.round(time/1000000)
      let d = moment.duration(ms)
      if (d.seconds() < 1) {
        return d.milliseconds() + " [ms]"
      }else if (d.minutes() < 1) {
        return d.seconds() +" [sec]"
      }
      return moment.duration(d).minutes()
    },
    prettyBytes: function (num){
      // jacked from: https://github.com/sindresorhus/pretty-bytes
      if (typeof num !== 'number' || isNaN(num)) {
        return 0
      }

      var exponent;
      var unit;
      var neg = num < 0;
      var units = ['B', 'kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

      if (neg) {
        num = -num;
      }

      if (num < 1) {
        return (neg ? '-' : '') + num + ' B';
      }

      exponent = Math.min(Math.floor(Math.log(num) / Math.log(1000)), units.length - 1);
      num = (num / Math.pow(1000, exponent)).toFixed(2) * 1;
      unit = units[exponent];

      return (neg ? '-' : '') + num + ' ' + unit;
    }
  },
  methods: {
    async getData () {
      this.dashboard = await this.getDashboard()
      // var topicSizes = this.dashboard.kafka.topicSizes
      // var now = new Date()
      // for (var key in topicSizes){
      //   if (!(key in this.topicSizes)){
      //     this.topicSizes[key] = []
      //   }
      //   this.topicSizes[key].push({x: now, y: topicSizes[key]})
      //   if (this.topicSizes[key].length > 1000) {
      //     this.topicSizes[key].shift()
      //   }
      // }
      // var datasets = []
      // for (var key in this.topicSizes){
      //   datasets.push({
      //     label: key,
      //     data: this.topicSizes[key],
      //     fill: false,
      //     borderColor: '#34b5e3',
      //     pointRadius: 0,
      //     lineTension: 0
      //   })
      // }
      // this.chartTopicSize = {
      //   datasets: datasets
      // }
      this.loaded = true
    },
  getRandomInt () {
      return Math.floor(Math.random() * (50 - 5 + 1)) + 5
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
    margin: auto;
    margin-top: 42px;
  }
  .page-header h2{
    font-weight: 700;
  }
  .card{
    margin: 7px;
  }
  .card p{
    margin-bottom: 0;
  }
  .info{
    font-size: 0.7rem;
    font-weight: 500;
  }
  .info-body{
    padding: 0.8rem;
  }
  .value {
    font-size: 1.8rem;
    font-weight: 600;
  }
  .additional{
    color: #a0a1a7;
    font-size: 0.8rem;
  }
  .legend-item {
    border: 0 none;
    font-weight: 600;
  }
  .response.icon{
    vertical-align: middle;
    font-size: 0.5rem;
  }
</style>
