<template>
  <div v-show="httpEnabled">
    <b-card-group
        deck
        v-show="($route.name === 'http' || $route.name === 'dashboard')"
    >
        <b-card class="text-center">
        <b-card-title class="info">HTTP Services</b-card-title>
        <b-table
            :items="httpServices"
            :fields="serviceFields"
            table-class="dataTable selectable"
            @row-clicked="httpClickHandler"
        >
            <template v-slot:cell(method)="data">
            <b-badge
                pill
                class="operation"
                :class="data.item.method.toLowerCase()"
            >{{ data.item.method }}</b-badge>
            </template>
            <template v-slot:cell(lastRequest)="data">
            <span>{{ metric(data.item.metrics, 'http_request_timestamp') | moment}}</span>
            </template>
            <template v-slot:cell(requests)="data">
            <span>{{ metric(data.item.metrics, 'http_requests_total') }}</span>
            </template>
            <template v-slot:cell(errors)="data">
            <span>{{ metric(data.item.metrics, 'http_requests_errors_total') }}</span>
            </template>
        </b-table>
        </b-card>
    </b-card-group>
    <div v-if="service !== null">
        <b-card-group deck>
            <b-card class="w-100">
                <b-row>
                    <b-col>
                        <p style="font-size: 1.25rem; font-weight: 500;">{{ service.name }}</p>
                    </b-col>
                    <b-col>
                        <p class="label">Version</p>
                        <p>{{ service.version }}</p>
                    </b-col>
                    <b-col>
                        <p class="label">Contact</p>
                        <p v-if="service.contact">
                            <a :href="service.contact.url">{{ service.contact.name }}</a>
                            <a :href="'mailto:' + service.contact.email" v-if="service.contact.email">
                                <b-icon-envelope></b-icon-envelope>
                            </a>
                        </p>
                    </b-col>
                </b-row>
                <b-row v-if="service.description !== ''">
                    <b-col>
                        <p class="label">Description</p>
                        <vue-simple-markdown :source="service.description" />
                    </b-col>
                </b-row>
            </b-card>
        </b-card-group>
        <b-card-group deck v-if="path !== null">
            <b-card class="w-100">
                <b-row>
                    <b-col>
                        <p style="font-size: 1.25rem; font-weight: 500;">{{ path.path }}</p>
                    </b-col>
                </b-row>
                <b-row v-if="path.summary">
                    <b-col>
                        <p class="label">Summary</p>
                        <vue-simple-markdown :source="path.summary" />
                    </b-col>
                </b-row>
                <b-row v-if="path.description">
                    <b-col>
                        <p class="label">Description</p>
                        <vue-simple-markdown :source="path.description" />
                    </b-col>
                </b-row>
                <b-row>
                    <b-col>
                        <p class="label">Methods</p>
                        <span v-for="( operation, index ) in path.operations" :key="index" class="mr-1">
                            <b-badge pill class="operation" :class="operation.method" >{{ operation.method }}</b-badge>
                        </span>
                    </b-col>
                </b-row>
            </b-card>
        </b-card-group>
        <b-card-group deck>
            <b-card
                body-class=".metric-card"
                class="text-center"
            >
                <b-card-title class="info">Requests</b-card-title>
                <b-card-text class="text-center value">{{ totalHttpRequests }}</b-card-text>
            </b-card>
            <b-card
                body-class=".metric-card"
                class="text-center"
            >
                <b-card-title class="info">Errors</b-card-title>
                <b-card-text 
                    class="text-center value"
                    v-bind:class="{'text-danger': totalHttpRequestErrors > 0}">
                    {{ totalHttpRequestErrors }}
                </b-card-text>
            </b-card>
        </b-card-group>

        <b-card-group deck v-show="$route.name === 'httpService2'">
            <b-card class="w-100">
                <b-card-title class="info text-center">Endpoints</b-card-title>
                <b-table
                small
                hover
                class="dataTable selectable"
                :items="endpoints"
                :fields="endpointsFields"
                style="table-layout: fixed"
                @row-clicked="endpointClickHandler"
                >
                <template v-slot:cell(operations)="data">
                    <span v-for="( operation, index ) in data.value" :key="index" class="mr-1">
                    <b-badge pill class="operation" :class="operation.method" >{{ operation.method }}</b-badge>
                    </span>
                </template>
                </b-table>
            </b-card>
        </b-card-group>
    </div>

    <http-requests v-show="showRequests" />
    <http-request v-show="$route.name === 'httpRequest'" />
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import Refresh from '@/mixins/Refresh'
import Metrics from '@/mixins/Metrics'
import Shortcut from '@/mixins/Shortcut'

import Header from '@/components/dashboard/Header'

import HttpRequests from '@/components/http/Requests'
import HttpRequest from '@/components/http/Request'

export default {
  name: 'HttpDashboard',
  mixins: [Api, Filters, Refresh, Metrics, Shortcut],
  props: ["services"],
  components: {
    'dashboard-header': Header,
    'http-requests': HttpRequests,
    'http-request': HttpRequest
  },
  data () {
    return {
      service: null,
      serviceFields: [
        { key: 'name', class: 'text-left' },
        { key: 'lastRequest', class: 'text-left' },
        'requests',
        'errors'
      ],
      endpointsFields: ['path', 'operations']
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
    httpEnabled: function () {
      return this.httpServices.length > 0
    },
    endpoints: function () {
      let result = this.service.paths.sort(this.compareEndpoint)
      return result
    },
    path: function() {
        const s = this.$route.params.path
        if (!this.service || !s){
            return null
        }
        const p = decodeURIComponent(s)
        for (let path of this.service.paths){
            if (path.path === p) {
                return path
            }
        }
        return null
    },
    showRequests: function() {
        return this.$route.name === 'http' || this.$route.name === 'httpService2' || this.$route.name === 'httpPath'
    },
    totalHttpRequests: function () {
      let labels = []

      const path = this.$route.params.path
      if (path && path !== '') {
        labels.push({name: 'endpoint', value: this.path.path})
      }

      return this.metric(this.service.metrics, 'http_requests_total', ...labels)
    },
    totalHttpRequestErrors: function () {
      let labels = []

      const path = this.$route.params.path
      if (path && path !== '') {
        labels.push({name: 'endpoint', value: this.path.path})
      }

      return this.metric(this.service.metrics, 'http_requests_errors_total', ...labels)
    },
  },
  methods: {
    async getData () {
      const name = this.$route.params.service
      if (!name) {
        this.service = null
        return
      }

      this.$http.get(this.baseUrl + '/api/services/http/' + name).then(
        r => {
          this.service = r.data
        },
        r => {
          this.service = null
        }
      )
    },
    httpClickHandler (record) {
      this.$router.push({
        name: 'httpService2',
        params: { service: record.name },
        query: { refresh: '5' }
      })
    },
    endpointClickHandler (record) {
      this.$router.push({
        name: 'httpPath',
        params: { service: this.service.name, path: record.path },
        query: { refresh: '5' }
      })
    },
    compareEndpoint (x, y) {
      return x.path.localeCompare(y.path)
    }
  }
}
</script>

<style scoped>
.dashboard {
  width: 90%;
  margin: 42px auto auto;
}
.page-header h2 {
  font-weight: 400;
}
.card {
  border-color: var(--var-border-color);
  margin: 7px;
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
</style>
<style>
.kafka-message {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 50%;
  width: 50%;
}
</style>
