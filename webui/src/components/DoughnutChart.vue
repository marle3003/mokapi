<script>
import { Doughnut } from 'vue-chartjs'

export default {
  extends: Doughnut,
  props: {
    chartdata: {
      type: Object,
      default: null
    },
    options: {
      type: Object,
      default: function () {return {responsive: false, maintainAspectRatio: true}}
    }
  },
  mounted () {
    this.renderChart(this.chartdata, this.options)
  }
}
</script>