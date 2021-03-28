<script>
export default {
  data () {
    return {
      baseUrl: process.env.VUE_APP_ApiBaseUrl
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
    }
  }
}
</script>
