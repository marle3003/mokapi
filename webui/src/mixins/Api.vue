<script>
export default {
  data () {
    let baseUrl = process.env.VUE_APP_ApiBaseUrl
    if (baseUrl === '' && window.location.pathname !== '/') {
      let host = window.location.origin
      let root = window.location.pathname
      baseUrl = (host + root).replace(/\/$/, '')
    }
    return {
      baseUrl: baseUrl
    }
  },
  methods: {
    async info () {
      let response = await this.$http.get(this.baseUrl + '/api/info')
      return response.data
    },
    async getService (serviceName) {
      let response = await this.$http.get(
        this.baseUrl + '/api/services/openapi/' + serviceName
      )
      return response.data
    },
    async getServices () {
      let response = await this.$http.get(this.baseUrl + '/api/services')
      return response.data
    },
    async getDashboard () {
      let response = await this.$http.get(this.baseUrl + '/api/dashboard')
      return response.data
    },
    async getHttpServices () {
      let response = await this.$http.get(this.baseUrl + '/api/services/http')
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
      return response.data.sort(compare)
    },
    async getKafkaServices () {
      let response = await this.$http.get(this.baseUrl + '/api/services/kafka')
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
      response.data.sort(compare)
      return response.data
    },
    async getAsyncApiService (serviceName) {
      let response = await this.$http.get(
        this.baseUrl + '/api/services/asyncapi/' + serviceName
      )
      return response.data
    },
    async getSmtpService (serviceName) {
      let response = await this.$http.get(
        this.baseUrl + '/api/services/smtp/' + serviceName
      )
      return response.data
    },
    async getHttpRequest (id) {
      let response = await this.$http.get(
        this.baseUrl + '/api/dashboard/http/requests/' + id
      )
      return response.data
    },
    async getSmtpMail (id) {
      let response = await this.$http.get(
        this.baseUrl + '/api/dashboard/smtp/mails/' + id
      )
      return response.data
    },
    async getKafkaTopic (kafka, topic) {
      let response = await this.$http.get(
        this.baseUrl + '/api/dashboard/kafka/' + kafka + '/topics/' + topic
      )
      return response.data
    },
    async getMetrics (...names) {
      let url = this.baseUrl + '/api/metrics'
      if (names.length > 0) {
        const p = names.join(',')
        url = url + '?names=' + p
      }
      let response = await this.$http.get(url)
      return response.data
    }
  }
}
</script>
