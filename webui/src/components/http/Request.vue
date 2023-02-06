<template>
  <div v-if="request !== null">
    <b-card class="w-100 mb-3">
      <b-row>
        <b-col>
          <p class="label">URL</p>
          <p>{{ request.url}}</p>
          <p class="label">Method</p>
          <p><b-badge pill class="operation" :class="request.method.toLowerCase()" >{{ request.method }}</b-badge></p>
          <p class="label">Status</p>
          <p>
            <b-badge pill variant="success" v-if="response.statusCode >= 200 && response.statusCode < 300">{{ response.statusCode }} {{ response.statusCode | httpStatusText }}</b-badge>
            <b-badge pill variant="warning" v-if="response.statusCode >= 300 && response.statusCode < 400">{{ response.statusCode }} {{ response.statusCode | httpStatusText }}</b-badge>
            <b-badge pill class="client-error" v-if="response.statusCode >= 400 && response.statusCode < 500">{{ response.statusCode }} {{ response.statusCode | httpStatusText }}</b-badge>
            <b-badge pill variant="danger" v-if="response.statusCode >= 500 && response.statusCode < 600">{{ response.statusCode }} {{ response.statusCode | httpStatusText }}</b-badge>
          </p>
        </b-col>
        <b-col>
          <p class="label">Time</p>
          <p> {{ event.time | moment}}</p>
          <p class="label">Duration</p>
          <p> {{ event.data.duration | duration}}</p>
          <p class="label">Size</p>
          <p> {{ response.size | prettyBytes}}</p>
        </b-col>
      </b-row>
    </b-card>
    <b-card class="mb-3">
      <b-row>
        <b-col>
          <b-tabs content-class="mt-3">
            <b-tab title-link-class="pt-1 pb-1" title="Params" active>
              <b-table small hover class="dataTable" :items="request.parameters" :fields="parameterFields">
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
            </b-tab>
            <b-tab title-link-class="pt-1 pb-1" title="Body">
              <p class="label">Content Type</p>
              <p>{{ request.contentType }}</p>
              <div v-if="request.body">
                <pre :class="getLanguage(request.contentType)"><code :class="getLanguage(request.contentType)" v-html="pretty(request.body, request.contentType)"></code></pre>
              </div>
            </b-tab>
            <b-tab title-link-class="pt-1 pb-1" title="Workflows">
              <workflows :workflows="event.workflows"></workflows>
            </b-tab>
          </b-tabs>
        </b-col>
      </b-row>
    </b-card>
    <b-card>
      <b-row>
        <b-col>
          <b-tabs content-class="mt-3">
            <b-tab title-link-class="pt-1 pb-1" title="Body" active>
              <div v-if="response.body">
                <pre :class="getLanguage(response.headers['Content-Type'])"><code :class="getLanguage(response.headers['Content-Type'])" v-html="pretty(response.body, response.headers['Content-Type'])"></code></pre>
              </div>
            </b-tab>
            <b-tab title-link-class="pt-1 pb-1" title="Headers">
              <b-table small hover class="dataTable" :items="reponseHeaders">
              </b-table>
            </b-tab>
          </b-tabs>
        </b-col>
      </b-row>
    </b-card>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import Workflows from '@/components/Workflows'
import xmlFormatter from 'xml-formatter'
import MIMEType from 'whatwg-mimetype'
import Shortcut from '@/mixins/Shortcut'

export default {
  name: 'HttpRequest',
  components: {
    'workflows': Workflows
  },
  mixins: [Api, Filters, Shortcut],
  data () {
    return {
      request: null,
      response: null,
      event: null,
      parameterFields: [{key: 'show_details', label: '', thStyle: 'width: 1%'}, 'name', {key: 'value', tdClass: 'break'}, 'type', {key: 'openapi', label: 'OpenApi'}],
      detailsShown: [],
      parseError: null
    }
  },
  computed: {
    reponseHeaders: function () {
      let result = []
      for (let key in this.response.headers) {
        result.push({name: key, value: this.response.headers[key]})
      }
      return result
    }
  },
  watch: {
    $route () {
      this.getData()
    }
  },
  methods: {
    async getData () {
      let id = this.$route.params.id
      if (!id) {
        return
      }

      this.event = await this.getHttpRequest(id)
      this.request = this.event.data.request
      this.response = this.event.data.response
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
      if (!s) {
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
