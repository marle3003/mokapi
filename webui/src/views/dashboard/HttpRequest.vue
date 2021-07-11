<template>
  <div class="dashboard" v-if="request !== null">
    <div class="page-header">
      <h2>HTTP Request</h2>
    </div>
    <div class="page-body">
      <b-card class="w-100">
        <b-row>
          <b-col>
            <p class="label">URL</p>
            <p>{{ request.url}}</p>
            <p class="label">Method</p>
            <p><b-badge pill class="operation" :class="request.method.toLowerCase()" >{{ request.method }}</b-badge></p>
            <p class="label">Status</p>
            <p>
              <b-badge pill variant="success" v-if="request.httpStatus >= 200 && request.httpStatus < 300">{{ request.httpStatus }} {{ request.httpStatus | reason }}</b-badge>
              <b-badge pill variant="warning" v-if="request.httpStatus >= 300 && request.httpStatus < 400">{{ request.httpStatus }} {{ request.httpStatus | reason }}</b-badge>
              <b-badge pill class="client-error" v-if="request.httpStatus >= 400 && request.httpStatus < 500">{{ request.httpStatus }} {{ request.httpStatus | reason }}</b-badge>
              <b-badge pill variant="danger" v-if="request.httpStatus >= 500 && request.httpStatus < 600">{{ request.httpStatus }} {{ request.httpStatus | reason }}</b-badge>
            </p>
          </b-col>
          <b-col>
            <p class="label">Content Type</p>
            <p>{{ request.contentType }}</p>
            <p class="label">Time</p>
            <p> {{ request.time | moment}}</p>
            <p class="label">Duration</p>
            <p> {{ request.responseTime | duration}}</p>
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <p class="label">Request Parameters</p>
            <b-table small hover class="dataTable" :items="request.parameters" :fields="parameters">
              <template v-slot:cell(show_details)="row">
                <div @click="toggleDetails(row)" v-if="showRawParameter(row.item)">
                  <b-icon v-if="row.detailsShowing" icon="dash-square"></b-icon>
                  <b-icon v-else icon="plus-square"></b-icon>
                </div>
              </template>
              <template v-slot:cell(value)="data">
                {{ data.item.value !== '' ? data.item.value : data.item.raw }}
              </template>
              <template v-slot:cell(openapi)="data">
                {{ data.item.value !== '' ? 'yes' : 'no' }}
              </template>
              <template v-slot:row-details="row">
                <b-card class="w-100">
                  <b-row class="mb-2">
                    <b-col sm="3" class="text-sm-right"><b>Raw:</b></b-col>
                    <b-col><vue-simple-markdown :source="row.item.raw" /></b-col>
                  </b-row>
                </b-card>
              </template>
            </b-table>

            <action :actions="request.actions"></action>

            <p class="label">Response Body</p>
            <pre :class="getLanguage(request.contentType)"><code :class="getLanguage(request.contentType)" v-html="pretty(request.responseBody, request.contentType)"></code></pre>
          </b-col>
        </b-row>
      </b-card>
    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import moment from 'moment'
import http from 'http-status-codes'
import Action from '@/components/Action'
import xmlFormatter from 'xml-formatter'

export default {
  name: 'HttpRequest',
  components: {
    'action': Action
  },
  mixins: [Api],
  data () {
    return {
      request: null,
      parameters: [{key: 'show_details', label: '', thStyle: 'width: 1%'}, 'name', 'value', 'type', {key: 'openapi', label: 'OpenApi'}],
      detailsShown: []
    }
  },
  created () {
    this.getData()
  },
  methods: {
    async getData () {
      let id = this.$route.params.id
      this.request = await this.getHttpRequest(id)
    },
    showRawParameter (p) {
      return p.value !== '' && p.value !== p.raw
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
    pretty (s, contentType) {
      switch (contentType) {
        case 'application/json':
          s = JSON.stringify(JSON.parse(s), null, 2)
          // eslint-disable-next-line no-undef
          return Prism.highlight(s, Prism.languages.json, 'json')
        case 'text/xml':
        case 'application/xml':
          s = xmlFormatter(s)
          // eslint-disable-next-line no-undef
          return Prism.highlight(s, Prism.languages.xml, 'xml')
      }
      return s
    },
    getLanguage (contentType) {
      // https://lucidar.me/en/web-dev/list-of-supported-languages-by-prism/
      switch (contentType) {
        case 'application/json':
          return 'language-json'
        case 'text/xml':
        case 'application/xml':
          return 'language-xml'
      }
      return ''
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
    reason: function (status) {
      return http.getStatusText(status)
    }
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
  .menu-item svg {
    -moz-transition: all .3s linear;
    -webkit-transition: all .3s linear;
    transition: all .3s linear;
  }
  .not-collapsed svg {
    -moz-transform:rotate(90deg);
    -webkit-transform:rotate(90deg);
    transform:rotate(90deg);
  }
</style>
