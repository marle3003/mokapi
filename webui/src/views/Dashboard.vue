<template>
  <div>
    <div v-if="error !== null">
      <b-alert
        show
        variant="danger"
      >{{ error }}</b-alert>
    </div>
    <div
      class="dashboard"
      v-if="loaded"
    >
      <div class="page-header">
        <b-navbar class="p-0">
          <b-navbar-nav>
            <b-nav-item :to="{ name: 'dashboard', query: {refresh: this.$route.query.refresh} }">Overview</b-nav-item>
            <b-nav-item
              :to="{ name: 'http', query: {refresh: this.$route.query.refresh} }"
              v-if="httpEnabled"
            >HTTP</b-nav-item>
            <b-nav-item
              :to="{ name: 'kafka', query: {refresh: this.$route.query.refresh} }"
              v-if="kafkaEnabled"
            >Kafka</b-nav-item>
            <b-nav-item
              :to="{ name: 'smtp', query: {refresh: this.$route.query.refresh} }"
              v-if="smtpEnabled"
            >SMTP</b-nav-item>
          </b-navbar-nav>
        </b-navbar>
      </div>
      <div class="page-body">

        <b-card-group deck>
          <b-card
            body-class="info-body"
            class="text-center"
          >
            <b-card-title class="info">Uptime Since</b-card-title>
            <b-card-text class="text-center value">{{ startTime | fromNow }}</b-card-text>
            <b-card-text class="text-right additional">{{ startTime | moment }}</b-card-text>
          </b-card>
          <b-card
            body-class="info-body"
            class="text-center"
          >
            <b-card-title class="info">Memory Usage</b-card-title>
            <b-card-text class="text-center value">{{ memoryUsage | prettyBytes }}</b-card-text>
          </b-card>
        </b-card-group>

        <b-card-group deck>
          <b-card
            body-class="info-body"
            class="text-center"
            v-if="httpEnabled"
          >
            <b-card-title class="info">Total HTTP Requests</b-card-title>
            <b-card-text class="text-center value">{{ this.metric(this.httpServices, 'http_requests_total') }}</b-card-text>
          </b-card>
          <b-card
            body-class="info-body"
            class="text-center"
            v-if="httpEnabled"
          >
            <b-card-title class="info">HTTP Request Errors</b-card-title>
            <b-card-text
              class="text-center value"
              v-bind:class="{'text-danger': this.metric(this.httpServices, 'http_requests_total_errors') > 0}"
            >{{ metric('http_requests_total_errors') }}</b-card-text>
          </b-card>
          <b-card
            body-class="info-body"
            class="text-center"
            v-if="kafkaEnabled"
          >
            <b-card-title class="info">Received Kafka Messages</b-card-title>
            <b-card-text class="text-center value">{{ this.totalKafkaMessages }}</b-card-text>
          </b-card>
          <b-card
            body-class="info-body"
            class="text-center"
            v-if="smtpEnabled"
          >
            <b-card-title class="info">Total Mails</b-card-title>
            <!-- <b-card-text class="text-center value">{{ dashboard.totalMails }}</b-card-text> -->
          </b-card>
        </b-card-group>

        <b-card-group
          deck
          v-show="httpEnabled && ($route.name === 'http' || $route.name === 'dashboard')"
        >
          <b-card
            body-class="info-body"
            class="text-center"
          >
            <b-card-title class="info">HTTP Services</b-card-title>
            <b-table
              :items="httpServices"
              :fields="httpFields"
              table-class="dataTable"
            >
              <template v-slot:cell(method)="data">
                <b-badge
                  pill
                  class="operation"
                  :class="data.item.method.toLowerCase()"
                >{{ data.item.method }}</b-badge>
              </template>
              <template v-slot:cell(lastRequest)="data">
                <span v-if="data.item.lastRequest === 0">-</span>
                <span v-else>{{ data.item.lastRequest | moment}}</span>
              </template>
            </b-table>
          </b-card>
        </b-card-group>

        <div v-show="kafkaEnabled && ($route.name === 'dashboard' || $route.name === 'kafka')">
          <b-card-group deck>
            <b-card
              body-class="info-body"
              class="text-center"
            >
              <b-card-title class="info">Kafka Clusters</b-card-title>
              <b-table
                :items="kafkaServices"
                :fields="kafkaFields"
                table-class="dataTable selectable"
                @row-clicked="kafkaClickHandler"
              >
                <template v-slot:cell(topics)="data">
                  <span>{{ data.item.topics.join(', ') }}</span>
                </template>
                <template v-slot:cell(lastMessage)="data">
                  <span>{{ metric(data.item.metrics, 'kafka_message_timestamp') | moment}}</span>
                </template>
                <template v-slot:cell(messages)="data">
                  <span>{{ metric(data.item.metrics, 'kafka_messages_total') }}</span>
                </template>
              </b-table>
            </b-card>
          </b-card-group>
        </div>

        <http-overview v-show="$route.name === 'http'" />

        <b-card-group
          deck
          v-show="$route.name === 'smtp'"
        >
          <b-card class="w-100">
            <b-card-title class="info text-center">Recent Mails</b-card-title>
            <b-table
              hover
              :items="lastMails"
              :fields="lastMailField"
              class="dataTable selectable"
              @row-clicked="mailClickHandler"
            >
              <template v-slot:cell(from)="data">
                <div
                  v-for="from in data.item.from"
                  :key="from.Address"
                >
                  <span v-if="from.Name !== ''">{{ from.Name }} &lt;</span><span>{{ from.Address }}</span><span v-if="from.Name !== ''">&gt;</span>
                </div>
              </template>
              <template v-slot:cell(to)="data">
                <div
                  v-for="to in data.item.to"
                  :key="to.Address"
                >
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
import Filters from '@/mixins/Filters'
import Refresh from '@/mixins/Refresh'
import Metrics from '@/mixins/Metrics'
import DoughnutChart from '@/components/DoughnutChart'
import TimeChart from '@/components/TimeChart'
import HttpOverview from '@/components/dashboard/HttpOverview'

export default {
  components: {
    'doughnut-chart': DoughnutChart,
    'time-chart': TimeChart,
    'http-overview': HttpOverview
  },
  mixins: [Api, Filters, Refresh, Metrics],
  data () {
    return {
      startTime: 0,
      memoryUsage: 0,
      httpServices: null,
      kafkaServices: null,
      lastRequests: null,
      lastErrors: null,
      lastMails: null,
      loaded: false,
      topicSizes: {},
      chartTopicSize: {},
      httpFields: [
        { key: 'name', class: 'text-left' },
        { key: 'lastRequest', class: 'text-left' },
        'requests',
        'errors'
      ],
      kafkaFields: [
        { key: 'name', class: 'text-left' },
        'topics',
        'lastMessage',
        'messages',
        'errors'
      ],
      lastMailField: [
        'from',
        'to',
        { key: 'subject', class: 'subject' },
        'time'
      ],
      error: null
    }
  },
  computed: {
    httpEnabled: function () {
      return this.httpServices !== null && this.httpServices.length > 0
    },
    kafkaEnabled: function () {
      return this.kafkaServices !== null && this.kafkaServices.length > 0
    },
    smtpEnabled: function () {
      return false // todo
    },
    serviceStatus: function () {
      let serviceStatus = this.dashboard.serviceStatus
      let success = serviceStatus.total - serviceStatus.errors
      return {
        datasets: [
          {
            data: [success, serviceStatus.errors],
            backgroundColor: ['rgb(110, 181, 110)', 'rgb(186, 86, 86)']
          }
        ],
        labels: ['Success', 'Errors']
      }
    },
    hasErrors: function () {
      return (
        this.dashboard.lastErrors !== undefined &&
        this.dashboard.lastErrors.length > 0
      )
    },
    totalKafkaMessages: function () {
      if (!this.kafkaServices) {
        return 0
      }
      let sum = 0
      for (let service of this.kafkaServices) {
        sum += this.metric(service.metrics, 'kafka_messages_total')
      }
      return sum
    }
  },
  methods: {
    async getData () {
      this.getHttpServices().then(
        r => {
          this.httpServices = r
          this.error = null
        },
        r => {
          this.httpServices = null
          this.error = r
        }
      )
      this.getKafkaServices().then(
        r => {
          this.kafkaServices = r
          this.error = null
        },
        r => {
          this.kafkaServices = null
          this.error = r
        }
      )
      this.getMetrics('app_start_timestamp', 'app_memory_usage_bytes').then(
        data => {
          this.startTime = this.metric(data, 'app_start_timestamp')
          this.memoryUsage = this.metric(data, 'app_memory_usage_bytes')
        },
        _ => {
          this.startTime = 0
          this.memoryUsage = 0
        }
      )
      this.loaded = true
    },
    mailClickHandler (record) {
      this.$router.push({ name: 'smtpMail', params: { id: record.id } })
    },
    kafkaClickHandler (record) {
      this.$router.push({
        name: 'kafkaCluster',
        params: { cluster: record.name },
        query: { refresh: '5' }
      })
    }
  }
}
</script>

<style scoped>
.dashboard {
  width: 90%;
  margin: 12px auto auto;
}
.page-header {
  margin-left: -8px;
}
.page-header .nav-link {
  color: var(--var-color-primary);
  position: relative;
  border-radius: 6px;
  margin-right: 5px;
}
.page-header .nav-link:hover {
  color: var(--var-color-primary);
  text-decoration: none;
  background-color: var(--var-bg-color-secondary);
  opacity: 0.8;
}
.page-header .nav-link.router-link-exact-active {
  color: var(--var-color-primary);
  text-decoration: none;
  background-color: var(--var-bg-color-secondary);
  opacity: 0.8;
}
.card {
  border-color: var(--var-border-color);
  margin: 7px;
}
.card p {
  margin-bottom: 0;
}
.info {
  font-size: 0.7rem;
  font-weight: 300;
}
.info-body {
  padding: 0.8rem;
}
.value {
  font-size: 2.25rem;
  font-weight: 300;
}
.additional {
  color: #a0a1a7;
  font-size: 0.7rem;
}
.legend-item {
  border: 0 none;
  font-weight: 600;
}
.response.icon {
  vertical-align: middle;
  font-size: 0.5rem;
}
.dataTable.selectable {
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
