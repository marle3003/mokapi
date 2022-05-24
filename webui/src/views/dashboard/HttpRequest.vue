<template>
  <div class="dashboard" v-if="request !== null">
    <div class="page-header">
      <b-row class="mb-2 ml-0 mr-0">
        <b-col cols="auto" class="mr-auto pl-0">
          <h2>HTTP Request</h2>
        </b-col>
        <b-col cols="auto" class="pr-0">
          <div class="close" @click="$router.go(-1)">
            <b-icon icon="x" class="border rounded p-1"></b-icon>
          </div>
        </b-col>
      </b-row>
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

            <workflows :workflows="request.workflows"></workflows>

            <p class="label">Response Body</p>
            <div v-if="parseError !== null">
              <b-alert show variant="danger">{{ parseError }}</b-alert>
            </div>
            <pre :class="getLanguage(request.contentType)"><code :class="getLanguage(request.contentType)" v-html="pretty(request.responseBody, request.contentType)"></code></pre>
          </b-col>
        </b-row>
      </b-card>
    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import http from 'http-status-codes'
import Workflows from '@/components/Workflows'
import xmlFormatter from 'xml-formatter'
import MIMEType from 'whatwg-mimetype'

export default {
  name: 'HttpRequest',
  components: {
    'workflows': Workflows
  },
  mixins: [Api, Filters],
  data () {
    return {
      request: null,
      parameters: [{key: 'show_details', label: '', thStyle: 'width: 1%'}, 'name', {key: 'value', tdClass: 'break'}, 'type', {key: 'openapi', label: 'OpenApi'}],
      detailsShown: [],
      parseError: null
    }
  },
  created () {
    this.getData()
    window.addEventListener('keyup', this.doCommand)
  },
  destroyed () {
    window.removeEventListener('keyup', this.doCommand)
  },
  methods: {
    async getData () {
      let id = this.$route.params.id
      this.request = await this.getHttpRequest(id)
    },
    showRawParameter (p) {
      return p.value !== '' && p.raw !== '' && p.value !== p.raw
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
      if (s === '' || s === null) {
        return s
      }
      const mimeType = new MIMEType(contentType)
      this.parseError = null
      switch (mimeType.subtype) {
        case 'json':
          try {
            s = JSON.stringify(JSON.parse(s), null, 2)
            // eslint-disable-next-line no-undef
            return Prism.highlight(s, Prism.languages.json, 'json')
          } catch (e) {
            this.parseError = e
            return s
          }
        case 'xml':
        case 'rss+xml':
          s = xmlFormatter(s)
          // eslint-disable-next-line no-undef
          return Prism.highlight(s, Prism.languages.xml, 'xml')
      }
      return s
    },
    getLanguage (contentType) {
      if (contentType === '' || contentType === null) {
        return ''
      }
      // https://lucidar.me/en/web-dev/list-of-supported-languages-by-prism/
      const mimeType = new MIMEType(contentType)
      switch (mimeType.subtype) {
        case 'json':
          return 'language-json'
        case 'xml':
        case 'rss+xml':
          return 'language-xml'
      }
      return ''
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
  .close {
    font-size: 2.2rem;
    cursor: pointer;
    border-color: var(--var-border-color);
    color: var(--var-color-primary);
  }
</style>
