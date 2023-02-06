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
      <dashboard-header />
      <div class="page-body">
        <b-card-group deck v-show="$route.name === 'dashboard'">
          <b-card
            body-class="metric-card"
            class="text-center"
          >
            <b-card-title class="info">Uptime Since</b-card-title>
            <b-card-text class="text-center value">{{ metric(metrics, 'app_start_timestamp') | fromNow }}</b-card-text>
            <b-card-text class="text-right additional">{{ metric(metrics, 'app_start_timestamp') | moment }}</b-card-text>
          </b-card>
          <b-card
            body-class="metric-card"
            class="text-center"
          >
            <b-card-title class="info">Memory Usage</b-card-title>
            <b-card-text class="text-center value">{{ metric(metrics, 'app_memory_usage_bytes') | prettyBytes }}</b-card-text>
          </b-card>
        </b-card-group>

        <b-card-group deck v-show="$route.name === 'dashboard' || $route.meta.showMetrics">
          <b-card
            body-class="metric-card"
            class="text-center"
            v-if="httpEnabled && $route.name === 'dashboard' || $route.name === 'http'"
          >
            <b-card-title class="info">Total HTTP Requests</b-card-title>
            <b-card-text class="text-center value">{{ totalHttpRequests }}</b-card-text>
          </b-card>
          <b-card
            body-class="metric-card"
            class="text-center"
            v-if="httpEnabled && $route.name === 'dashboard' || $route.name === 'http'"
          >
            <b-card-title class="info">HTTP Request Errors</b-card-title>
            <b-card-text
              class="text-center value"
              v-bind:class="{'text-danger': totalHttpRequestErrors > 0}"
            >
            {{ totalHttpRequestErrors }}
            </b-card-text>
          </b-card>
          <b-card
            body-class="metric-card"
            class="text-center"
            v-if="kafkaEnabled && $route.name === 'dashboard' || $route.name === 'kafka'"
          >
            <b-card-title class="info">Kafka Messages</b-card-title>
            <b-card-text class="text-center value">{{ this.totalKafkaMessages }}</b-card-text>
          </b-card>
          <b-card
            body-class="metric-card"
            class="text-center"
            v-if="smtpEnabled && $route.name === 'dashboard' || $route.name === 'smtp'"
          >
            <b-card-title class="info">Total Mails</b-card-title>
            <b-card-text class="text-center value">{{ this.totalSmtpMails }}</b-card-text>
          </b-card>
        </b-card-group>

        <http-dashboard :services="services" v-show="httpEnabled" />
        <kafka-dashboard :services="services" v-show="kafkaEnabled" />

        <smtp-services :services="smtpServices" v-show="smtpEnabled && ($route.name === 'dashboard' || $route.name === 'smtp')" />
        <smtp-mails v-show="$route.name === 'smtp'" />

      </div>
    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import Refresh from '@/mixins/Refresh'
import Metrics from '@/mixins/Metrics'
import Shortcut from '@/mixins/Shortcut'

import Header from '@/components/dashboard/Header'

import Http from '@/components/http/Dashboard'
import Kafka from '@/components/kafka/Dashboard'

import SmtpServices from '@/components/smtp/Services'
import SmtpMails from '@/components/smtp/Mails'

export default {
  mixins: [Api, Filters, Refresh, Metrics, Shortcut],
  components: {
    'dashboard-header': Header,
    'http-dashboard': Http,
    'kafka-dashboard': Kafka,
    'smtp-services': SmtpServices,
    'smtp-mails': SmtpMails
  },
  data () {
    return {
      metrics: [],
      services: null,
      loaded: false,
      error: null
    }
  },
  computed: {
    httpServices: function () {
      let result = []
      if (!this.services) {
        return result
      }
      for (let service of this.services) {
        if (service.type === 'http') {
          result.push(service)
        }
      }
      return result
    },
    kafkaServices: function () {
      let result = []
      if (!this.services) {
        return result
      }
      for (let service of this.services) {
        if (service.type === 'kafka') {
          service.topics = service.topics.sort()
          result.push(service)
        }
      }
      return result
    },
    smtpServices: function () {
      let result = []
      if (!this.services) {
        return result
      }
      for (let service of this.services) {
        if (service.type === 'smtp') {
          result.push(service)
        }
      }
      return result
    },
    httpEnabled: function () {
      return this.httpServices !== null && this.httpServices.length > 0
    },
    kafkaEnabled: function () {
      return this.kafkaServices !== null && this.kafkaServices.length > 0
    },
    smtpEnabled: function () {
      return this.smtpServices !== null && this.smtpServices.length > 0
    },
    totalHttpRequests: function () {
      let n = 0
      for (let http of this.httpServices) {
        n += this.metric(http.metrics, 'http_requests_total')
      }
      return n
    },
    totalHttpRequestErrors: function () {
      let n = 0
      for (let http of this.httpServices) {
        n += this.metric(http.metrics, 'http_requests_errors_total')
      }
      return n
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
    },
    totalSmtpMails: function () {
      if (!this.smtpServices) {
        return 0
      }
      let sum = 0
      for (let service of this.smtpServices) {
        sum += this.metric(service.metrics, 'mails_total')
      }
      return sum
    }
  },
  methods: {
    async getData () {
      this.getServices().then(
        r => {
          this.services = r
          this.error = null
        },
        r => {
          this.services = null
          this.error = r
        }
      )
      this.getMetrics('app').then(
        r => {
          this.metrics = r
          this.error = null
        },
        r => {
          this.metrics = null
          this.error = r
        }
      )
      this.loaded = true
    },
    shortcut (e) {
      let cmd = e.key.toLowerCase()
      if (cmd === 'escape' && this.$route.name !== 'dashboard') {
        this.$router.go(-1)
      }
    }
  }
}
</script>

<style scoped>
.dashboard {
  width: 90%;
  margin: 12px auto auto;
}
</style>
<style>
.dashboard .info {
  font-size: 0.7rem;
  font-weight: 300;
}
.dashboard .metric-card {
  padding: 0.8rem;
  margin-bottom: 0;
}
.dashboard .card {
  border-color: var(--var-border-color);
  margin: 7px;
}
.dashboard .value {
  font-size: 2.25rem;
  font-weight: 300;
}
.dashboard .additional {
  color: #a0a1a7;
  font-size: 0.7rem;
}
.dashboard .legend-item {
  border: 0 none;
  font-weight: 600;
}
.dashboard .response.icon {
  vertical-align: middle;
  font-size: 0.5rem;
}
.subject {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 250px;
  width: 250px;
}
</style>
