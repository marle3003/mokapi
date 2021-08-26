<script>
export default {
  data () {
    let baseUrl = process.env.VUE_APP_ApiBaseUrl
    if (baseUrl === '' && window.location.pathname !== '/') {
      let host = window.location.origin
      let root = window.location.pathname
      baseUrl = host + root
    }
    return {
      baseUrl: baseUrl
    }
  },
  methods: {
    async getService (serviceName) {
      let response = await this.$http.get(this.baseUrl + '/api/services/openapi/' + serviceName)
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
    async getAsyncApiService (serviceName) {
      let response = await this.$http.get(this.baseUrl + '/api/services/asyncapi/' + serviceName)
      return response.data
    },
    async getSmtpService (serviceName) {
      let response = await this.$http.get(this.baseUrl + '/api/services/smtp/' + serviceName)
      return response.data
    },
    async getHttpRequest (id) {
      let response = await this.$http.get(this.baseUrl + '/api/dashboard/http/requests/' + id)
      return response.data
    },
    async getSmtpMail (id) {
      let response = await this.$http.get(this.baseUrl + '/api/dashboard/smtp/mails/' + id)
      return response.data
    },
    async getKafkaTopic (kafka, topic) {
      let response = await this.$http.get(this.baseUrl + '/api/dashboard/kafka/' + kafka + '/topics/' + topic)
      return response.data
    }
  }
}
</script>
