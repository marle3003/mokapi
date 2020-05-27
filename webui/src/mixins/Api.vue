<script>
export default {
  data() {
      return {
          baseUrl: process.env.VUE_APP_ApiBaseUrl
      }
  },
  methods: {
      async getService(serviceName){
          let response = await this.$http.get(this.baseUrl + '/api/services/'+serviceName)
          return response.data
      },
      async getServices() {
          let response = await this.$http.get(this.baseUrl + '/api/services')
          return response.data
      }
  }
}
</script>