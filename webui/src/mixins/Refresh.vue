<script>
export default {
  data () {
    return {
      timer: null
    }
  },
  created () {
    this.init()
  },
  beforeDestroy () {
    clearInterval(this.timer)
  },
  watch: {
    $route () {
      this.init()
    }
  },
  methods: {
    init () {
      this.getData()
      clearInterval(this.timer)
      let refresh = this.$route.query.refresh
      if (refresh && refresh.length > 0) {
        let i = parseInt(refresh)
        if (!isNaN(i)) {
          this.timer = setInterval(this.getData, i * 1000)
        }
      }
    }
  }
}
</script>
